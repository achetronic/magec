// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package flowexit

import (
	"iter"
	"testing"

	"google.golang.org/adk/session"

	toolsflowstate "github.com/achetronic/magec/server/agent/tools/flowstate"
)

func TestEvaluateExitWhen_True(t *testing.T) {
	prog, err := Compile(`state.approved == true`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if !EvaluateExitWhen(prog, `state.approved == true`, map[string]any{"approved": true}) {
		t.Fatal("expected true")
	}
}

func TestEvaluateExitWhen_False(t *testing.T) {
	prog, err := Compile(`state.approved == true`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if EvaluateExitWhen(prog, `state.approved == true`, map[string]any{"approved": false}) {
		t.Fatal("expected false")
	}
}

func TestEvaluateExitWhen_RuntimeErrorTreatedAsFalse(t *testing.T) {
	// Compile expects state.score to be comparable to 0.5 — passing a
	// string at runtime triggers a CEL no-such-overload error.
	prog, err := Compile(`state.score > 0.5`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if EvaluateExitWhen(prog, `state.score > 0.5`, map[string]any{"score": "high"}) {
		t.Fatal("expected false on runtime error")
	}
}

func TestEvaluateExitWhen_MissingKeyTreatedAsFalse(t *testing.T) {
	prog, err := Compile(`has(state.approved) && state.approved == true`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	// has() guards the access, so missing key returns false, no error.
	if EvaluateExitWhen(prog, `has(state.approved) && state.approved == true`, map[string]any{}) {
		t.Fatal("expected false when key absent")
	}
}

// fakeReadonlyState is a minimal session.State backed by a map. We only
// implement what ExtractFlowState touches.
type fakeReadonlyState struct {
	store map[string]any
}

func (s *fakeReadonlyState) Get(key string) (any, error) {
	if v, ok := s.store[key]; ok {
		return v, nil
	}
	return nil, nil
}

func (s *fakeReadonlyState) Set(key string, val any) error {
	s.store[key] = val
	return nil
}

func (s *fakeReadonlyState) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for k, v := range s.store {
			if !yield(k, v) {
				return
			}
		}
	}
}

func TestExtractFlowState_StripsPrefix(t *testing.T) {
	s := &fakeReadonlyState{store: map[string]any{
		toolsflowstate.StateKeyPrefix + "approved": true,
		toolsflowstate.StateKeyPrefix + "score":    0.9,
		"app:other":   "ignored",
		"user:name":   "ignored",
		"unprefixed":  "ignored",
	}}
	got := ExtractFlowState(s)
	if len(got) != 2 {
		t.Fatalf("expected 2 keys, got %d (%#v)", len(got), got)
	}
	if got["approved"] != true {
		t.Fatalf("expected approved=true, got %#v", got["approved"])
	}
	if got["score"] != 0.9 {
		t.Fatalf("expected score=0.9, got %#v", got["score"])
	}
}

func TestExtractFlowState_EmptyWhenNoFlowKeys(t *testing.T) {
	s := &fakeReadonlyState{store: map[string]any{
		"app:x":  "x",
		"user:y": "y",
	}}
	if got := ExtractFlowState(s); len(got) != 0 {
		t.Fatalf("expected empty map, got %#v", got)
	}
}

// Compile-time assertion: fakeReadonlyState satisfies session.State, so
// ExtractFlowState callers do not have to learn a second mock interface.
var _ session.State = (*fakeReadonlyState)(nil)

func TestNewExitWhenAgent_Construction(t *testing.T) {
	prog, err := Compile(`state.approved == true`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	a, err := NewExitWhenAgent("loop_exit", prog, `state.approved == true`)
	if err != nil {
		t.Fatalf("NewExitWhenAgent: %v", err)
	}
	if a.Name() != "loop_exit" {
		t.Fatalf("expected name 'loop_exit', got %q", a.Name())
	}
}

func TestNewExitWhenAgent_RejectsNilProgram(t *testing.T) {
	if _, err := NewExitWhenAgent("x", nil, ""); err == nil {
		t.Fatal("expected error for nil program")
	}
}
