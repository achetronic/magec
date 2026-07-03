// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package flowgraph

import (
	"strings"
	"testing"

	"github.com/achetronic/magec/server/store"
)

// linearFlow is a minimal valid graph: START -> a -> b, b terminal.
func linearFlow() *store.FlowDefinition {
	return &store.FlowDefinition{
		ID:    "f1",
		Name:  "linear",
		Entry: "a",
		Nodes: []store.FlowNode{
			{ID: "a", Type: store.FlowNodeAgent, AgentID: "agent-a"},
			{ID: "b", Type: store.FlowNodeAgent, AgentID: "agent-b", ResponseAgent: true},
		},
		Edges: []store.FlowEdge{
			{From: "a", To: "b"},
		},
	}
}

// routerFlow exercises a router with two rules plus a default, each wired to a
// distinct terminal agent. This is the canonical CEL-as-router shape.
func routerFlow() *store.FlowDefinition {
	return &store.FlowDefinition{
		ID:    "f2",
		Name:  "router",
		Entry: "classify",
		Nodes: []store.FlowNode{
			{ID: "classify", Type: store.FlowNodeRouter, DefaultRoute: "revise", Rules: []store.FlowRule{
				{When: "state.score >= 0.8", Route: "accept"},
				{When: "state.score < 0.3", Route: "reject"},
			}},
			{ID: "publish", Type: store.FlowNodeAgent, AgentID: "ag-pub"},
			{ID: "discard", Type: store.FlowNodeAgent, AgentID: "ag-dis"},
			{ID: "rewrite", Type: store.FlowNodeAgent, AgentID: "ag-rew"},
		},
		Edges: []store.FlowEdge{
			{From: "classify", To: "publish", Route: "accept"},
			{From: "classify", To: "discard", Route: "reject"},
			{From: "classify", To: "rewrite", Route: "revise"},
		},
	}
}

// fanFlow exercises fan-out from one agent to two workers and fan-in through a
// join barrier into a terminal agent.
func fanFlow() *store.FlowDefinition {
	return &store.FlowDefinition{
		ID:    "f3",
		Name:  "fan",
		Entry: "split",
		Nodes: []store.FlowNode{
			{ID: "split", Type: store.FlowNodeAgent, AgentID: "ag-split"},
			{ID: "w1", Type: store.FlowNodeAgent, AgentID: "ag-w1"},
			{ID: "w2", Type: store.FlowNodeAgent, AgentID: "ag-w2"},
			{ID: "merge", Type: store.FlowNodeJoin},
			{ID: "done", Type: store.FlowNodeAgent, AgentID: "ag-done", ResponseAgent: true},
		},
		Edges: []store.FlowEdge{
			{From: "split", To: "w1"},
			{From: "split", To: "w2"},
			{From: "w1", To: "merge"},
			{From: "w2", To: "merge"},
			{From: "merge", To: "done"},
		},
	}
}

// loopFlow exercises a back edge: a router decides to loop back to the worker
// or exit. Cycles must be accepted.
func loopFlow() *store.FlowDefinition {
	return &store.FlowDefinition{
		ID:    "f4",
		Name:  "loop",
		Entry: "work",
		Nodes: []store.FlowNode{
			{ID: "work", Type: store.FlowNodeAgent, AgentID: "ag-work"},
			{ID: "again", Type: store.FlowNodeRouter, DefaultRoute: "continue", Rules: []store.FlowRule{
				{When: "state.done == true || iterations >= 5", Route: "exit"},
			}},
			{ID: "out", Type: store.FlowNodeAgent, AgentID: "ag-out", ResponseAgent: true},
		},
		Edges: []store.FlowEdge{
			{From: "work", To: "again"},
			{From: "again", To: "work", Route: "continue"},
			{From: "again", To: "out", Route: "exit"},
		},
	}
}

