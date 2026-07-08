// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/achetronic/magec/server/agent/runrecorder"
)

// TestProjectActivations runs table-driven unit tests on projectActivations.
// It exercises consecutive collapsing, fallbacks, route deduplication, and payload extracting.
func TestProjectActivations(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Second)
	t3 := t2.Add(time.Second)

	tests := []struct {
		name       string
		events     []runrecorder.EventRecord
		runAppName string
		expected   []runActivation
	}{
		{
			name: "consecutive events same author collapse into one activation",
			events: []runrecorder.EventRecord{
				{Seq: 1, Timestamp: t1, Author: "agent-1"},
				{Seq: 2, Timestamp: t2, Author: "agent-1"},
			},
			runAppName: "my-app",
			expected: []runActivation{
				{
					Node:      "agent-1",
					Seq:       1,
					StartedAt: t1,
					EndedAt:   t2,
					Events:    2,
				},
			},
		},
		{
			name: "author changing A B A yields three activations",
			events: []runrecorder.EventRecord{
				{Seq: 1, Timestamp: t1, Author: "A"},
				{Seq: 2, Timestamp: t2, Author: "B"},
				{Seq: 3, Timestamp: t3, Author: "A"},
			},
			runAppName: "my-app",
			expected: []runActivation{
				{Node: "A", Seq: 1, StartedAt: t1, EndedAt: t1, Events: 1},
				{Node: "B", Seq: 2, StartedAt: t2, EndedAt: t2, Events: 1},
				{Node: "A", Seq: 3, StartedAt: t3, EndedAt: t3, Events: 1},
			},
		},
		{
			name: "event with empty author or identical to appName falls back to NodePath",
			events: []runrecorder.EventRecord{
				{Seq: 1, Timestamp: t1, Author: "", NodePath: "path/to/node1"},
				{Seq: 2, Timestamp: t2, Author: "my-app", NodePath: "path/to/node2"},
				{Seq: 3, Timestamp: t3, Author: "", NodePath: ""},
			},
			runAppName: "my-app",
			expected: []runActivation{
				{Node: "node1", Seq: 1, StartedAt: t1, EndedAt: t1, Events: 1},
				{Node: "node2", Seq: 2, StartedAt: t2, EndedAt: t2, Events: 1},
				{Node: "workflow", Seq: 3, StartedAt: t3, EndedAt: t3, Events: 1},
			},
		},
		{
			name: "routes deduplicated",
			events: []runrecorder.EventRecord{
				{Seq: 1, Timestamp: t1, Author: "A", Routes: []string{"route1", "route2"}},
				{Seq: 2, Timestamp: t2, Author: "A", Routes: []string{"route2", "route3"}},
			},
			runAppName: "my-app",
			expected: []runActivation{
				{
					Node:      "A",
					Seq:       1,
					StartedAt: t1,
					EndedAt:   t2,
					Events:    2,
					Routes:    []string{"route1", "route2", "route3"},
				},
			},
		},
		{
			name: "output preview extracted and truncated",
			events: []runrecorder.EventRecord{
				{
					Seq:       1,
					Timestamp: t1,
					Author:    "A",
					Payload:   []byte(`{"output": "hello"}`),
				},
				{
					Seq:       2,
					Timestamp: t2,
					Author:    "A",
					Payload:   []byte(`{"output": "` + strings.Repeat("x", 210) + `"}`),
				},
			},
			runAppName: "my-app",
			expected: []runActivation{
				{
					Node:          "A",
					Seq:           1,
					StartedAt:     t1,
					EndedAt:       t2,
					Events:        2,
					OutputPreview: strings.Repeat("x", 200) + "...",
				},
			},
		},
		{
			name: "errorMessage in payload lands in Error",
			events: []runrecorder.EventRecord{
				{
					Seq:       1,
					Timestamp: t1,
					Author:    "A",
					Payload:   []byte(`{"errorMessage": "something failed"}`),
				},
			},
			runAppName: "my-app",
			expected: []runActivation{
				{
					Node:      "A",
					Seq:       1,
					StartedAt: t1,
					EndedAt:   t1,
					Events:    1,
					Error:     "something failed",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := projectActivations(tc.events, tc.runAppName, "", nil)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected %+v, got %+v", tc.expected, got)
			}
		})
	}
}

// TestProjectActivations_DerivedInputAndState verifies the derived chain: an
// activation's InputPreview is the previous activation's output, StateDelta
// captures flow-prefixed writes with internal keys hidden, and StateAfter
// accumulates across activations.
func TestProjectActivations_DerivedInputAndState(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	events := []runrecorder.EventRecord{
		{Seq: 0, Timestamp: t1, Author: "writer",
			Payload: []byte(`{"Output":"draft v1","Actions":{"StateDelta":{"flow:score":0.9,"flow:__iter__router_1":1,"other:key":"hidden"}}}`)},
		{Seq: 1, Timestamp: t1.Add(time.Second), Author: "router_1",
			Payload: []byte(`{"Output":"draft v1"}`)},
		{Seq: 2, Timestamp: t1.Add(2 * time.Second), Author: "publisher",
			Payload: []byte(`{"output":"published","Actions":{"StateDelta":{"flow:done":true}}}`)},
	}

	got := projectActivations(events, "my-flow", "", nil)

	if len(got) != 3 {
		t.Fatalf("expected 3 activations, got %d", len(got))
	}
	if got[0].InputPreview != "" {
		t.Fatalf("first activation should have no derived input, got %q", got[0].InputPreview)
	}
	if got[1].InputPreview != "draft v1" {
		t.Fatalf("router input should be writer output, got %q", got[1].InputPreview)
	}
	if got[2].InputPreview != "draft v1" {
		t.Fatalf("publisher input should be router passthrough, got %q", got[2].InputPreview)
	}

	wantDelta := map[string]any{"score": 0.9}
	if !reflect.DeepEqual(got[0].StateDelta, wantDelta) {
		t.Fatalf("writer stateDelta mismatch (internal and non-flow keys must be hidden): %+v", got[0].StateDelta)
	}
	if got[1].StateDelta != nil {
		t.Fatalf("router wrote no state, got %+v", got[1].StateDelta)
	}

	if !reflect.DeepEqual(got[1].StateAfter, map[string]any{"score": 0.9}) {
		t.Fatalf("stateAfter should carry accumulated state through activations: %+v", got[1].StateAfter)
	}
	wantFinal := map[string]any{"score": 0.9, "done": true}
	if !reflect.DeepEqual(got[2].StateAfter, wantFinal) {
		t.Fatalf("final stateAfter mismatch: %+v", got[2].StateAfter)
	}
}

