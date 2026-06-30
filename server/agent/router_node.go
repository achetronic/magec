// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package agent

import (
	"fmt"

	"github.com/google/cel-go/cel"
	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/workflow"

	"github.com/achetronic/magec/server/agent/flowexit"
	toolsflowstate "github.com/achetronic/magec/server/agent/tools/flowstate"
	"github.com/achetronic/magec/server/store"
)

// maxLoopIterations is the hard runaway guard for a router node. A flow loop is
// a back edge with no built-in cap, and adk-go does not bound conditional
// cycles, so a router that never emits its exit route would spin forever. When
// a single router is activated more than this many times in one run it fails
// the workflow rather than looping silently. Operators express their own,
// lower caps in CEL via the `iterations` variable; this is only the last line
// of defence.
const maxLoopIterations = 1000

// iterationKeyPrefix namespaces the per-router activation counter inside the
// shared flow state. The counter lives under the "flow:" namespace so it
// persists across loop-back activations (the adk scheduler resets a node's
// lifecycle on every loop-back, but session state survives). The "__" marks it
// internal so ExtractFlowState hides it from operator CEL expressions.
const iterationKeyPrefix = toolsflowstate.StateKeyPrefix + "__iter__"

// compiledRule pairs a router rule's source expression with its compiled CEL
// program, prepared once at build time so the hot path only evaluates.
type compiledRule struct {
	prog  cel.Program
	expr  string
	route string
}

// buildRouterNode turns a router FlowNode into an adk workflow node. The node
// reads the shared flow state and its own activation count, evaluates the
// rules in order, and emits the route label of the first rule whose CEL guard
// is true (or the default route when none match). The emitted Event.Routes is
// what the outgoing StringRoute edges match against; Event.Output passes the
// upstream input through unchanged so the chosen successor still receives it.
func buildRouterNode(node store.FlowNode) (workflow.Node, error) {
	rules := make([]compiledRule, 0, len(node.Rules))
	for i, r := range node.Rules {
		prog, err := flowexit.Compile(r.When)
		if err != nil {
			// Should not happen: the admin API validates guards at save time.
			// Recompiling here is necessary because programs are not
			// serialisable, and failing loudly beats routing on a broken guard.
			return nil, fmt.Errorf("router %q rule %d: %w", node.ID, i, err)
		}
		rules = append(rules, compiledRule{prog: prog, expr: r.When, route: r.Route})
	}

	defaultRoute := node.DefaultRoute
	counterKey := iterationKeyPrefix + node.ID

	fn := func(ctx adkagent.Context, input any, emit func(*session.Event) error) (any, error) {
		st := ctx.State()
		iterations := readInt(st, counterKey) + 1
		if iterations > maxLoopIterations {
			return nil, fmt.Errorf("router %q exceeded the maximum of %d loop iterations; check its exit condition", node.ID, maxLoopIterations)
		}

		flowState := flowexit.ExtractFlowState(st)
		route := defaultRoute
		for _, r := range rules {
			if flowexit.Evaluate(r.prog, r.expr, flowState, iterations) {
				route = r.route
				break
			}
		}

		ev := session.NewEvent(ctx, ctx.InvocationID())
		ev.Author = node.ID
		ev.Routes = []string{route}
		ev.Output = input
		// Persist the incremented counter so the next loop-back activation of
		// this router observes it. NewEvent allocates StateDelta for us.
		ev.Actions.StateDelta[counterKey] = iterations
		if err := emit(ev); err != nil {
			return nil, err
		}
		// Returning nil suppresses the default terminal event: the routing
		// event we emitted is this node's sole output.
		return nil, nil
	}

	return workflow.NewEmittingFunctionNode[any, any](node.ID, fn, workflow.NodeConfig{}), nil
}

// readInt reads an integer counter from session state, tolerating the numeric
// types a value may take after a JSON round-trip through session persistence
// (int, int64 or float64). A missing or unparsable key reads as zero.
func readInt(st session.State, key string) int {
	v, err := st.Get(key)
	if err != nil || v == nil {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}
