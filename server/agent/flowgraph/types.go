// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Package flowgraph holds the graph model that replaces the recursive
// store.FlowStep tree (see .agents/WORKFLOW_GRAPH_REDESIGN.md). A flow is a
// set of named nodes plus a set of directed edges, which maps one-to-one onto
// the adk-go v2 workflow engine ([]workflow.Edge wired into workflowagent.New).
//
// The model and its validation are prototyped here, with tests, so the design
// is exercised end to end before the adk v2 module bump. At cutover these
// types are lifted into server/store as the new FlowDefinition shape and this
// package is removed; the type names are kept deliberately store-friendly.
package flowgraph

// Node kinds. A node's Type selects which fields are meaningful.
const (
	// NodeAgent wraps an AgentDefinition (or a sub-flow) and runs it.
	NodeAgent = "agent"
	// NodeRouter evaluates an ordered list of CEL rules against the shared
	// flow state and emits a single route label, which the source node's
	// outgoing edges match. This replaces the old flowexit Escalate hack.
	NodeRouter = "router"
	// NodeJoin is a fan-in barrier: it fires once after every declared
	// predecessor has completed. Routing into a join node is forbidden.
	NodeJoin = "join"
)

// Start is the reserved identifier for the graph entry sentinel. An edge whose
// From equals Start (or is empty) is wired to workflow.Start by the builder.
// It is reserved so an operator cannot name a real node "START".
const Start = "START"

// Rule is one ordered branch of a router node. When the CEL guard evaluates to
// true against the flow state, the router emits Route as its label and stops
// evaluating later rules. When is a CEL expression over a single `state`
// variable (map<string, dyn>), the same environment used by flowexit.
type Rule struct {
	When  string `json:"when" yaml:"when"`
	Route string `json:"route" yaml:"route"`
}

// Node is one vertex of the flow graph. ID is unique within the flow and
// becomes the adk workflow Node.Name(), so it must be a safe identifier: it
// also appears as the event Author used by the response filter and as a
// fragment of session-state keys.
type Node struct {
	ID   string `json:"id" yaml:"id"`
	Type string `json:"type" yaml:"type"`

	// AgentID is the referenced AgentDefinition or FlowDefinition ID.
	// Required when Type is NodeAgent, ignored otherwise.
	AgentID string `json:"agentId,omitempty" yaml:"agentId,omitempty"`

	// ResponseAgent marks an agent node whose output is included in the final
	// response when the flow is invoked via webhook or cron. Only meaningful
	// when Type is NodeAgent.
	ResponseAgent bool `json:"responseAgent,omitempty" yaml:"responseAgent,omitempty"`

	// Rules and DefaultRoute drive a router node. Rules are evaluated in order;
	// DefaultRoute is emitted when no rule matches. Only meaningful when Type
	// is NodeRouter.
	Rules        []Rule `json:"rules,omitempty" yaml:"rules,omitempty"`
	DefaultRoute string `json:"defaultRoute,omitempty" yaml:"defaultRoute,omitempty"`
}

// Edge is a directed connection between two nodes. Route is only meaningful
// when From is a router node: it names the label the router must emit for this
// edge to be taken. An empty Route is an unconditional edge.
type Edge struct {
	From  string `json:"from" yaml:"from"`
	To    string `json:"to" yaml:"to"`
	Route string `json:"route,omitempty" yaml:"route,omitempty"`
}

// Definition is a multi-agent workflow stored as a directed graph. It is the
// graph-model replacement for store.FlowDefinition. Entry names the node wired
// to the Start sentinel; it is recorded explicitly rather than inferred so a
// multi-root graph keeps the operator's intent.
type Definition struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Entry       string `json:"entry" yaml:"entry"`
	Nodes       []Node `json:"nodes" yaml:"nodes"`
	Edges       []Edge `json:"edges" yaml:"edges"`
}

// nodeByID indexes the definition's nodes by ID for O(1) lookups during
// validation and building. Returns the index and whether the lookup is total
// (no duplicate IDs); duplicate detection is left to Validate.
func (d *Definition) nodeByID() map[string]*Node {
	index := make(map[string]*Node, len(d.Nodes))
	for i := range d.Nodes {
		index[d.Nodes[i].ID] = &d.Nodes[i]
	}
	return index
}

// outgoing returns the edges leaving the given node ID.
func (d *Definition) outgoing(id string) []Edge {
	var edges []Edge
	for _, e := range d.Edges {
		if e.From == id {
			edges = append(edges, e)
		}
	}
	return edges
}
