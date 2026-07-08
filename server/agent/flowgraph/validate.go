// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

// Package flowgraph validates a flow graph (store.FlowDefinition) at save
// time, before it is handed to the builder. The graph model and its types
// live in package store; this package owns the rules a graph must satisfy
// and reuses flowexit to compile router CEL guards. See
// .agents/WORKFLOW_GRAPH_REDESIGN.md section 6.
package flowgraph

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/achetronic/magec/server/agent/flowexit"
	"github.com/achetronic/magec/server/store"
)

// idPattern is the safe-identifier shape required of node IDs and route
// labels. Node IDs become adk node names, event Authors and session-state key
// fragments, so they are kept to letters, digits, underscores and hyphens and
// must start with a letter or underscore.
var idPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

// stateKeyPattern is the shape required of a node's OutputKey. It is stricter
// than idPattern (no hyphen) because the key is referenced downstream as
// state.<key> in CEL, and a hyphen would parse as subtraction.
var stateKeyPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// nodeByID indexes a definition's nodes by ID for O(1) lookups.
func nodeByID(d *store.FlowDefinition) map[string]*store.FlowNode {
	index := make(map[string]*store.FlowNode, len(d.Nodes))
	for i := range d.Nodes {
		index[d.Nodes[i].ID] = &d.Nodes[i]
	}
	return index
}

// outgoing returns the edges leaving the given node ID.
func outgoing(d *store.FlowDefinition, id string) []store.FlowEdge {
	var edges []store.FlowEdge
	for _, e := range d.Edges {
		if e.From == id {
			edges = append(edges, e)
		}
	}
	return edges
}

// Validate checks a flow graph at save time. It enforces the rules in section
// 6 of .agents/WORKFLOW_GRAPH_REDESIGN.md. Errors are returned to the caller
// (the admin API) which surfaces them to the operator; nothing here mutates
// the definition. Cycles are allowed on purpose: a loop is a back edge, capped
// at runtime, not rejected here.
func Validate(d *store.FlowDefinition) error {
	if err := validateNodes(d); err != nil {
		return err
	}
	index := nodeByID(d)
	if err := validateEntry(d, index); err != nil {
		return err
	}
	if err := validateEdges(d, index); err != nil {
		return err
	}
	if err := validateRouters(d); err != nil {
		return err
	}
	if err := validateJoins(d); err != nil {
		return err
	}
	if err := validateReachability(d); err != nil {
		return err
	}
	return nil
}

// validateNodes enforces unique, well-formed IDs and per-type field rules
// (agents need an agentId, routers need rules with compilable CEL guards plus
// a default route, IDs match the safe pattern and are not the reserved start).
func validateNodes(d *store.FlowDefinition) error {
	if len(d.Nodes) == 0 {
		return fmt.Errorf("flow has no nodes")
	}
	seen := make(map[string]bool, len(d.Nodes))
	for i := range d.Nodes {
		n := &d.Nodes[i]
		if n.ID == "" {
			return fmt.Errorf("node has empty id")
		}
		if n.ID == store.FlowStart {
			return fmt.Errorf("node id %q is reserved for the entry sentinel", store.FlowStart)
		}
		if strings.HasPrefix(n.ID, "__") {
			return fmt.Errorf("node id %q is reserved: the double-underscore prefix marks internal nodes", n.ID)
		}
		if !idPattern.MatchString(n.ID) {
			return fmt.Errorf("node id %q must match [a-zA-Z_][a-zA-Z0-9_-]*", n.ID)
		}
		if seen[n.ID] {
			return fmt.Errorf("duplicate node id %q", n.ID)
		}
		seen[n.ID] = true

		switch n.Type {
		case store.FlowNodeAgent:
			if n.AgentID == "" {
				return fmt.Errorf("agent node %q requires agentId", n.ID)
			}
		case store.FlowNodeRouter:
			if len(n.Rules) == 0 {
				return fmt.Errorf("router node %q requires at least one rule", n.ID)
			}
			for j := range n.Rules {
				r := n.Rules[j]
				if r.Route == "" {
					return fmt.Errorf("router node %q rule %d has empty route", n.ID, j)
				}
				if !idPattern.MatchString(r.Route) {
					return fmt.Errorf("router node %q route %q must match [a-zA-Z_][a-zA-Z0-9_-]*", n.ID, r.Route)
				}
				if r.Route == store.RouterOtherwiseRoute {
					return fmt.Errorf("router node %q rule %d uses the reserved route %q", n.ID, j, store.RouterOtherwiseRoute)
				}
				// CEL guards are compiled here so malformed expressions are
				// caught at save time, not at flow build time. Compile also
				// rejects non-bool expressions.
				if _, err := flowexit.Compile(r.When); err != nil {
					return fmt.Errorf("router node %q rule %d: %w", n.ID, j, err)
				}
			}
		case store.FlowNodeJoin:
			// No per-node fields; structural checks happen in validateJoins.
		case store.FlowNodeParallel:
			if n.AgentID == "" {
				return fmt.Errorf("parallel node %q requires agentId (the agent run once per list item)", n.ID)
			}
			if n.MaxConcurrency < 0 {
				return fmt.Errorf("parallel node %q has negative maxConcurrency", n.ID)
			}
		case store.FlowNodeSubflow:
			if n.FlowID == "" {
				return fmt.Errorf("subflow node %q requires flowId", n.ID)
			}
		case store.FlowNodeExpression:
			if n.Expression == "" {
				return fmt.Errorf("expression node %q requires an expression", n.ID)
			}
			if _, err := flowexit.CompileValue(n.Expression); err != nil {
				return fmt.Errorf("expression node %q: %w", n.ID, err)
			}
			if n.OutputKey != "" && !stateKeyPattern.MatchString(n.OutputKey) {
				return fmt.Errorf("expression node %q outputKey %q must match [a-zA-Z_][a-zA-Z0-9_]*", n.ID, n.OutputKey)
			}
		case store.FlowNodeTemplate:
			if n.Template == "" {
				return fmt.Errorf("template node %q requires a template", n.ID)
			}
			if n.OutputKey != "" && !stateKeyPattern.MatchString(n.OutputKey) {
				return fmt.Errorf("template node %q outputKey %q must match [a-zA-Z_][a-zA-Z0-9_]*", n.ID, n.OutputKey)
			}
		case store.FlowNodeCode:
			if n.Script == "" {
				return fmt.Errorf("code node %q requires a script", n.ID)
			}
			if n.OutputKey != "" && !stateKeyPattern.MatchString(n.OutputKey) {
				return fmt.Errorf("code node %q outputKey %q must match [a-zA-Z_][a-zA-Z0-9_]*", n.ID, n.OutputKey)
			}
			if n.TimeoutMs < 0 {
				return fmt.Errorf("code node %q has negative timeoutMs", n.ID)
			}
			if n.MaxOutputBytes < 0 {
				return fmt.Errorf("code node %q has negative maxOutputBytes", n.ID)
			}
		default:
			return fmt.Errorf("node %q has unknown type %q", n.ID, n.Type)
		}
	}
	return nil
}

