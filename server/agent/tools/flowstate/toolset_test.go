// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package flowstate

import (
	"context"
	"errors"
	"iter"
	"testing"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/toolconfirmation"
	"google.golang.org/genai"
)

// fakeState is a tiny session.State implementation backed by a map. It
// also routes Get through the supplied delta first, mirroring the
// real callbackContextState semantics that the toolset relies on.
type fakeState struct {
	store map[string]any
	delta map[string]any
}

func newFakeState() *fakeState {
	return &fakeState{store: map[string]any{}, delta: map[string]any{}}
}

func (s *fakeState) Get(key string) (any, error) {
	if v, ok := s.delta[key]; ok {
		return v, nil
	}
	if v, ok := s.store[key]; ok {
		return v, nil
	}
	return nil, errors.New("not found")
}

func (s *fakeState) Set(key string, val any) error {
	s.delta[key] = val
	s.store[key] = val
	return nil
}

func (s *fakeState) All() iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for k, v := range s.store {
			if !yield(k, v) {
				return
			}
		}
	}
}

// fakeToolContext is the smallest tool.Context that satisfies the
// interfaces touched by the flowstate tools. Anything we do not exercise
// stays nil — the tests fail loudly if we accidentally start using it.
type fakeToolContext struct {
	context.Context
	state   *fakeState
	actions *session.EventActions
}

func newFakeToolContext() *fakeToolContext {
	return &fakeToolContext{
		Context: context.Background(),
		state:   newFakeState(),
		actions: &session.EventActions{StateDelta: map[string]any{}},
	}
}

func (c *fakeToolContext) State() session.State                 { return c.state }
func (c *fakeToolContext) ReadonlyState() session.ReadonlyState { return c.state }
func (c *fakeToolContext) Actions() *session.EventActions       { return c.actions }
func (c *fakeToolContext) Artifacts() agent.Artifacts           { return nil }
func (c *fakeToolContext) FunctionCallID() string               { return "fc-1" }
func (c *fakeToolContext) SearchMemory(context.Context, string) (*memory.SearchResponse, error) {
	return nil, nil
}
func (c *fakeToolContext) AgentName() string                              { return "test-agent" }
func (c *fakeToolContext) InvocationID() string                           { return "inv-1" }
func (c *fakeToolContext) UserContent() *genai.Content                    { return nil }
func (c *fakeToolContext) AppName() string                                { return "test-app" }
func (c *fakeToolContext) Branch() string                                 { return "" }
func (c *fakeToolContext) SessionID() string                              { return "sess-1" }
func (c *fakeToolContext) UserID() string                                 { return "user-1" }
func (c *fakeToolContext) ToolConfirmation() *toolconfirmation.ToolConfirmation { return nil }
func (c *fakeToolContext) RequestConfirmation(string, any) error          { return nil }

// Compile-time check that fakeToolContext implements tool.Context.
var _ tool.Context = (*fakeToolContext)(nil)

func TestSetState_RoundTrip(t *testing.T) {
	ctx := newFakeToolContext()

	res, err := setState(ctx, SetArgs{Key: "approved", Value: true})
	if err != nil {
		t.Fatalf("setState returned error: %v", err)
	}
	if !res.Success {
		t.Fatalf("expected success, got %#v", res)
	}

	stored, ok := ctx.state.store[StateKeyPrefix+"approved"]
	if !ok {
		t.Fatalf("expected key %q in state", StateKeyPrefix+"approved")
	}
	if stored != true {
		t.Fatalf("expected stored value true, got %#v", stored)
	}

	got, err := getState(ctx, GetArgs{Key: "approved"})
	if err != nil {
		t.Fatalf("getState returned error: %v", err)
	}
	if !got.Found || got.Value != true {
		t.Fatalf("expected found=true value=true, got %#v", got)
	}
}

func TestSetState_RejectInvalidKeys(t *testing.T) {
	cases := []struct {
		name string
		key  string
	}{
		{"empty", ""},
		{"with-colon", "flow:foo"},
		{"app-prefix", "app:foo"},
		{"leading-digit", "1foo"},
		{"with-space", "foo bar"},
		{"with-dash", "foo-bar"},
		{"with-dot", "foo.bar"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newFakeToolContext()
			res, err := setState(ctx, SetArgs{Key: tc.key, Value: "v"})
			if err != nil {
				t.Fatalf("unexpected go error: %v", err)
			}
			if res.Success {
				t.Fatalf("expected rejection for key %q, got success", tc.key)
			}
			if len(ctx.state.store) != 0 {
				t.Fatalf("nothing should have been written, got %v", ctx.state.store)
			}
		})
	}
}

func TestGetState_MissingKeyReturnsNotFound(t *testing.T) {
	ctx := newFakeToolContext()
	res, err := getState(ctx, GetArgs{Key: "missing"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Found {
		t.Fatalf("expected found=false, got %#v", res)
	}
}

func TestGetState_RejectInvalidKey(t *testing.T) {
	ctx := newFakeToolContext()
	res, err := getState(ctx, GetArgs{Key: "flow:approved"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Found {
		t.Fatalf("expected rejection, got %#v", res)
	}
}

func TestSetState_KeysAreNamespaced(t *testing.T) {
	ctx := newFakeToolContext()
	if _, err := setState(ctx, SetArgs{Key: "x", Value: 42}); err != nil {
		t.Fatalf("setState: %v", err)
	}
	for k := range ctx.state.store {
		if len(k) < len(StateKeyPrefix) || k[:len(StateKeyPrefix)] != StateKeyPrefix {
			t.Fatalf("expected stored key to carry %q prefix, got %q", StateKeyPrefix, k)
		}
	}
}

func TestNewToolset_RegistersBothTools(t *testing.T) {
	ts, err := NewToolset()
	if err != nil {
		t.Fatalf("NewToolset: %v", err)
	}
	tools, err := ts.Tools(nil)
	if err != nil {
		t.Fatalf("Tools: %v", err)
	}
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	names := map[string]bool{}
	for _, tl := range tools {
		names[tl.Name()] = true
	}
	if !names["set_state"] || !names["get_state"] {
		t.Fatalf("expected set_state and get_state, got %v", names)
	}
}
