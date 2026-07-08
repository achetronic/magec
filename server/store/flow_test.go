// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"reflect"
	"sort"
	"testing"
)

// In the graph model a node's ID is its adk Node.Name() and therefore the
// event.Author the response filter matches against, so ResponseAgentNames
// returns node IDs directly — no synthetic naming recipe to keep in lockstep
// with the builder.

func TestResponseAgentNames_ReturnsNodeIDs(t *testing.T) {
	f := FlowDefinition{
		ID:    "flow1",
		Entry: "gen",
		Nodes: []FlowNode{
			{ID: "gen", Type: FlowNodeAgent, AgentID: "agent-a"},
			{ID: "critic", Type: FlowNodeAgent, AgentID: "agent-b", ResponseAgent: true},
		},
		Edges: []FlowEdge{{From: "gen", To: "critic"}},
	}
	got := f.ResponseAgentNames()
	want := []string{"critic"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestResponseAgentNames_MultipleMarked(t *testing.T) {
	f := FlowDefinition{
		ID:    "flow1",
		Entry: "a",
		Nodes: []FlowNode{
			{ID: "a", Type: FlowNodeAgent, AgentID: "agent-a", ResponseAgent: true},
			{ID: "b", Type: FlowNodeAgent, AgentID: "agent-b", ResponseAgent: true},
			{ID: "c", Type: FlowNodeAgent, AgentID: "agent-c"},
		},
		Edges: []FlowEdge{{From: "a", To: "b"}, {From: "b", To: "c"}},
	}
	got := f.ResponseAgentNames()
	sort.Strings(got)
	want := []string{"a", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestResponseAgentNames_NoneMarked(t *testing.T) {
	f := FlowDefinition{
		ID:    "flow1",
		Entry: "a",
		Nodes: []FlowNode{
			{ID: "a", Type: FlowNodeAgent, AgentID: "agent-a"},
			{ID: "b", Type: FlowNodeAgent, AgentID: "agent-b"},
		},
		Edges: []FlowEdge{{From: "a", To: "b"}},
	}
	if got := f.ResponseAgentNames(); len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestResponseAgentIDs_ReturnsAgentDefIDs(t *testing.T) {
	f := FlowDefinition{
		ID:    "flow1",
		Entry: "a",
		Nodes: []FlowNode{
			{ID: "a", Type: FlowNodeAgent, AgentID: "agent-a", ResponseAgent: true},
			{ID: "b", Type: FlowNodeAgent, AgentID: "agent-b"},
		},
	}
	got := f.ResponseAgentIDs()
	want := []string{"agent-a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestFirstAgentID_FromEntryBreadthFirst(t *testing.T) {
	// Entry is a router; the first agent reachable from it is the answer.
	f := FlowDefinition{
		ID:    "flow1",
		Entry: "decide",
		Nodes: []FlowNode{
			{ID: "decide", Type: FlowNodeRouter, Rules: []FlowRule{{When: "true", Route: "go"}}},
			{ID: "worker", Type: FlowNodeAgent, AgentID: "agent-worker"},
		},
		Edges: []FlowEdge{{From: "decide", To: "worker", Route: "go"}},
	}
	if got := f.FirstAgentID(); got != "agent-worker" {
		t.Fatalf("got %q, want %q", got, "agent-worker")
	}
}

func TestFirstAgentID_EntryIsAgent(t *testing.T) {
	f := FlowDefinition{
		ID:    "flow1",
		Entry: "a",
		Nodes: []FlowNode{
			{ID: "a", Type: FlowNodeAgent, AgentID: "agent-a"},
			{ID: "b", Type: FlowNodeAgent, AgentID: "agent-b"},
		},
		Edges: []FlowEdge{{From: "a", To: "b"}},
	}
	if got := f.FirstAgentID(); got != "agent-a" {
		t.Fatalf("got %q, want %q", got, "agent-a")
	}
}

func TestAgentIDs_UniqueAcrossNodes(t *testing.T) {
	f := FlowDefinition{
		ID:    "flow1",
		Entry: "a",
		Nodes: []FlowNode{
			{ID: "a", Type: FlowNodeAgent, AgentID: "shared"},
			{ID: "router", Type: FlowNodeRouter, Rules: []FlowRule{{When: "true", Route: "x"}}},
			{ID: "b", Type: FlowNodeAgent, AgentID: "shared"}, // same AgentID, deduped
			{ID: "c", Type: FlowNodeAgent, AgentID: "other"},
		},
	}
	got := f.AgentIDs()
	sort.Strings(got)
	want := []string{"other", "shared"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAgentIDs_IncludesParallelAndSubflowRefs(t *testing.T) {
	f := FlowDefinition{
		ID:    "flow1",
		Entry: "a",
		Nodes: []FlowNode{
			{ID: "a", Type: FlowNodeAgent, AgentID: "ag-a"},
			{ID: "p", Type: FlowNodeParallel, AgentID: "ag-map"},
			{ID: "s", Type: FlowNodeSubflow, FlowID: "other-flow"},
			{ID: "j", Type: FlowNodeJoin},
		},
	}
	got := f.AgentIDs()
	sort.Strings(got)
	want := []string{"ag-a", "ag-map", "other-flow"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