func TestValidate_AcceptsValidGraphs(t *testing.T) {
	cases := map[string]*store.FlowDefinition{
		"linear":    linearFlow(),
		"router":    routerFlow(),
		"fan_join":  fanFlow(),
		"loop_back": loopFlow(),
	}
	for name, def := range cases {
		t.Run(name, func(t *testing.T) {
			if err := Validate(def); err != nil {
				t.Fatalf("expected valid graph, got error: %v", err)
			}
		})
	}
}

func TestValidate_RejectsInvalidGraphs(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*store.FlowDefinition)
		wantSub string
	}{
		{
			name:    "no nodes",
			mutate:  func(d *store.FlowDefinition) { d.Nodes = nil },
			wantSub: "no nodes",
		},
		{
			name: "duplicate node id",
			mutate: func(d *store.FlowDefinition) {
				d.Nodes = append(d.Nodes, store.FlowNode{ID: "a", Type: store.FlowNodeAgent, AgentID: "x"})
			},
			wantSub: "duplicate node id",
		},
		{
			name:    "reserved start id",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].ID = store.FlowStart },
			wantSub: "reserved",
		},
		{
			name:    "reserved internal prefix",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].ID = "__meta__"; d.Entry = "__meta__" },
			wantSub: "reserved",
		},
		{
			name:    "bad id pattern",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].ID = "has space"; d.Entry = "has space" },
			wantSub: "must match",
		},
		{
			name:    "agent without agentId",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].AgentID = "" },
			wantSub: "requires agentId",
		},
		{
			name:    "unknown node type",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].Type = "frobnicate" },
			wantSub: "unknown type",
		},
		{
			name:    "entry does not exist",
			mutate:  func(d *store.FlowDefinition) { d.Entry = "ghost" },
			wantSub: "entry node \"ghost\" does not exist",
		},
		{
			name:    "edge to missing node",
			mutate:  func(d *store.FlowDefinition) { d.Edges = append(d.Edges, store.FlowEdge{From: "a", To: "ghost"}) },
			wantSub: "target node does not exist",
		},
		{
			name:    "edge from missing node",
			mutate:  func(d *store.FlowDefinition) { d.Edges = append(d.Edges, store.FlowEdge{From: "ghost", To: "b"}) },
			wantSub: "source node does not exist",
		},
		{
			name: "orphan node unreachable",
			mutate: func(d *store.FlowDefinition) {
				d.Nodes = append(d.Nodes, store.FlowNode{ID: "island", Type: store.FlowNodeAgent, AgentID: "x"})
			},
			wantSub: "not reachable",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def := linearFlow()
			tc.mutate(def)
			err := Validate(def)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantSub)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantSub)
			}
		})
	}
}

func TestValidate_RouterRules(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*store.FlowDefinition)
		wantSub string
	}{
		{
			name:    "router without rules",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].Rules = nil },
			wantSub: "at least one rule",
		},
		{
			name:    "router without default",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].DefaultRoute = "" },
			wantSub: "requires a defaultRoute",
		},
		{
			name:    "invalid CEL guard",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].Rules[0].When = "this is not cel ((" },
			wantSub: "invalid CEL",
		},
		{
			name:    "non-bool CEL guard",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].Rules[0].When = "state.score + 1" },
			wantSub: "must return bool",
		},
		{
			name: "emittable label without edge",
			mutate: func(d *store.FlowDefinition) {
				d.Nodes[0].Rules = append(d.Nodes[0].Rules, store.FlowRule{When: "state.x == true", Route: "extra"})
			},
			wantSub: "no edge for it",
		},
		{
			name: "edge with route no rule emits",
			mutate: func(d *store.FlowDefinition) {
				d.Edges = append(d.Edges, store.FlowEdge{From: "classify", To: "publish", Route: "bogus"})
			},
			wantSub: "no rule or default emits",
		},
		{
			name: "unconditional edge from router",
			mutate: func(d *store.FlowDefinition) {
				for i := range d.Edges {
					if d.Edges[i].From == "classify" && d.Edges[i].Route == "accept" {
						d.Edges[i].Route = ""
					}
				}
			},
			wantSub: "router edges must be routed",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def := routerFlow()
			tc.mutate(def)
			err := Validate(def)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantSub)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantSub)
			}
		})
	}
}

