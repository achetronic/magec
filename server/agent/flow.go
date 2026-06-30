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
	"google.golang.org/adk/v2/agent/workflowagents/loopagent"
	"google.golang.org/adk/v2/agent/workflowagents/parallelagent"
	"google.golang.org/adk/v2/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/v2/memory"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"

	"github.com/achetronic/magec/server/agent/flowexit"
	"github.com/achetronic/magec/server/store"
)

// FlowBuildDeps bundles every dependency needed to construct the ADK agent
// instances inside a flow tree. The flow builder uses these to invoke
// BuildAgentInstance once per agent appearance, with extra toolsets that
// reflect the surrounding flow scope (e.g. exit_loop only inside a loop
// that opted into LLM-driven exit).
//
// The standalone agent catalogue still lives in agent.New: that copy is
// what direct callers (clients invoking an agent by its ID) reach. Flows
// build separate instances so a single AgentDefinition can appear in
// multiple flows with different surrounding capabilities, and so it can
// even appear several times inside the same flow tree without violating
// ADK's single-parent constraint.
//
// FlowAgents holds previously-built flow agents indexed by flow ID so a
// flow step can reference another flow as its agent (flow-as-step
// composition). Such referenced flow agents are reused as-is via wrapAgent
// — they are themselves graphs of fresh leaf instances.
//
// FlowStateToolset is the shared set_state/get_state toolset injected into
// every agent inside any flow. ExitLoopTool is the singleton exit_loop
// tool, conditionally injected only into agents that descend from a loop
// step whose ExitLoop flag is true. Both are stateless and safe to share.
type FlowBuildDeps struct {
	Ctx          context.Context
	AgentDefs    map[string]store.AgentDefinition
	FlowAgents   map[string]adkagent.Agent
	BackendMap   map[string]store.BackendDefinition
	MCPServerMap map[string]store.MCPServer
	// SkillSlugs maps skill ID -> on-disk slug. Forwarded verbatim
	// to BuildAgentInstance so per-flow agent instances build their
	// own skilltoolset filtered by the agent's whitelist.
	SkillSlugs map[string]string
	// SkillsDir is the absolute path to data/skills/. Empty disables
	// skill loading for every agent in the flow.
	SkillsDir        string
	MemorySvc        memory.Service
	BaseToolset      tool.Toolset
	FlowStateToolset tool.Toolset
	ExitLoopTool     tool.Tool
}

// BuildFlowAgent recursively translates a FlowDefinition into an ADK agent
// tree, building a fresh ADK instance for every agent appearance. The root
// step uses the flow ID as its ADK agent name so flows are addressable by
// ID, consistent with how individual agents are addressed.
func BuildFlowAgent(flow store.FlowDefinition, deps FlowBuildDeps) (adkagent.Agent, error) {
	return buildStep(flow.ID, &flow.Root, deps, "", false)
}

