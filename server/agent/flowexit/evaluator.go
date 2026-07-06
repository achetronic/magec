// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package flowexit

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/cel-go/cel"
	"google.golang.org/adk/v2/session"

	toolsflowstate "github.com/achetronic/magec/server/agent/tools/flowstate"
)

// Evaluate runs a compiled CEL program against the node input, the supplied
// flow-state snapshot and the iteration count, returning the boolean result.
// It is the evaluation half of the router node: the builder compiles each
// rule's guard with Compile and the router calls Evaluate per rule, taking the
// first that returns true.
//
// expr is included only for log clarity when the program errors at evaluation
// time; it is not re-parsed. On a runtime error (missing key, type mismatch)
// or a non-bool result the function logs a warning and returns false, so a
// faulty guard never silently routes traffic: it simply does not match.
func Evaluate(prog cel.Program, expr string, input any, state map[string]any, iterations int) bool {
	out, _, err := prog.Eval(map[string]any{
		"input":      input,
		"state":      state,
		"iterations": iterations,
	})
	if err != nil {
		slog.Warn("flowexit: CEL evaluation failed; treating as false",
			"expression", expr, "error", err, "state_keys", keysOf(state))
		return false
	}
	b, ok := out.Value().(bool)
	if !ok {
		slog.Warn("flowexit: CEL did not return bool; treating as false",
			"expression", expr, "result", out.Value())
		return false
	}
	return b
}

// EvaluateValue runs a compiled transform-node program against the node input
// and the flow-state snapshot, returning the produced value. Unlike Evaluate
// (which is a loop guard and swallows errors as false), a transform failure is
// a real problem the operator must fix, so the error is returned to fail the
// node loudly rather than silently emitting a zero value.
func EvaluateValue(prog cel.Program, expr string, input any, state map[string]any) (any, error) {
	out, _, err := prog.Eval(map[string]any{
		"input": input,
		"state": state,
	})
	if err != nil {
		return nil, fmt.Errorf("evaluating %q: %w", expr, err)
	}
	return out.Value(), nil
}

// ExtractFlowState returns the subset of session state keys that live under
// the flow_state namespace, with the prefix stripped. CEL expressions
// reference these as `state.<key>`.
//
// Exported so the router node and tests can reuse the same prefix-stripping
// logic without re-importing flowstate.StateKeyPrefix.
func ExtractFlowState(s session.State) map[string]any {
	out := map[string]any{}
	for k, v := range s.All() {
		after, ok := strings.CutPrefix(k, toolsflowstate.StateKeyPrefix)
		if !ok {
			continue
		}
		// Keys beginning with "__" are internal bookkeeping (e.g. the router
		// iteration counter) and are not exposed to operator CEL expressions.
		if strings.HasPrefix(after, "__") {
			continue
		}
		out[after] = v
	}
	return out
}

func keysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
