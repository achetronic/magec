// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package flowgraph

import (
	"fmt"
	"regexp"

	"github.com/achetronic/magec/server/agent/flowexit"
)

// idPattern is the safe-identifier shape required of node IDs and route
// labels. Node IDs become adk node names, event Authors and session-state key
// fragments, so they are kept to letters, digits, underscores and hyphens and
// must start with a letter or underscore.
var idPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

// Validate checks a flow graph at save time. It enforces the rules in section
// 6 of .agents/WORKFLOW_GRAPH_REDESIGN.md. Errors are returned to the caller
// (the admin API) which surfaces them to the operator; nothing here mutates
// the definition. Cycles are allowed on purpose: a loop is a back edge, capped
// at runtime, not rejected here.
func Validate(d *Definition) error {
	if err := validateNodes(d); err != nil {
		return err
	}
	index := d.nodeByID()
	if err := validateEntry(d, index); err != nil {
		return err
	}
	if err := validateEdges(d, index); err != nil {
		return err
	}
	if err := validateRouters(d, index); err != nil {
		return err
	}
	if err := validateJoins(d); err != nil {
		return err
	}
	if err := validateReachability(d, index); err != nil {
		return err
	}
	return nil
}

// validateNodes enforces unique, well-formed IDs and per-type field rules
// (rule 3 for agents, rule 6 for IDs, and the structural part of rule 4 for
// routers).
func validateNodes(d *Definition) error {
	if len(d.Nodes) == 0 {
		return fmt.Errorf("flow has no nodes")
	}
	seen := make(map[string]bool, len(d.Nodes))
	for i := range d.Nodes {
		n := &d.Nodes[i]
		if n.ID == "" {
			return fmt.Errorf("node has empty id")
		}
		if n.ID == Start {
			return fmt.Errorf("node id %q is reserved for the entry sentinel", Start)
		}
		if !idPattern.MatchString(n.ID) {
			return fmt.Errorf("node id %q must match [a-zA-Z_][a-zA-Z0-9_-]*", n.ID)
		}
		if seen[n.ID] {
			return fmt.Errorf("duplicate node id %q", n.ID)
		}
		seen[n.ID] = true

		switch n.Type {
		case NodeAgent:
			if n.AgentID == "" {
				return fmt.Errorf("agent node %q requires agentId", n.ID)
			}
		case NodeRouter:
			if len(n.Rules) == 0 {
				return fmt.Errorf("router node %q requires at least one rule", n.ID)
			}
			if n.DefaultRoute == "" {
				return fmt.Errorf("router node %q requires a defaultRoute", n.ID)
			}
			for j := range n.Rules {
				r := n.Rules[j]
				if r.Route == "" {
					return fmt.Errorf("router node %q rule %d has empty route", n.ID, j)
				}
				if !idPattern.MatchString(r.Route) {
					return fmt.Errorf("router node %q route %q must match [a-zA-Z_][a-zA-Z0-9_-]*", n.ID, r.Route)
				}
				// CEL guards are compiled here so malformed expressions are
				// caught at save time, not at flow build time. Compile also
				// rejects non-bool expressions.
				if _, err := flowexit.Compile(r.When); err != nil {
					return fmt.Errorf("router node %q rule %d: %w", n.ID, j, err)
				}
			}
			if !idPattern.MatchString(n.DefaultRoute) {
				return fmt.Errorf("router node %q defaultRoute %q must match [a-zA-Z_][a-zA-Z0-9_-]*", n.ID, n.DefaultRoute)
			}
		case NodeJoin:
			// No per-node fields; structural checks happen in validateJoins.
		default:
			return fmt.Errorf("node %q has unknown type %q", n.ID, n.Type)
		}
	}
	return nil
}

// validateEntry enforces rule 2: Entry must name an existing node.
func validateEntry(d *Definition, index map[string]*Node) error {
	if d.Entry == "" {
		return fmt.Errorf("flow has no entry node")
	}
	if _, ok := index[d.Entry]; !ok {
		return fmt.Errorf("entry node %q does not exist", d.Entry)
	}
	return nil
}