// buildStep recurses through the flow tree. insideLoopWithExitLoop is
// inherited from the closest enclosing loop step that opted into the
// exit_loop tool — once enabled, every agent in the subtree (including
// agents nested arbitrarily deep through other sequential/parallel/loop
// containers) gets the tool, since ADK's loopagent reacts to the
// Escalate event regardless of which descendant emitted it.
func buildStep(flowID string, step *store.FlowStep, deps FlowBuildDeps, path string, insideLoopWithExitLoop bool) (adkagent.Agent, error) {
	stepName := flowID
	if path != "" {
		stepName = fmt.Sprintf("%s_%s", flowID, path)
	}

	switch step.Type {
	case store.FlowStepAgent:
		if def, ok := deps.AgentDefs[step.AgentID]; ok {
			extraToolsets := []tool.Toolset{}
			if deps.FlowStateToolset != nil {
				extraToolsets = append(extraToolsets, deps.FlowStateToolset)
			}
			var extraTools []tool.Tool
			if insideLoopWithExitLoop && deps.ExitLoopTool != nil {
				extraTools = append(extraTools, deps.ExitLoopTool)
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
				InstanceName:                stepName,
				ExtraToolsets:               extraToolsets,
				ExtraTools:                  extraTools,
				IncludeFlowStateInstruction: deps.FlowStateToolset != nil,
				IncludeExitLoopInstruction:  insideLoopWithExitLoop && deps.ExitLoopTool != nil,
			})
			if err != nil {
				return nil, fmt.Errorf("flow %q step %q: %w", flowID, stepName, err)
			}
			return instance, nil
		}
		// Flow-as-step composition: the referenced ID names another flow.
		// Re-use its already-built ADK tree behind a wrapper so it can hang
		// off the parent flow without violating ADK's single-parent
		// constraint.
		if subFlow, ok := deps.FlowAgents[step.AgentID]; ok {
			return wrapAgent(stepName, subFlow)
		}
		return nil, fmt.Errorf("agent %q referenced by flow not found", step.AgentID)

	case store.FlowStepSequential:
		children, err := buildChildren(flowID, step.Steps, deps, path, insideLoopWithExitLoop)
		if err != nil {
			return nil, err
		}
		return sequentialagent.New(sequentialagent.Config{
			AgentConfig: adkagent.Config{
				Name:      stepName,
				SubAgents: children,
			},
		})

	case store.FlowStepParallel:
		children, err := buildChildren(flowID, step.Steps, deps, path, insideLoopWithExitLoop)
		if err != nil {
			return nil, err
		}
		return parallelagent.New(parallelagent.Config{
			AgentConfig: adkagent.Config{
				Name:      stepName,
				SubAgents: children,
			},
		})

	case store.FlowStepLoop:
		// A loop step turns on the exit_loop injection for its entire
		// subtree when ExitLoop is set. If we are already inside an
		// outer loop with exit_loop enabled, we keep it on regardless
		// (nested loops inherit the capability — exit_loop bubbles
		// Escalate up to the nearest loopagent that owns the iteration).
		childInside := insideLoopWithExitLoop || step.ExitLoop
		children, err := buildChildren(flowID, step.Steps, deps, path, childInside)
		if err != nil {
			return nil, err
		}
		// When the operator supplied an ExitWhen CEL expression, append a
		// synthetic evaluator agent as the last child of the loop. ADK
		// runs it after every iteration of the user-defined sub-agents;
		// when the expression evaluates to true it emits an event with
		// Actions.Escalate=true, which the surrounding loopagent already
		// honours by terminating the loop. The expression has been
		// validated at admin save time but is recompiled here because
		// programs are not serialisable.
		if step.ExitWhen != "" {
			prog, err := flowexit.Compile(step.ExitWhen)
			if err != nil {
				return nil, fmt.Errorf("flow %q step %q: invalid exitWhen: %w", flowID, stepName, err)
			}
			evalAgent, err := flowexit.NewExitWhenAgent(stepName+"_exitwhen", prog, step.ExitWhen)
			if err != nil {
				return nil, fmt.Errorf("flow %q step %q: %w", flowID, stepName, err)
			}
			children = append(children, evalAgent)
		}
		return loopagent.New(loopagent.Config{
			AgentConfig: adkagent.Config{
				Name:      stepName,
				SubAgents: children,
			},
			MaxIterations: step.MaxIterations,
		})

	default:
		return nil, fmt.Errorf("unknown flow step type %q", step.Type)
	}
}

// wrapAgent creates a uniquely-named agent that delegates execution to the
// original. Used only for flow-as-step composition (a leaf whose AgentID
// names another flow): each flow tree is built once at startup with its
// own fresh leaves, and the wrapper lets it appear under a different
// parent flow without violating ADK's single-parent constraint.
func wrapAgent(uniqueName string, delegate adkagent.Agent) (adkagent.Agent, error) {
	return adkagent.New(adkagent.Config{
		Name:        uniqueName,
		Description: delegate.Description(),
		Run: func(ctx adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return delegate.Run(ctx)
		},
	})
}

func buildChildren(flowID string, steps []store.FlowStep, deps FlowBuildDeps, parentPath string, insideLoopWithExitLoop bool) ([]adkagent.Agent, error) {
	children := make([]adkagent.Agent, 0, len(steps))
	for i := range steps {
		childPath := fmt.Sprintf("%d", i)
		if parentPath != "" {
			childPath = fmt.Sprintf("%s_%d", parentPath, i)
		}
		child, err := buildStep(flowID, &steps[i], deps, childPath, insideLoopWithExitLoop)
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return children, nil
}