func TestValidate_RoutedEdgeFromNonRouter(t *testing.T) {
	def := linearFlow()
	for i := range def.Edges {
		if def.Edges[i].From == "a" {
			def.Edges[i].Route = "somelabel"
		}
	}
	err := Validate(def)
	if err == nil {
		t.Fatal("expected error for routed edge from non-router, got nil")
	}
	if !strings.Contains(err.Error(), "is not a router") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_RoutedEdgeIntoJoinIsRejected(t *testing.T) {
	// A router feeding a join through a routed edge: the barrier would wait
	// for a predecessor that may be route-skipped, deadlocking it.
	def := &store.FlowDefinition{
		ID:    "badjoin",
		Name:  "badjoin",
		Entry: "decide",
		Nodes: []store.FlowNode{
			{ID: "decide", Type: store.FlowNodeRouter, DefaultRoute: "skip", Rules: []store.FlowRule{
				{When: "state.go == true", Route: "go"},
			}},
			{ID: "merge", Type: store.FlowNodeJoin},
			{ID: "skipped", Type: store.FlowNodeAgent, AgentID: "ag-skip"},
			{ID: "done", Type: store.FlowNodeAgent, AgentID: "ag-done"},
		},
		Edges: []store.FlowEdge{
			{From: "decide", To: "merge", Route: "go"},
			{From: "decide", To: "skipped", Route: "skip"},
			{From: "merge", To: "done"},
		},
	}
	err := Validate(def)
	if err == nil {
		t.Fatal("expected error for routed edge into join, got nil")
	}
	if !strings.Contains(err.Error(), "joins require unconditional predecessors") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_NoTerminalNode(t *testing.T) {
	// A pure two-node cycle with no exit has no terminal node.
	def := &store.FlowDefinition{
		ID:    "cyc",
		Name:  "cycle",
		Entry: "a",
		Nodes: []store.FlowNode{
			{ID: "a", Type: store.FlowNodeAgent, AgentID: "x"},
			{ID: "b", Type: store.FlowNodeAgent, AgentID: "y"},
		},
		Edges: []store.FlowEdge{
			{From: "a", To: "b"},
			{From: "b", To: "a"},
		},
	}
	err := Validate(def)
	if err == nil {
		t.Fatal("expected error for graph with no terminal node, got nil")
	}
	if !strings.Contains(err.Error(), "no terminal node") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// parallelSubflowFlow exercises a parallel node and a subflow node in a valid
// graph: START -> fanout(parallel) -> sub(subflow) -> done.
func parallelSubflowFlow() *store.FlowDefinition {
	return &store.FlowDefinition{
		ID:    "f5",
		Name:  "parsub",
		Entry: "fanout",
		Nodes: []store.FlowNode{
			{ID: "fanout", Type: store.FlowNodeParallel, AgentID: "ag-map", MaxConcurrency: 4},
			{ID: "sub", Type: store.FlowNodeSubflow, FlowID: "other-flow"},
			{ID: "done", Type: store.FlowNodeAgent, AgentID: "ag-done", ResponseAgent: true},
		},
		Edges: []store.FlowEdge{
			{From: "fanout", To: "sub"},
			{From: "sub", To: "done"},
		},
	}
}

func TestValidate_AcceptsParallelAndSubflow(t *testing.T) {
	if err := Validate(parallelSubflowFlow()); err != nil {
		t.Fatalf("expected valid graph, got error: %v", err)
	}
}

func TestValidate_ParallelAndSubflowRules(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*store.FlowDefinition)
		wantSub string
	}{
		{
			name:    "parallel without agentId",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].AgentID = "" },
			wantSub: "parallel node \"fanout\" requires agentId",
		},
		{
			name:    "parallel negative concurrency",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].MaxConcurrency = -1 },
			wantSub: "negative maxConcurrency",
		},
		{
			name:    "subflow without flowId",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[1].FlowID = "" },
			wantSub: "subflow node \"sub\" requires flowId",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def := parallelSubflowFlow()
			tc.mutate(def)
			err := Validate(def)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantSub)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantSub)
			}
		})
	}
}

