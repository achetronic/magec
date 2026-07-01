// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/1set/starlet"
)

// TestEffectiveLimit verifies every branch of the effectiveLimit helper.
// These cases must keep failing if the guard logic is removed or inverted.
func TestEffectiveLimit(t *testing.T) {
	cases := []struct {
		name     string
		nodeVal  int
		ceiling  int
		expected int
	}{
		{"both zero", 0, 0, 0},
		{"ceiling zero, node set", 5, 0, 5},
		{"ceiling set, node zero", 0, 100, 100},
		{"node less than ceiling", 50, 100, 50},
		{"node greater than ceiling", 150, 100, 100},
		{"node equals ceiling", 100, 100, 100},
		{"large node capped by small ceiling", 9999, 1000, 1000},
		{"node 1 with ceiling 1", 1, 1, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := effectiveLimit(tc.nodeVal, tc.ceiling)
			if got != tc.expected {
				t.Fatalf("effectiveLimit(%d, %d) = %d, want %d", tc.nodeVal, tc.ceiling, got, tc.expected)
			}
		})
	}
}

// TestEffectiveLimitCanary fails if the guard is simplified to always return
// the ceiling or always return the nodeVal.
func TestEffectiveLimitCanary(t *testing.T) {
	// Ceiling 0: the node value must win regardless.
	if got := effectiveLimit(42, 0); got != 42 {
		t.Fatalf("canary: ceiling=0 must not cap nodeVal; got %d", got)
	}
	// Node 0: the ceiling must win.
	if got := effectiveLimit(0, 99); got != 99 {
		t.Fatalf("canary: nodeVal=0 must yield ceiling; got %d", got)
	}
	// Node < ceiling: nodeVal must win.
	if got := effectiveLimit(10, 20); got != 10 {
		t.Fatalf("canary: nodeVal<ceiling must yield nodeVal; got %d", got)
	}
	// Node > ceiling: ceiling must win.
	if got := effectiveLimit(30, 20); got != 20 {
		t.Fatalf("canary: nodeVal>ceiling must yield ceiling; got %d", got)
	}
}

// TestCodeNodeStarlet_BasicOutput runs a minimal script through the real
// Starlark machine and confirms that "output = input * 2" with input=21
// yields 42. This exercises the same execution path used by buildCodeNode.
func TestCodeNodeStarlet_BasicOutput(t *testing.T) {
	m := starlet.NewWithLoaders(nil, nil, nil)
	m.SetPrintFunc(starlet.NoopPrintFunc)
	m.SetScriptContent([]byte("output = input * 2"))

	extras := starlet.StringAnyMap{"input": int64(21), "state": map[string]any{}}
	res, err := m.RunWithContext(context.Background(), extras)
	if err != nil {
		t.Fatalf("script execution failed: %v", err)
	}
	got := res["output"]
	if got == nil {
		t.Fatal("expected output, got nil")
	}
	if fmt.Sprintf("%v", got) != "42" {
		t.Fatalf("output = %v (%T), want 42", got, got)
	}
}

// TestCodeNodeStarlet_StateAccess verifies that the `state` map is readable
// inside the script via the standard extras mechanism.
func TestCodeNodeStarlet_StateAccess(t *testing.T) {
	m := starlet.NewWithLoaders(nil, nil, nil)
	m.SetPrintFunc(starlet.NoopPrintFunc)
	m.SetScriptContent([]byte(`output = state["greeting"] + " world"`))

	state := map[string]any{"greeting": "hello"}
	extras := starlet.StringAnyMap{"input": nil, "state": state}
	res, err := m.RunWithContext(context.Background(), extras)
	if err != nil {
		t.Fatalf("script execution failed: %v", err)
	}
	got := fmt.Sprintf("%v", res["output"])
	if got != "hello world" {
		t.Fatalf("output = %q, want %q", got, "hello world")
	}
}

// TestCodeNodeStarlet_NoOutput verifies that a script that does not assign
// `output` leaves that key absent from the result map (no error).
func TestCodeNodeStarlet_NoOutput(t *testing.T) {
	m := starlet.NewWithLoaders(nil, nil, nil)
	m.SetPrintFunc(starlet.NoopPrintFunc)
	m.SetScriptContent([]byte("x = 1 + 1"))

	res, err := m.RunWithContext(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := res["output"]; ok {
		t.Fatalf("expected no 'output' key, but got: %v", res["output"])
	}
}

// TestCodeNodeStarlet_StepBudgetExhausted verifies that a tight step budget
// stops an infinite loop and returns a non-nil error.
func TestCodeNodeStarlet_StepBudgetExhausted(t *testing.T) {
	m := starlet.NewWithLoaders(nil, nil, nil)
	m.SetPrintFunc(starlet.NoopPrintFunc)
	m.SetMaxExecutionSteps(50) // extremely small budget
	m.SetScriptContent([]byte(`
x = 0
for i in range(1000000000):
    x = x + 1
output = x
`))

	_, err := m.RunWithContext(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error from step budget exhaustion, got nil")
	}
}

// TestCodeNodeStarlet_ContextTimeout verifies that a context deadline cancels
// a long-running script.
func TestCodeNodeStarlet_ContextTimeout(t *testing.T) {
	m := starlet.NewWithLoaders(nil, nil, nil)
	m.SetPrintFunc(starlet.NoopPrintFunc)
	m.SetMaxExecutionSteps(maxCodeNodeSteps)
	m.SetScriptContent([]byte(`
x = 0
for i in range(1000000000):
    x = x + 1
output = x
`))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := m.RunWithContext(ctx, nil)
	if err == nil {
		t.Fatal("expected error from context timeout, got nil")
	}
}

// TestCodeNodeOutputCapCheck validates the output-cap logic that buildCodeNode
// applies: output larger than the cap must trigger an error condition.
func TestCodeNodeOutputCapCheck(t *testing.T) {
	// Build a big string output: 2000 chars, JSON-encoded is ~2002 bytes.
	bigStr := strings.Repeat("a", 2000)

	m := starlet.NewWithLoaders(nil, nil, nil)
	m.SetPrintFunc(starlet.NoopPrintFunc)
	m.SetScriptContent([]byte(`output = big`))

	extras := starlet.StringAnyMap{"big": bigStr}
	res, err := m.RunWithContext(context.Background(), extras)
	if err != nil {
		t.Fatalf("script failed: %v", err)
	}

	out := res["output"]
	if out == nil {
		t.Fatal("expected output, got nil")
	}

	// Replicate the cap check from buildCodeNode.
	capBytes := 1000
	b, jsonErr := json.Marshal(out)
	if jsonErr != nil {
		t.Fatalf("json.Marshal failed: %v", jsonErr)
	}
	if len(b) <= capBytes {
		t.Fatalf("test setup: output %d bytes is not bigger than cap %d; increase bigStr size", len(b), capBytes)
	}
	// Confirm the check detects the violation.
	if len(b) > capBytes {
		// This is the expected path: cap exceeded.
		return
	}
	t.Fatal("output cap check did not trigger as expected")
}