// validateEntry enforces that Entry names an existing node.
func validateEntry(d *store.FlowDefinition, index map[string]*store.FlowNode) error {
	if d.Entry == "" {
		return fmt.Errorf("flow has no entry node")
	}
	if _, ok := index[d.Entry]; !ok {
		return fmt.Errorf("entry node %q does not exist", d.Entry)
	}
	return nil
}

// validateEdges enforces that every endpoint references an existing node. The
// Start sentinel is reserved: operators never wire it themselves (the builder
// synthesizes Start -> Entry), so an edge that references FlowStart is an error.
func validateEdges(d *store.FlowDefinition, index map[string]*store.FlowNode) error {
	for _, e := range d.Edges {
		if e.From == "" {
			return fmt.Errorf("edge to %q has empty source", e.To)
		}
		if e.From == store.FlowStart || e.To == store.FlowStart {
			return fmt.Errorf("edge references the reserved entry sentinel %q; set Entry instead of wiring Start", store.FlowStart)
		}
		if _, ok := index[e.From]; !ok {
			return fmt.Errorf("edge from %q: source node does not exist", e.From)
		}
		if e.To == "" {
			return fmt.Errorf("edge from %q has empty target", e.From)
		}
		if _, ok := index[e.To]; !ok {
			return fmt.Errorf("edge to %q: target node does not exist", e.To)
		}
	}
	return nil
}

// validateRouters enforces that every label a router can emit (each rule Route
// plus the fixed otherwise route) has exactly one matching outgoing edge, and that every
// outgoing edge of a router carries a Route the router can actually emit.
// Non-router nodes must not carry routed outgoing edges.
func validateRouters(d *store.FlowDefinition) error {
	for i := range d.Nodes {
		n := &d.Nodes[i]
		out := outgoing(d, n.ID)

		if n.Type != store.FlowNodeRouter {
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
		labels[store.RouterOtherwiseRoute] = true

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

// validateJoins enforces that no conditional (routed) edge targets a join
// node, because the barrier waits for every declared predecessor and a
// route-skipped predecessor would never fire, deadlocking the join.
func validateJoins(d *store.FlowDefinition) error {
	joins := make(map[string]bool)
	for i := range d.Nodes {
		if d.Nodes[i].Type == store.FlowNodeJoin {
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

// validateReachability enforces that every node is reachable from Entry (no
// orphans) and the graph has at least one terminal node (no outgoing edges),
// so the flow can actually finish. The builder wires Start -> Entry, so
// reachability is measured from Entry.
func validateReachability(d *store.FlowDefinition) error {
	successors := make(map[string][]string, len(d.Nodes))
	for _, e := range d.Edges {
		successors[e.From] = append(successors[e.From], e.To)
	}
	reachable := map[string]bool{d.Entry: true}
	queue := []string{d.Entry}
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
		if len(outgoing(d, d.Nodes[i].ID)) == 0 {
			hasTerminal = true
			break
		}
	}
	if !hasTerminal {
		return fmt.Errorf("flow has no terminal node; every node has an outgoing edge so it can never finish")
	}
	return nil
}
