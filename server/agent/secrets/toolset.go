// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"iter"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/genai"
)

// runnableTool is the shape adk requires to execute a tool. Declared locally
// because the upstream interface lives in an internal package.
type runnableTool interface {
	tool.Tool
	Declaration() *genai.FunctionDeclaration
	Run(ctx agent.Context, args any) (map[string]any, error)
}

// streamingTool is the shape adk requires for streaming tools.
type streamingTool interface {
	tool.Tool
	Declaration() *genai.FunctionDeclaration
	RunStream(ctx agent.Context, args any) iter.Seq2[string, error]
}

// WrapToolset decorates every runnable tool of a toolset so placeholders in
// its arguments expand right before execution and known secret values are
// redacted from its result. Arguments are expanded into a copy: the original
// map is shared with the model's functionCall event and must keep its
// placeholders. Tools with shapes this package does not know pass through
// unwrapped.
func (s *Snapshot) WrapToolset(ts tool.Toolset, allowed map[string]string) tool.Toolset {
	if len(allowed) == 0 {
		return ts
	}
	return &wrappedToolset{inner: ts, snapshot: s, allowed: allowed}
}

// wrappedToolset decorates the tools of an inner toolset.
type wrappedToolset struct {
	inner    tool.Toolset
	snapshot *Snapshot
	allowed  map[string]string
}

// Name implements tool.Toolset.
func (w *wrappedToolset) Name() string {
	return w.inner.Name()
}

// Tools implements tool.Toolset.
func (w *wrappedToolset) Tools(ctx agent.ReadonlyContext) ([]tool.Tool, error) {
	tools, err := w.inner.Tools(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]tool.Tool, len(tools))
	for i, t := range tools {
		out[i] = wrapTool(t, w.snapshot, w.allowed)
	}
	return out, nil
}

// Close forwards to the inner toolset when it is closeable.
func (w *wrappedToolset) Close() error {
	if c, ok := w.inner.(interface{ Close() error }); ok {
		return c.Close()
	}
	return nil
}

// wrapTool picks the decorator matching the tool's execution shape.
func wrapTool(t tool.Tool, s *Snapshot, allowed map[string]string) tool.Tool {
	switch tt := t.(type) {
	case runnableTool:
		return &wrappedTool{inner: tt, snapshot: s, allowed: allowed}
	case streamingTool:
		return &wrappedStreamingTool{inner: tt, snapshot: s, allowed: allowed}
	default:
		return t
	}
}

// wrappedTool expands secrets into a copy of the arguments and redacts the
// result before it becomes a functionResponse event.
type wrappedTool struct {
	inner    runnableTool
	snapshot *Snapshot
	allowed  map[string]string
}

func (t *wrappedTool) Name() string                            { return t.inner.Name() }
func (t *wrappedTool) Description() string                     { return t.inner.Description() }
func (t *wrappedTool) IsLongRunning() bool                     { return t.inner.IsLongRunning() }
func (t *wrappedTool) Declaration() *genai.FunctionDeclaration { return t.inner.Declaration() }

func (t *wrappedTool) Run(ctx agent.Context, args any) (map[string]any, error) {
	expanded := t.snapshot.ExpandValue(args, t.allowed)
	result, err := t.inner.Run(ctx, expanded)
	if err != nil {
		return result, err
	}
	redacted, _ := t.snapshot.RedactValue(result).(map[string]any)
	if redacted == nil {
		return result, nil
	}
	return redacted, nil
}

// wrappedStreamingTool expands secrets into a copy of the arguments and
// redacts every streamed chunk.
type wrappedStreamingTool struct {
	inner    streamingTool
	snapshot *Snapshot
	allowed  map[string]string
}

func (t *wrappedStreamingTool) Name() string                            { return t.inner.Name() }
func (t *wrappedStreamingTool) Description() string                     { return t.inner.Description() }
func (t *wrappedStreamingTool) IsLongRunning() bool                     { return t.inner.IsLongRunning() }
func (t *wrappedStreamingTool) Declaration() *genai.FunctionDeclaration { return t.inner.Declaration() }

func (t *wrappedStreamingTool) RunStream(ctx agent.Context, args any) iter.Seq2[string, error] {
	expanded := t.snapshot.ExpandValue(args, t.allowed)
	return func(yield func(string, error) bool) {
		for chunk, err := range t.inner.RunStream(ctx, expanded) {
			if !yield(t.snapshot.RedactString(chunk), err) {
				return
			}
		}
	}
}
