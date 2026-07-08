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
	"encoding/json"
	"fmt"
	"time"

	"github.com/1set/starlet"
	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"

	"github.com/achetronic/magec/server/agent/flowexit"
	toolsflowstate "github.com/achetronic/magec/server/agent/tools/flowstate"
	"github.com/achetronic/magec/server/store"
)

// maxCodeNodeSteps is the per-run Starlark step budget applied when a
// wall-clock timeout is also active. It prevents tight infinite loops from
// burning the full timeout window before the context deadline fires.
const maxCodeNodeSteps = uint64(1_000_000_000)

// effectiveLimit resolves the actual ceiling for a per-node numeric limit
// (timeout or output cap) given the node's own value and the admin ceiling.
//
// Rules:
//   - ceiling == 0 (no global cap): effective = nodeVal (0 = unlimited).
//   - ceiling > 0, nodeVal == 0 (no per-node override): effective = ceiling.
//   - both > 0: effective = min(nodeVal, ceiling).
func effectiveLimit(nodeVal, ceiling int) int {
	if ceiling == 0 {
		return nodeVal
	}
	if nodeVal == 0 {
		return ceiling
	}
	if nodeVal < ceiling {
		return nodeVal
	}
	return ceiling
}

// buildCodeNode constructs the workflow node that executes a user-supplied
// Starlark script per invocation. The script receives `input` (the upstream
// node output) and `state` (the shared flow state map) as top-level
// variables, and is expected to assign `output`. If the script does not
// assign `output`, the node emits nil without error.
//
// A fresh Machine is created per execution so concurrent activations (e.g.
// inside a parallel worker) are fully isolated. The shared prebuilt loader
// list from deps.StarletLoaders is safe to reference from multiple goroutines.
func buildCodeNode(node store.FlowNode, deps FlowBuildDeps) (workflow.Node, error) {
	script := node.Script
	outputKey := node.OutputKey
	loaders := deps.StarletLoaders
	flowsSettings := deps.FlowsSettings
	secretMap := flowSecretMap(deps)

	fn := func(ctx adkagent.Context, input any, emit func(*session.Event) error) (any, error) {
		effTimeout := effectiveLimit(node.TimeoutMs, flowsSettings.ExecutionTimeoutMs)
		effMaxOutput := effectiveLimit(node.MaxOutputBytes, flowsSettings.MaxOutputBytes)

		flowState := flowexit.ExtractFlowState(ctx.State())

		m := starlet.NewWithLoaders(nil, loaders, nil)
		m.SetPrintFunc(starlet.NoopPrintFunc)
		m.SetScriptContent([]byte(script))

		secret := map[string]any{}
		for k, v := range secretMap {
			secret[k] = v
		}
		extras := starlet.StringAnyMap{"input": input, "state": flowState, "secret": secret}

		var (
			res starlet.StringAnyMap
			err error
		)
		if effTimeout > 0 {
			m.SetMaxExecutionSteps(maxCodeNodeSteps)
			runCtx, cancel := context.WithTimeout(ctx, time.Duration(effTimeout)*time.Millisecond)
			defer cancel()
			res, err = m.RunWithContext(runCtx, extras)
		} else {
			m.SetMaxExecutionSteps(0)
			res, err = m.RunWithContext(ctx, extras)
		}
		if err != nil {
			return nil, fmt.Errorf("code node %q: %w", node.ID, err)
		}

		out := res["output"]

		if effMaxOutput > 0 && out != nil {
			b, jsonErr := json.Marshal(out)
			if jsonErr == nil && len(b) > effMaxOutput {
				return nil, fmt.Errorf("code node %q: output size %d bytes exceeds cap of %d bytes", node.ID, len(b), effMaxOutput)
			}
		}

		ev := session.NewEvent(ctx, ctx.InvocationID())
		ev.Author = node.ID
		ev.Output = out
		if outputKey != "" {
			ev.Actions.StateDelta[toolsflowstate.StateKeyPrefix+outputKey] = out
		}
		if err := emit(ev); err != nil {
			return nil, err
		}
		return nil, nil
	}

	return workflow.NewEmittingFunctionNode[any, any](node.ID, fn, workflow.NodeConfig{}), nil
}