// validateEdges enforces rule 1: every endpoint references an existing node,
// with the Start sentinel allowed as a source.
func validateEdges(d *Definition, index map[string]*Node) error {
	for _, e := range d.Edges {
		if e.From != Start && e.From != "" {
			if _, ok := index[e.From]; !ok {
				return fmt.Errorf("edge from %q: source node does not exist", e.From)
			}
		}
		if e.To == "" {
			return fmt.Errorf("edge from %q has empty target", e.From)
		}
		if e.To == Start {
			return fmt.Errorf("edge targets the reserved entry sentinel %q", Start)
		}
		if _, ok := index[e.To]; !ok {
			return fmt.Errorf("edge to %q: target node does not exist", e.To)
		}
	}
	return nil
}

// validateRouters enforces rule 4's edge side: every label a router can emit
// (each rule Route plus DefaultRoute) has exactly one matching outgoing edge,
// and every outgoing edge of a router carries a Route that the router can
// actually emit. Non-router nodes must not carry routed outgoing edges.
func validateRouters(d *Definition, index map[string]*Node) error {
	for i := range d.Nodes {
		n := &d.Nodes[i]
		out := d.outgoing(n.ID)

		if n.Type != NodeRouter {
			for _, e := range out {
				if e.Route != "" {
					return fmt.Errorf("node %q is not a router but edge to %q carries route %q", n.ID, e.To, e.Route)
				}
			}
			continue
		}

		// Labels the router may emit.
		labels := make(map[string]bool, len(n.Rules)+1)
		for _, r := range n.Rules {
			labels[r.Route] = true
		}
		labels[n.DefaultRoute] = true

		// Each emittable label needs exactly one outgoing edge.
		edgeByRoute := make(map[string]int, len(out))
		for _, e := range out {
			if e.Route == "" {
				return fmt.Errorf("router node %q has an unconditional edge to %q; router edges must be routed", n.ID, e.To)
			}
			if !labels[e.Route] {
				return fmt.Errorf("router node %q edge to %q has route %q that no rule or default emits", n.ID, e.To, e.Route)
			}
			edgeByRoute[e.Route]++
			if edgeByRoute[e.Route] > 1 {
				return fmt.Errorf("router node %q has more than one edge for route %q", n.ID, e.Route)
			}
		}
		for label := range labels {
			if edgeByRoute[label] == 0 {
				return fmt.Errorf("router node %q can emit route %q but has no edge for it", n.ID, label)
			}
		}
	}
	return nil
}

// validateJoins enforces rule 5: no conditional (routed) edge may target a
// join node, because the barrier waits for every declared predecessor and a
// route-skipped predecessor would never fire, deadlocking the join.
func validateJoins(d *Definition) error {
	joins := make(map[string]bool)
	for i := range d.Nodes {
		if d.Nodes[i].Type == NodeJoin {
			joins[d.Nodes[i].ID] = true
		}
	}
	for _, e := range d.Edges {
		if joins[e.To] && e.Route != "" {
			return fmt.Errorf("join node %q is targeted by a routed edge from %q; joins require unconditional predecessors", e.To, e.From)
		}
	}
	return nil
}

// validateReachability enforces rule 7: every node is reachable from Entry (no
// orphans) and the graph has at least one terminal node (a node with no
// outgoing edges), so the flow can actually finish.
func validateReachability(d *Definition, index map[string]*Node) error {
	// BFS from Entry over the successor relation.
	successors := make(map[string][]string, len(d.Nodes))
	for _, e := range d.Edges {
		from := e.From
		if from == Start || from == "" {
			continue
		}
		successors[from] = append(successors[from], e.To)
	}
	// Edges from Start seed the reachable set alongside Entry itself.
	reachable := map[string]bool{d.Entry: true}
	queue := []string{d.Entry}
	for _, e := range d.Edges {
		if (e.From == Start || e.From == "") && !reachable[e.To] {
			reachable[e.To] = true
			queue = append(queue, e.To)
		}
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, next := range successors[cur] {
			if !reachable[next] {
				reachable[next] = true
				queue = append(queue, next)
			}
		}
	}
	for i := range d.Nodes {
		if !reachable[d.Nodes[i].ID] {
			return fmt.Errorf("node %q is not reachable from entry %q", d.Nodes[i].ID, d.Entry)
		}
	}

	hasTerminal := false
	for i := range d.Nodes {
		if len(d.outgoing(d.Nodes[i].ID)) == 0 {
			hasTerminal = true
			break
		}
	}
	if !hasTerminal {
		return fmt.Errorf("flow has no terminal node; every node has an outgoing edge so it can never finish")
	}
	return nil
}
