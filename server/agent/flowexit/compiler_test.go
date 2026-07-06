// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package flowexit

import (
	"strings"
	"testing"
)

func TestCompile_ValidExpressions(t *testing.T) {
	cases := []string{
		`state.approved == true`,
		`state.score > 0.5 && state.attempts < 5`,
		`has(state.tag)`,
		`size(state.items) > 0`,
		`state.name == "alby"`,
		`!state.blocked`,
		`state.x in ["a", "b", "c"]`,
	}
	for _, expr := range cases {
		t.Run(expr, func(t *testing.T) {
			prog, err := Compile(expr)
			if err != nil {
				t.Fatalf("Compile(%q) failed: %v", expr, err)
			}
			if prog == nil {
				t.Fatalf("Compile(%q) returned nil program", expr)
			}
		})
	}
}

func TestCompile_RejectsEmpty(t *testing.T) {
	if _, err := Compile(""); err == nil {
		t.Fatal("expected error for empty expression")
	}
}

func TestCompile_RejectsSyntaxErrors(t *testing.T) {
	cases := []string{
		`state.x ==`,          // dangling operator
		`state.x ===`,         // bad operator
		`state.x &&& state.y`, // bad operator
		`state.(x)`,           // bad selector
	}
	for _, expr := range cases {
		t.Run(expr, func(t *testing.T) {
			if _, err := Compile(expr); err == nil {
				t.Fatalf("expected syntax error for %q", expr)
			}
		})
	}
}

func TestCompile_RejectsUndeclaredVariable(t *testing.T) {
	if _, err := Compile(`stat.x == true`); err == nil {
		t.Fatal("expected undeclared variable error")
	} else if !strings.Contains(err.Error(), "stat") {
		t.Fatalf("error should mention undeclared variable, got: %v", err)
	}
}

func TestCompile_RejectsNonBooleanOutput(t *testing.T) {
	cases := []string{
		`state.x + 1`,       // int
		`state.name`,        // dyn / string
		`size(state.items)`, // int
	}
	for _, expr := range cases {
		t.Run(expr, func(t *testing.T) {
			if _, err := Compile(expr); err == nil {
				t.Fatalf("expected non-bool error for %q", expr)
			}
		})
	}
}

func TestCompile_ProgramEvaluatesAgainstStateMap(t *testing.T) {
	prog, err := Compile(`state.approved == true && state.score > 0.5`)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	out, _, err := prog.Eval(map[string]any{
		"state": map[string]any{"approved": true, "score": 0.8},
	})
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}
	if out.Value() != true {
		t.Fatalf("expected true, got %#v", out.Value())
	}

	out2, _, err := prog.Eval(map[string]any{
		"state": map[string]any{"approved": true, "score": 0.3},
	})
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}
	if out2.Value() != false {
		t.Fatalf("expected false, got %#v", out2.Value())
	}
}
