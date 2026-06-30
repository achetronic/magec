// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"reflect"
	"sort"
	"testing"
)

// The synthetic ADK agent names produced by ResponseAgentNames must match
// exactly the naming recipe used by server/agent/flow.go when it builds
// per-appearance ADK instances. If these tests fail, the FlowResponseFilter
// will silently drop every event because the runtime author values won't
// be in the allow-set.

func TestResponseAgentNames_SingleAgentRoot(t *testing.T) {
	f := FlowDefinition{
		ID: "flow1",
		Root: FlowStep{
			Type:          FlowStepAgent,
			AgentID:       "agent-a",
			ResponseAgent: true,
		},
	}
	got := f.ResponseAgentNames()
	want := []string{"flow1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestResponseAgentNames_SequentialChildren(t *testing.T) {
	f := FlowDefinition{
		ID: "flow1",
		Root: FlowStep{
			Type: FlowStepSequential,
			Steps: []FlowStep{
				{Type: FlowStepAgent, AgentID: "a", ResponseAgent: false},
				{Type: FlowStepAgent, AgentID: "b", ResponseAgent: true},
			},
		},
	}
	got := f.ResponseAgentNames()
	want := []string{"flow1_1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestResponseAgentNames_NestedLoopWithSequential(t *testing.T) {
	// loop → sequential → [Generator, Critic{response}]
	f := FlowDefinition{
		ID: "flow1",
		Root: FlowStep{
			Type: FlowStepLoop,
			Steps: []FlowStep{
				{
					Type: FlowStepSequential,
					Steps: []FlowStep{
						{Type: FlowStepAgent, AgentID: "gen"},
						{Type: FlowStepAgent, AgentID: "critic", ResponseAgent: true},
					},
				},
			},
		},
	}
	got := f.ResponseAgentNames()
	// path: loop=root → sequential is child 0 → critic is child 1 of seq.
	// stepName = "flow1_0_1"
	want := []string{"flow1_0_1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestResponseAgentNames_MultipleResponseAgents(t *testing.T) {
	f := FlowDefinition{
		ID: "flow1",
		Root: FlowStep{
			Type: FlowStepSequential,
			Steps: []FlowStep{
				{Type: FlowStepAgent, AgentID: "a", ResponseAgent: true},
				{
					Type: FlowStepParallel,
					Steps: []FlowStep{
						{Type: FlowStepAgent, AgentID: "b", ResponseAgent: true},
						{Type: FlowStepAgent, AgentID: "c"},
					},
				},
			},
		},
	}
	got := f.ResponseAgentNames()
	sort.Strings(got)
	want := []string{"flow1_0", "flow1_1_0"}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestResponseAgentNames_NoneMarked(t *testing.T) {
	f := FlowDefinition{
		ID: "flow1",
		Root: FlowStep{
			Type: FlowStepSequential,
			Steps: []FlowStep{
				{Type: FlowStepAgent, AgentID: "a"},
				{Type: FlowStepAgent, AgentID: "b"},
			},
		},
	}
	if got := f.ResponseAgentNames(); len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}
