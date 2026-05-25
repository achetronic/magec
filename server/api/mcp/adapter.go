package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/achetronic/openapi2tools/mcptools"
	"github.com/google/jsonschema-go/jsonschema"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerDescriptors installs every openapi2tools tool descriptor on the
// given go-sdk server. The library ships an adapter for an older go-sdk
// version; we keep our own thin one here so we can stay on the currently
// vendored v1.4.1 without dragging an incompatible transitive dep.
func registerDescriptors(s *sdk.Server, descriptors []mcptools.ToolDescriptor) (int, error) {
	for _, td := range descriptors {
		if err := registerDescriptor(s, td); err != nil {
			return 0, fmt.Errorf("register tool %q: %w", td.Name, err)
		}
	}
	return len(descriptors), nil
}

// registerDescriptor turns a single openapi2tools descriptor into a fully
// wired tool on the go-sdk server.
func registerDescriptor(s *sdk.Server, td mcptools.ToolDescriptor) error {
	schema, err := schemaFromMap(td.InputSchema)
	if err != nil {
		return fmt.Errorf("build input schema: %w", err)
	}
	s.AddTool(&sdk.Tool{
		Name:        td.Name,
		Description: td.Description,
		InputSchema: schema,
		Annotations: annotationsForMethod(td.Route.Method),
	}, makeToolHandler(td))
	return nil
}

// makeToolHandler builds the go-sdk handler closure that bridges the
// JSON-RPC envelope to the library-agnostic openapi2tools.ToolHandler.
// Extracted into a named helper so the registration loop stays compact and
// the dispatch path is easy to read in isolation.
func makeToolHandler(td mcptools.ToolDescriptor) sdk.ToolHandler {
	inner := td.Handler
	return func(ctx context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
		if inner == nil {
			return errorResult("no handler configured for this tool"), nil
		}
		args, err := decodeArguments(req)
		if err != nil {
			return errorResult(fmt.Sprintf("invalid arguments: %s", err)), nil
		}
		result, err := inner(ctx, mcptools.ToolCall{
			Name:      td.Name,
			Arguments: args,
		})
		if err != nil {
			return nil, err
		}
		if result == nil {
			result = &mcptools.ToolResult{}
		}
		return &sdk.CallToolResult{
			Content: []sdk.Content{&sdk.TextContent{Text: result.Text}},
			IsError: result.IsError,
		}, nil
	}
}

// schemaFromMap roundtrips a JSON-schema-compatible map through JSON into
// the go-sdk's typed jsonschema.Schema. AddTool requires a non-nil schema;
// an empty object accepts anything, which is the right default for tools
// that take no input.
func schemaFromMap(m map[string]any) (*jsonschema.Schema, error) {
	if m == nil {
		return &jsonschema.Schema{Type: "object"}, nil
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshal schema map: %w", err)
	}
	var schema jsonschema.Schema
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, fmt.Errorf("unmarshal schema: %w", err)
	}
	return &schema, nil
}

// decodeArguments unmarshals the JSON-RPC arguments envelope into a
// map[string]any. The go-sdk represents Arguments as json.RawMessage so
// each tool can decode lazily and avoid a per-tool input type.
func decodeArguments(req *sdk.CallToolRequest) (map[string]any, error) {
	if req == nil || len(req.Params.Arguments) == 0 {
		return map[string]any{}, nil
	}
	var args map[string]any
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return nil, fmt.Errorf("decode arguments: %w", err)
	}
	if args == nil {
		args = map[string]any{}
	}
	return args, nil
}

// annotationsForMethod derives the MCP tool hints from the HTTP verb. The
// admin REST API is conventional enough — GET reads, DELETE removes, PUT
// updates — that this mapping is accurate without per-route overrides.
func annotationsForMethod(method string) *sdk.ToolAnnotations {
	switch method {
	case http.MethodGet:
		return &sdk.ToolAnnotations{ReadOnlyHint: true}
	case http.MethodPut:
		return &sdk.ToolAnnotations{IdempotentHint: true}
	case http.MethodDelete:
		destructive := true
		return &sdk.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true}
	}
	return nil
}

// errorResult returns a CallToolResult flagged as a tool-level error with
// the given text. Used by the dispatcher when the call cannot reach the
// inner handler (bad arguments, missing handler) — distinct from a JSON-RPC
// protocol error, which is returned as the second value instead.
func errorResult(text string) *sdk.CallToolResult {
	return &sdk.CallToolResult{
		Content: []sdk.Content{&sdk.TextContent{Text: text}},
		IsError: true,
	}
}
