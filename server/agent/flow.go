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
	"iter"

	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/workflowagent"
	"google.golang.org/adk/v2/memory"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/workflow"

	"github.com/achetronic/magec/server/store"
)

// FlowBuildDeps bundles every dependency needed to construct the adk workflow
// nodes of a flow graph. The builder invokes BuildAgentInstance once per agent
// node, naming the instance after the node ID so the response filter can match
// event.Author against the node IDs the operator declared.
//
// FlowAgents holds previously-built flow agents indexed by flow ID so an agent
// node can reference another flow as its agent (flow-as-step composition). A
// referenced flow agent is wrapped under the node's own ID before it becomes a
// graph node.
//
// FlowStateToolset is the shared set_state/get_state toolset injected into
// every agent node so agents in the same flow share a scratchpad. It is
// stateless and safe to share.
type FlowBuildDeps struct {
	Ctx          context.Context
	AgentDefs    map[string]store.AgentDefinition
	FlowAgents   map[string]adkagent.Agent
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
}

// BuildFlowAgent translates a FlowDefinition graph into an adk workflow agent.
// It builds one workflow node per FlowNode, wires the operator's edges plus a
// synthetic Start -> Entry edge, and returns the graph behind a workflowagent
// whose name is the flow ID (so the flow is addressable by ID, like an agent).
func BuildFlowAgent(flow store.FlowDefinition, deps FlowBuildDeps) (adkagent.Agent, error) {
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

	return workflowagent.New(workflowagent.Config{
		Name:        flow.ID,
		Description: flow.Description,
		Edges:       edges,
	})
}

// buildNode constructs the adk workflow node for a single FlowNode. The node
// name is always the FlowNode ID, which keeps node ID == adk Node.Name() ==
// event.Author and lets the response filter work off the operator's IDs.
func buildNode(n store.FlowNode, deps FlowBuildDeps) (workflow.Node, error) {
	switch n.Type {
	case store.FlowNodeAgent:
		instance, err := buildAgentNodeAgent(n, deps)
		if err != nil {
			return nil, err
		}
		return workflow.NewAgentNode(instance, workflow.NodeConfig{})

	case store.FlowNodeRouter:
		return buildRouterNode(n)

	case store.FlowNodeJoin:
		return workflow.NewJoinNode(n.ID), nil

	default:
		return nil, fmt.Errorf("unknown node type %q", n.Type)
	}
}

// buildAgentNodeAgent resolves an agent node's underlying adk agent, named
// after the node ID. The AgentID either references an AgentDefinition (built
// fresh as a flow-scoped instance) or another flow (reused, renamed to the
// node ID so the graph node carries the operator's ID rather than the
// sub-flow's).
func buildAgentNodeAgent(n store.FlowNode, deps FlowBuildDeps) (adkagent.Agent, error) {
	if def, ok := deps.AgentDefs[n.AgentID]; ok {
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
			InstanceName:                n.ID,
			ExtraToolsets:               extraToolsets,
			IncludeFlowStateInstruction: deps.FlowStateToolset != nil,
		})
		if err != nil {
			return nil, err
		}
		return instance, nil
	}

	// Flow-as-node composition: the AgentID names another flow. Reuse its
	// already-built agent, renamed to this node's ID so the graph node is
	// addressable and matchable by the operator's ID.
	if subFlow, ok := deps.FlowAgents[n.AgentID]; ok {
		return renameAgent(n.ID, subFlow)
	}
	return nil, fmt.Errorf("agent %q referenced by node not found", n.AgentID)
}

// renameAgent returns an agent that delegates execution to the original but
// reports the given name. Used so a sub-flow reused as a node carries the
// node's ID instead of the sub-flow's own ID.
func renameAgent(name string, delegate adkagent.Agent) (adkagent.Agent, error) {
	return adkagent.New(adkagent.Config{
		Name:        name,
		Description: delegate.Description(),
		Run: func(ctx adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return delegate.Run(ctx)
		},
	})
}
