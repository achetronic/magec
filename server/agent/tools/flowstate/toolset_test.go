// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package flowstate

import (
	"context"
	"encoding/json"
	"errors"
	"iter"
	"testing"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
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

// fakeToolContext embeds agent.StrictContextMock so it stays compatible as the
// unified context surface grows: only the methods the flowstate tools actually
// touch (State, ReadonlyState, Actions) are overridden; everything else panics
// loudly if a test accidentally starts using it.
type fakeToolContext struct {
	agent.StrictContextMock
	state   *fakeState
	actions *session.EventActions
}

func newFakeToolContext() *fakeToolContext {
	return &fakeToolContext{
		StrictContextMock: agent.NewStrictContextMock(context.Background()),
		state:             newFakeState(),
		actions:           &session.EventActions{StateDelta: map[string]any{}},
	}
}

func (c *fakeToolContext) State() session.State                 { return c.state }
func (c *fakeToolContext) ReadonlyState() session.ReadonlyState { return c.state }
func (c *fakeToolContext) Actions() *session.EventActions       { return c.actions }

// Compile-time check that fakeToolContext implements the unified agent.Context.
var _ agent.Context = (*fakeToolContext)(nil)

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

// declarationOf digs the genai FunctionDeclaration out of a tool via the
// unexported runnableTool shape adk uses.
func declarationOf(t *testing.T, tl any) *genai.FunctionDeclaration {
	t.Helper()
	d, ok := tl.(interface {
		Declaration() *genai.FunctionDeclaration
	})
	if !ok {
		t.Fatalf("tool %T does not expose Declaration()", tl)
	}
	return d.Declaration()
}

// schemaAsMap marshals a declared schema and requires it to be a JSON object
// at the top level.
func schemaAsMap(t *testing.T, name string, schema any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("%s: marshal schema: %v", name, err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("%s: schema is not an object (boolean schema at top level?): %v", name, err)
	}
	return m
}

// assertNoBooleanProperties fails when any property of the schema is not an
// object schema.
func assertNoBooleanProperties(t *testing.T, name string, m map[string]any) {
	t.Helper()
	props, ok := m["properties"].(map[string]any)
	if !ok {
		t.Fatalf("%s: schema has no properties object: %v", name, m)
	}
	for propName, prop := range props {
		if _, ok := prop.(map[string]any); !ok {
			t.Errorf("%s: property %q is not an object schema (got %T: %v) - boolean schemas break Ollama and strict jinja templates", name, propName, prop, prop)
		}
	}
}

// TestSchemas_NoBooleanProperties: every property of every flowstate tool
// schema, input and output, must be an object schema.
func TestSchemas_NoBooleanProperties(t *testing.T) {
	ts, err := NewToolset()
	if err != nil {
		t.Fatalf("NewToolset: %v", err)
	}
	tools, err := ts.Tools(nil)
	if err != nil {
		t.Fatalf("Tools: %v", err)
	}
	for _, tl := range tools {
		decl := declarationOf(t, tl)
		if decl.ParametersJsonSchema != nil {
			assertNoBooleanProperties(t, tl.Name()+" input", schemaAsMap(t, tl.Name(), decl.ParametersJsonSchema))
		}
		if decl.ResponseJsonSchema != nil {
			assertNoBooleanProperties(t, tl.Name()+" output", schemaAsMap(t, tl.Name(), decl.ResponseJsonSchema))
		}
		if decl.ParametersJsonSchema == nil && decl.ResponseJsonSchema == nil {
			t.Errorf("%s: declaration exposes no schemas at all", tl.Name())
		}
	}
}

// TestSetStateSchema_ValueAcceptsEveryJSONType: the explicit union must
// accept every JSON type.
func TestSetStateSchema_ValueAcceptsEveryJSONType(t *testing.T) {
	ts, _ := NewToolset()
	tools, _ := ts.Tools(nil)
	for _, tl := range tools {
		if tl.Name() != "set_state" {
			continue
		}
		schema := schemaAsMap(t, "set_state", declarationOf(t, tl).ParametersJsonSchema)
		props := schema["properties"].(map[string]any)
		value, ok := props["value"].(map[string]any)
		if !ok {
			t.Fatalf("value is not an object schema: %v", props["value"])
		}
		types, ok := value["type"].([]any)
		if !ok {
			t.Fatalf("value.type is not a union: %v", value["type"])
		}
		want := map[string]bool{"string": true, "number": true, "integer": true, "boolean": true, "array": true, "object": true, "null": true}
		for _, tp := range types {
			delete(want, tp.(string))
		}
		if len(want) != 0 {
			t.Fatalf("value.type union is missing %v", want)
		}
		return
	}
	t.Fatal("set_state tool not found")
}
