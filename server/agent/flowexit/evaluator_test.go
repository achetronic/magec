// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package flowexit

import (
	"iter"
	"testing"

	"google.golang.org/adk/v2/session"

	toolsflowstate "github.com/achetronic/magec/server/agent/tools/flowstate"
)

func TestEvaluate_True(t *testing.T) {
	prog, err := Compile(`state.approved == true`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if !Evaluate(prog, `state.approved == true`, nil, map[string]any{"approved": true}, 0) {
		t.Fatal("expected true")
	}
}

func TestEvaluate_False(t *testing.T) {
	prog, err := Compile(`state.approved == true`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if Evaluate(prog, `state.approved == true`, nil, map[string]any{"approved": false}, 0) {
		t.Fatal("expected false")
	}
}

func TestEvaluate_IterationsGuard(t *testing.T) {
	// The iterations variable lets an operator cap a loop in CEL. Below the
	// cap the guard is false; at the cap it flips to true.
	const expr = `state.done == true || iterations >= 5`
	prog, err := Compile(expr)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if Evaluate(prog, expr, nil, map[string]any{"done": false}, 4) {
		t.Fatal("expected false below the iteration cap")
	}
	if !Evaluate(prog, expr, nil, map[string]any{"done": false}, 5) {
		t.Fatal("expected true once the iteration cap is reached")
	}
}

func TestEvaluate_RuntimeErrorTreatedAsFalse(t *testing.T) {
	// Compile expects state.score to be comparable to 0.5 — passing a
	// string at runtime triggers a CEL no-such-overload error.
	prog, err := Compile(`state.score > 0.5`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if Evaluate(prog, `state.score > 0.5`, nil, map[string]any{"score": "high"}, 0) {
		t.Fatal("expected false on runtime error")
	}
}

func TestEvaluate_MissingKeyTreatedAsFalse(t *testing.T) {
	prog, err := Compile(`has(state.approved) && state.approved == true`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	// has() guards the access, so missing key returns false, no error.
	if Evaluate(prog, `has(state.approved) && state.approved == true`, nil, map[string]any{}, 0) {
		t.Fatal("expected false when key absent")
	}
}

func TestEvaluate_InputGuard(t *testing.T) {
	// Router guards can branch on the upstream node's output directly.
	const expr = `input.contains("error")`
	prog, err := Compile(expr)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if !Evaluate(prog, expr, "an error happened", map[string]any{}, 0) {
		t.Fatal("expected true when the input contains the needle")
	}
	if Evaluate(prog, expr, "all good", map[string]any{}, 0) {
		t.Fatal("expected false when the input does not contain the needle")
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
		"app:other":                                "ignored",
		"user:name":                                "ignored",
		"unprefixed":                               "ignored",
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

func TestCompileValue_AllowsNonBool(t *testing.T) {
	// Unlike Compile, CompileValue accepts any output type.
	for _, expr := range []string{
		`input`,
		`input.split(",")`,
		`state.name`,
		`"hello " + input`,
		`[input]`,
	} {
		if _, err := CompileValue(expr); err != nil {
			t.Fatalf("CompileValue(%q) unexpected error: %v", expr, err)
		}
	}
}

func TestCompileValue_RejectsGarbage(t *testing.T) {
	if _, err := CompileValue(`this is not (( cel`); err == nil {
		t.Fatal("expected compile error for garbage expression")
	}
}

func TestEvaluateValue_SplitProducesList(t *testing.T) {
	prog, err := CompileValue(`input.split(",")`)
	if err != nil {
		t.Fatalf("CompileValue: %v", err)
	}
	out, err := EvaluateValue(prog, `input.split(",")`, "a,b,c", map[string]any{}, nil)
	if err != nil {
		t.Fatalf("EvaluateValue: %v", err)
	}
	list, ok := out.([]string)
	if !ok {
		// cel-go may return []ref.Val-backed slice; normalise via length check
		if l, ok2 := out.([]any); ok2 {
			if len(l) != 3 {
				t.Fatalf("expected 3 items, got %d", len(l))
			}
			return
		}
		t.Fatalf("expected a slice, got %T (%v)", out, out)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list))
	}
}

func TestEvaluateValue_ReadsState(t *testing.T) {
	prog, err := CompileValue(`state.greeting + " " + input`)
	if err != nil {
		t.Fatalf("CompileValue: %v", err)
	}
	out, err := EvaluateValue(prog, `state.greeting + " " + input`, "world", map[string]any{"greeting": "hello"}, nil)
	if err != nil {
		t.Fatalf("EvaluateValue: %v", err)
	}
	if out != "hello world" {
		t.Fatalf("got %q, want %q", out, "hello world")
	}
}