// transformFlow: START -> expr(expression) -> tmpl(template) -> done.
func transformFlow() *store.FlowDefinition {
	return &store.FlowDefinition{
		ID:    "f6",
		Name:  "transform",
		Entry: "expr",
		Nodes: []store.FlowNode{
			{ID: "expr", Type: store.FlowNodeExpression, Expression: `input.split(",")`, OutputKey: "items"},
			{ID: "tmpl", Type: store.FlowNodeTemplate, Template: "Got: {{ input }} for {{ state.items }}"},
			{ID: "done", Type: store.FlowNodeAgent, AgentID: "ag-done", ResponseAgent: true},
		},
		Edges: []store.FlowEdge{
			{From: "expr", To: "tmpl"},
			{From: "tmpl", To: "done"},
		},
	}
}

func TestValidate_AcceptsExpressionAndTemplate(t *testing.T) {
	if err := Validate(transformFlow()); err != nil {
		t.Fatalf("expected valid graph, got error: %v", err)
	}
}

func TestValidate_TransformRules(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*store.FlowDefinition)
		wantSub string
	}{
		{
			name:    "expression empty",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].Expression = "" },
			wantSub: "expression node \"expr\" requires an expression",
		},
		{
			name:    "expression invalid CEL",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].Expression = "input.((" },
			wantSub: "invalid CEL",
		},
		{
			name:    "expression bad outputKey (hyphen)",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].OutputKey = "my-key" },
			wantSub: "outputKey \"my-key\" must match",
		},
		{
			name:    "template empty",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[1].Template = "" },
			wantSub: "template node \"tmpl\" requires a template",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def := transformFlow()
			tc.mutate(def)
			err := Validate(def)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantSub)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantSub)
			}
		})
	}
}

// codeFlow returns a minimal valid flow with a single code node.
// START -> code_node (terminal).
func codeFlow() *store.FlowDefinition {
	return &store.FlowDefinition{
		ID:    "f7",
		Name:  "code",
		Entry: "transform",
		Nodes: []store.FlowNode{
			{
				ID:     "transform",
				Type:   store.FlowNodeCode,
				Script: "output = input",
			},
		},
		Edges: nil,
	}
}

func TestValidate_AcceptsCodeNode(t *testing.T) {
	if err := Validate(codeFlow()); err != nil {
		t.Fatalf("expected valid code node flow, got error: %v", err)
	}
}

func TestValidate_AcceptsCodeNodeWithOutputKey(t *testing.T) {
	def := codeFlow()
	def.Nodes[0].OutputKey = "result"
	if err := Validate(def); err != nil {
		t.Fatalf("expected valid code node flow with outputKey, got error: %v", err)
	}
}

func TestValidate_CodeNodeRules(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*store.FlowDefinition)
		wantSub string
	}{
		{
			name:    "missing script",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].Script = "" },
			wantSub: "code node \"transform\" requires a script",
		},
		{
			name:    "bad outputKey with hyphen",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].OutputKey = "my-key" },
			wantSub: "outputKey \"my-key\" must match",
		},
		{
			name:    "negative timeoutMs",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].TimeoutMs = -1 },
			wantSub: "negative timeoutMs",
		},
		{
			name:    "negative maxOutputBytes",
			mutate:  func(d *store.FlowDefinition) { d.Nodes[0].MaxOutputBytes = -1 },
			wantSub: "negative maxOutputBytes",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			def := codeFlow()
			tc.mutate(def)
			err := Validate(def)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantSub)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantSub)
			}
		})
	}
}
