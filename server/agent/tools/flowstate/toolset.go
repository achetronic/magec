// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Package flowstate exposes the set_state and get_state tools, which let
// agents that run inside a flow share a key/value scratchpad backed by the
// underlying ADK session state.
//
// Keys are stored under the "flow:" prefix to keep flow-shared values
// isolated from ContextGuard summaries (which use the agent name as key) and
// from arbitrary outputKey writes (which use whatever the operator named).
// The prefix is internal: callers (agents) only see plain keys; the toolset
// adds and strips the prefix transparently.
//
// See decision #28 in .agents/DECISIONS.md for the full rationale.
package flowstate

import (
	"fmt"
	"regexp"

	"github.com/google/jsonschema-go/jsonschema"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
)

// StateKeyPrefix namespaces every key written by this toolset inside the
// shared session state. Exported because the loop wrapper in
// server/agent/flowexit needs to filter session state down to the same
// namespace before evaluating user-supplied CEL expressions.
const StateKeyPrefix = "flow:"

// keyPattern matches the subset of key names we accept from the LLM. It
// rules out colons (which would let the model escape the namespace into
// "app:" or "user:" tiers) and any character that is not a typical
// identifier. The pattern is intentionally restrictive: flow-shared state
// is not meant to carry structured paths.
var keyPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Toolset bundles the set_state and get_state tools. It implements
// tool.Toolset and is wired into every agent that participates in a flow,
// regardless of the flow shape (sequential, parallel, loop, or nested).
type Toolset struct {
	tools []tool.Tool
}

// anyValueSchema describes a value of any JSON type as an explicit union.
// Schemas inferred from `any` fields become the boolean schema `true`, which
// Ollama's request parser rejects and strict jinja chat templates choke on.
var anyValueSchema = &jsonschema.Schema{
	Types:       []string{"string", "number", "integer", "boolean", "array", "object", "null"},
	Description: "The value to store. Any JSON type is accepted.",
}

// NewToolset builds the set_state/get_state tools. The toolset is stateless
// and safe to share across all agent appearances inside all flows.
func NewToolset() (*Toolset, error) {
	ts := &Toolset{}

	setTool, err := functiontool.New(
		functiontool.Config{
			Name: "set_state",
			Description: "Record a value in the shared flow state so other agents in the same workflow can read it. " +
				"Keys must be simple identifiers (letters, digits, underscores). " +
				"Values can be strings, numbers, booleans, lists or objects. " +
				"State persists for the duration of the conversation and is visible to every other agent in the same flow.",
			InputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"key":   {Type: "string", Description: "Identifier: letters, digits, underscores."},
					"value": anyValueSchema,
				},
				Required: []string{"key", "value"},
			},
		},
		setState,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create set_state tool: %w", err)
	}

	getTool, err := functiontool.New(
		functiontool.Config{
			Name: "get_state",
			Description: "Read a value previously stored in the shared flow state by another agent (or by an earlier turn of this agent). " +
				"Returns {found: true, value: ...} when the key exists and {found: false} when it does not. " +
				"Use this to check signals, decisions or intermediate results produced by upstream agents.",
			// Explicit because GetResult.Value is `any`; see anyValueSchema.
			OutputSchema: &jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"found":   {Type: "boolean", Description: "Whether the key exists in the flow state."},
					"value":   anyValueSchema,
					"message": {Type: "string", Description: "Optional validation message."},
				},
				Required: []string{"found"},
			},
		},
		getState,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create get_state tool: %w", err)
	}

	ts.tools = []tool.Tool{setTool, getTool}
	return ts, nil
}

// Name implements tool.Toolset.
func (ts *Toolset) Name() string {
	return "flow_state_toolset"
}

// Tools implements tool.Toolset.
func (ts *Toolset) Tools(_ agent.ReadonlyContext) ([]tool.Tool, error) {
	return ts.tools, nil
}

// SetArgs is the input schema exposed to the LLM for set_state.
type SetArgs struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// SetResult is the structured response returned to the LLM after a set.
type SetResult struct {
	Success bool   `json:"success"`
	Key     string `json:"key,omitempty"`
	Message string `json:"message"`
}

// setState writes args.Value into the session state under the namespaced
// key. We go through tool.Context.State().Set, which routes the write
// through the current event's StateDelta first (so a subsequent get_state
// in the same turn observes the new value) and then through the session
// service for persistence. Validation failures are returned as
// SetResult{Success:false} rather than as Go errors so the LLM gets a
// structured message it can act on.
func setState(ctx agent.Context, args SetArgs) (SetResult, error) {
	if args.Key == "" {
		return SetResult{Success: false, Message: "key is required"}, nil
	}
	if !keyPattern.MatchString(args.Key) {
		return SetResult{Success: false, Message: "key must match [a-zA-Z_][a-zA-Z0-9_]*"}, nil
	}
	if err := ctx.State().Set(StateKeyPrefix+args.Key, args.Value); err != nil {
		return SetResult{Success: false, Message: fmt.Sprintf("failed to write state: %v", err)}, nil
	}
	return SetResult{
		Success: true,
		Key:     args.Key,
		Message: fmt.Sprintf("State key %q updated.", args.Key),
	}, nil
}

// GetArgs is the input schema exposed to the LLM for get_state.
type GetArgs struct {
	Key string `json:"key"`
}

// GetResult is the structured response returned to the LLM after a get.
// Found is false when the key is absent or unreadable; Value is omitted
// from the JSON serialisation in that case.
type GetResult struct {
	Found   bool   `json:"found"`
	Value   any    `json:"value,omitempty"`
	Message string `json:"message,omitempty"`
}

// getState reads args.Key from the session state. We go through
// tool.Context.State() rather than ReadonlyState() so the lookup observes
// values written in the current event's StateDelta, i.e. a get_state
// after a set_state inside the same turn returns the freshly written
// value, not the pre-event one. A missing key is not an error: it returns
// {found: false} so the model can branch on absence without spurious
// failures.
func getState(ctx agent.Context, args GetArgs) (GetResult, error) {
	if args.Key == "" {
		return GetResult{Found: false, Message: "key is required"}, nil
	}
	if !keyPattern.MatchString(args.Key) {
		return GetResult{Found: false, Message: "key must match [a-zA-Z_][a-zA-Z0-9_]*"}, nil
	}
	val, err := ctx.State().Get(StateKeyPrefix + args.Key)
	if err != nil || val == nil {
		return GetResult{Found: false}, nil
	}
	return GetResult{Found: true, Value: val}, nil
}
