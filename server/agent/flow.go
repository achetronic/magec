// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package agent

import (
	"context"
	"fmt"

	"github.com/1set/starlet"
	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/workflowagent"
	"google.golang.org/adk/v2/memory"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/workflow"

	"github.com/achetronic/magec/server/store"
)

// FlowBuildDeps bundles every dependency needed to construct the adk workflow
// nodes of a flow graph. The builder invokes BuildAgentInstance once per agent
// node, naming the instance after the node ID so the response filter can match
// event.Author against the node IDs the operator declared.
//
// FlowDefs holds every flow definition by ID so a subflow node can embed
// another flow as a nested workflow (adk's WorkflowNode), built from that
// flow's own edges.
//
// FlowStateToolset is the shared set_state/get_state toolset injected into
// every agent node so agents in the same flow share a scratchpad. It is
// stateless and safe to share.
type FlowBuildDeps struct {
	Ctx          context.Context
	AgentDefs    map[string]store.AgentDefinition
	FlowDefs     map[string]store.FlowDefinition
	BackendMap   map[string]store.BackendDefinition
	MCPServerMap map[string]store.MCPServer
	// SkillSlugs maps skill ID -> on-disk slug. Forwarded verbatim to
	// BuildAgentInstance so per-flow agent instances build their own
	// skilltoolset filtered by the agent's whitelist.
	SkillSlugs       map[string]string
	SkillsDir        string
	MemorySvc        memory.Service
	BaseToolset      tool.Toolset
	FlowStateToolset tool.Toolset
	// StarletLoaders is the prebuilt list of enabled Starlark module loaders.
	// Safe to share across concurrent code-node executions; each run builds
	// its own fresh Machine.
	StarletLoaders starlet.ModuleLoaderList
	// FlowsSettings carries the admin ceilings (timeout, output cap) for code
	// nodes. Computed once in agent.New and forwarded through FlowBuildDeps.
	FlowsSettings store.FlowsSettings
}

// BuildFlowAgent translates a FlowDefinition graph into an adk workflow agent
// whose name is the flow ID (so the flow is addressable by ID, like an agent).
func BuildFlowAgent(flow store.FlowDefinition, deps FlowBuildDeps) (adkagent.Agent, error) {
	edges, err := buildEdges(flow, deps)
	if err != nil {
		return nil, err
	}
	return workflowagent.New(workflowagent.Config{
		Name:        flow.ID,
		Description: flow.Description,
		Edges:       edges,
	})
}

// buildEdges builds one workflow node per FlowNode and wires the operator's
// edges plus a synthetic Start -> Entry edge. It is reused both for a top-level
// flow agent and for a subflow embedded via WorkflowNode.
func buildEdges(flow store.FlowDefinition, deps FlowBuildDeps) ([]workflow.Edge, error) {
	nodeMap := make(map[string]workflow.Node, len(flow.Nodes))
	for i := range flow.Nodes {
		n := flow.Nodes[i]
		node, err := buildNode(n, deps)
		if err != nil {
			return nil, fmt.Errorf("flow %q node %q: %w", flow.ID, n.ID, err)
		}
		nodeMap[n.ID] = node
	}

	entryNode, ok := nodeMap[flow.Entry]
	if !ok {
		return nil, fmt.Errorf("flow %q: entry node %q not found", flow.ID, flow.Entry)
	}

	// Entry is the single source of truth for where the graph starts: wire the
	// Start sentinel to it. Operators never manage the Start node themselves.
	edges := []workflow.Edge{{From: workflow.Start, To: entryNode}}

	for _, e := range flow.Edges {
		from, ok := nodeMap[e.From]
		if !ok {
			return nil, fmt.Errorf("flow %q: edge from unknown node %q", flow.ID, e.From)
		}
		to, ok := nodeMap[e.To]
		if !ok {
			return nil, fmt.Errorf("flow %q: edge to unknown node %q", flow.ID, e.To)
		}
		edge := workflow.Edge{From: from, To: to}
		if e.Route != "" {
			edge.Route = workflow.StringRoute(e.Route)
		}
		edges = append(edges, edge)
	}
	return edges, nil
}

// buildNode constructs the adk workflow node for a single FlowNode. The node
// name is always the FlowNode ID, which keeps node ID == adk Node.Name() ==
// event.Author and lets the response filter work off the operator's IDs.
func buildNode(n store.FlowNode, deps FlowBuildDeps) (workflow.Node, error) {
	switch n.Type {
	case store.FlowNodeAgent:
		instance, err := buildFlowScopedAgent(n.ID, n.AgentID, deps)
		if err != nil {
			return nil, err
		}
		return workflow.NewAgentNode(instance, workflow.NodeConfig{})

	case store.FlowNodeRouter:
		return buildRouterNode(n)

	case store.FlowNodeJoin:
		return workflow.NewJoinNode(n.ID), nil

	case store.FlowNodeParallel:
		// Wrap the agent in an AgentNode, then run it once per list item.
		instance, err := buildFlowScopedAgent(n.ID, n.AgentID, deps)
		if err != nil {
			return nil, err
		}
		inner, err := workflow.NewAgentNode(instance, workflow.NodeConfig{})
		if err != nil {
			return nil, err
		}
		return workflow.NewParallelWorker(n.ID, inner, n.MaxConcurrency, workflow.NodeConfig{})

	case store.FlowNodeSubflow:
		sub, ok := deps.FlowDefs[n.FlowID]
		if !ok {
			return nil, fmt.Errorf("subflow node %q references unknown flow %q", n.ID, n.FlowID)
		}
		subEdges, err := buildEdges(sub, deps)
		if err != nil {
			return nil, fmt.Errorf("subflow node %q: %w", n.ID, err)
		}
		return workflow.NewWorkflowNode(n.ID, subEdges)

	case store.FlowNodeExpression:
		return buildExpressionNode(n)

	case store.FlowNodeTemplate:
		return buildTemplateNode(n)

	case store.FlowNodeCode:
		return buildCodeNode(n, deps)

	default:
		return nil, fmt.Errorf("unknown node type %q", n.Type)
	}
}

// buildFlowScopedAgent builds a fresh adk agent instance for an AgentDefinition,
// named after the graph node (so event.Author == node ID), with the shared
// flow-state toolset injected. instanceName is the node ID; agentID is the
// referenced AgentDefinition.
func buildFlowScopedAgent(instanceName, agentID string, deps FlowBuildDeps) (adkagent.Agent, error) {
	def, ok := deps.AgentDefs[agentID]
	if !ok {
		return nil, fmt.Errorf("agent %q referenced by node not found", agentID)
	}
	var extraToolsets []tool.Toolset
	if deps.FlowStateToolset != nil {
		extraToolsets = append(extraToolsets, deps.FlowStateToolset)
	}
	instance, _, err := BuildAgentInstance(BuildAgentInstanceParams{
		Ctx:                         deps.Ctx,
		AgentDef:                    def,
		BackendMap:                  deps.BackendMap,
		MCPServerMap:                deps.MCPServerMap,
		SkillSlugs:                  deps.SkillSlugs,
		SkillsDir:                   deps.SkillsDir,
		MemorySvc:                   deps.MemorySvc,
		BaseToolset:                 deps.BaseToolset,
		InstanceName:                instanceName,
		ExtraToolsets:               extraToolsets,
		IncludeFlowStateInstruction: deps.FlowStateToolset != nil,
	})
	if err != nil {
		return nil, err
	}
	return instance, nil
}