// TestBuildActivation_AgentContentTextFallback verifies that an agent event
// whose answer travels as model content text (no workflow Output) still
// produces an output preview.
func TestBuildActivation_AgentContentTextFallback(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	events := []runrecorder.EventRecord{
		{Seq: 0, Timestamp: t1, Author: "agent_1",
			Payload: []byte(`{"Content":{"role":"model","parts":[{"functionCall":{"name":"set_state"}}]}}`)},
		{Seq: 1, Timestamp: t1.Add(time.Second), Author: "agent_1",
			Payload: []byte(`{"Content":{"role":"model","parts":[{"text":"the final "},{"text":"answer"}]}}`)},
	}

	got := projectActivations(events, "my-flow", "", nil)

	if len(got) != 1 {
		t.Fatalf("expected 1 activation, got %d", len(got))
	}
	if got[0].OutputPreview != "the final answer" {
		t.Fatalf("expected content text as output preview, got %q", got[0].OutputPreview)
	}
}

// TestProjectActivations_RunInputSeedsFirstActivation verifies that the run's
// own input becomes the first activation's input, which is how standalone
// agent runs (a single activation) surface what the user asked.
func TestProjectActivations_RunInputSeedsFirstActivation(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	events := []runrecorder.EventRecord{
		{Seq: 0, Timestamp: t1, Author: "agent_1",
			Payload: []byte(`{"Content":{"role":"model","parts":[{"text":"the answer"}]}}`)},
	}

	got := projectActivations(events, "my-agent", "what is the answer?", nil)

	if len(got) != 1 {
		t.Fatalf("expected 1 activation, got %d", len(got))
	}
	if got[0].InputPreview != "what is the answer?" {
		t.Fatalf("run input must seed the first activation, got %q", got[0].InputPreview)
	}
}

// TestProjectActivations_NodeTypesResolved guards the type lookup: activations
// keyed by node ID resolve directly, join activations keyed by a composite
// NodeInfo path resolve through their last path segment, and unknown keys
// stay untyped.
func TestProjectActivations_NodeTypesResolved(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	events := []runrecorder.EventRecord{
		{Seq: 1, Timestamp: t1, Author: "prep"},
		// A pure join node's Author is the workflow name, so the projection
		// falls back to its NodePath as the activation key.
		{Seq: 2, Timestamp: t1.Add(time.Second), Author: "my-flow", NodePath: "my-flow@1/gather@1"},
		{Seq: 3, Timestamp: t1.Add(2 * time.Second), Author: "mystery"},
	}
	nodeTypes := map[string]string{"prep": "expression", "gather": "join"}

	got := projectActivations(events, "my-flow", "", nodeTypes)

	if len(got) != 3 {
		t.Fatalf("expected 3 activations, got %d", len(got))
	}
	if got[0].NodeType != "expression" {
		t.Fatalf("direct node ID lookup failed: %q", got[0].NodeType)
	}
	if got[1].NodeType != "join" {
		t.Fatalf("composite path lookup failed: %q", got[1].NodeType)
	}
	if got[1].Node != "gather" {
		t.Fatalf("composite path must be shortened to the node segment, got %q", got[1].Node)
	}
	if got[2].NodeType != "" {
		t.Fatalf("unknown node must stay untyped, got %q", got[2].NodeType)
	}
}

// TestProjectActivations_InternalNodesHidden guards the plumbing filter: the
// __meta__ prefilter activation must not surface in the timeline, but the
// state it wrote must still appear on downstream activations.
func TestProjectActivations_InternalNodesHidden(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	metaPayload := []byte(`{"Author":"__meta__","Output":"clean input","Actions":{"StateDelta":{"flow:magec_meta":{"source":"telegram"}}}}`)
	events := []runrecorder.EventRecord{
		{Seq: 1, Timestamp: t1, Author: "__meta__", Payload: metaPayload},
		{Seq: 2, Timestamp: t1.Add(time.Second), Author: "prep"},
	}

	got := projectActivations(events, "my-flow", "raw input", nil)

	if len(got) != 1 {
		t.Fatalf("expected the internal node to be hidden, got %d activations", len(got))
	}
	if got[0].Node != "prep" {
		t.Fatalf("expected prep, got %q", got[0].Node)
	}
	if got[0].InputPreview != "clean input" {
		t.Fatalf("derived input must chain through the hidden node, got %q", got[0].InputPreview)
	}
	if _, ok := got[0].StateAfter["magec_meta"]; !ok {
		t.Fatalf("state written by the hidden node must surface downstream: %+v", got[0].StateAfter)
	}
}
