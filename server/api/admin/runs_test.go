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
				{Node: "path/to/node1", Seq: 1, StartedAt: t1, EndedAt: t1, Events: 1},
				{Node: "path/to/node2", Seq: 2, StartedAt: t2, EndedAt: t2, Events: 1},
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
			got := projectActivations(tc.events, tc.runAppName)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected %+v, got %+v", tc.expected, got)
			}
		})
	}
}
