package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/achetronic/openapi2tools/mcptools"
	"github.com/google/jsonschema-go/jsonschema"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerDescriptors installs every openapi2tools tool descriptor on the
// given go-sdk server. The library ships an adapter for an older go-sdk
// version; we keep our own thin one here so we can stay on the currently
// vendored v1.4.1 without dragging an incompatible transitive dep.
func registerDescriptors(s *sdk.Server, descriptors []mcptools.ToolDescriptor) (int, error) {
	count := 0
	for i := range descriptors {
		if err := registerDescriptor(s, descriptors[i]); err != nil {
			return count, fmt.Errorf("register tool %q: %w", descriptors[i].Name, err)
		}
		count++
	}
	return count, nil
}

func registerDescriptor(s *sdk.Server, td mcptools.ToolDescriptor) error {
	schema, err := schemaFromMap(td.InputSchema)
	if err != nil {
		return fmt.Errorf("build input schema: %w", err)
	}

	tool := &sdk.Tool{
		Name:        td.Name,
		Description: td.Description,
		InputSchema: schema,
		Annotations: annotationsForMethod(td.Route.Method),
	}

	handler := td.Handler
	s.AddTool(tool, func(ctx context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, error) {
		args, err := decodeArguments(req)
		if err != nil {
			return errorResult(fmt.Sprintf("invalid arguments: %s", err)), nil
		}
		if handler == nil {
			return errorResult("no handler configured for this tool"), nil
		}

		out, err := handler(ctx, mcptools.ToolCall{
			Name:      td.Name,
			Arguments: args,
		})
		if err != nil {
			return nil, err
		}
		if out == nil {
			out = &mcptools.ToolResult{}
		}
		return &sdk.CallToolResult{
			Content: []sdk.Content{&sdk.TextContent{Text: out.Text}},
			IsError: out.IsError,
		}, nil
	})
	return nil
}

// schemaFromMap roundtrips a JSON-schema-compatible map through JSON into the
// go-sdk's typed jsonschema.Schema. AddTool requires a non-nil schema; an
// empty object accepts anything which is fine for tools without parameters.
func schemaFromMap(m map[string]any) (*jsonschema.Schema, error) {
	if m == nil {
		return &jsonschema.Schema{Type: "object"}, nil
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var schema jsonschema.Schema
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, err
	}
	return &schema, nil
}

// decodeArguments unmarshals the JSON-RPC arguments envelope into a
// map[string]any. The go-sdk represents Arguments as json.RawMessage so we
// can decode lazily and avoid a per-tool input type.
func decodeArguments(req *sdk.CallToolRequest) (map[string]any, error) {
	if req == nil || len(req.Params.Arguments) == 0 {
		return map[string]any{}, nil
	}
	var args map[string]any
	if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
		return nil, err
	}
	if args == nil {
		args = map[string]any{}
	}
	return args, nil
}

// annotationsForMethod derives the MCP tool hints from the HTTP verb. The
// admin REST API is conventional enough (GET reads, DELETE removes, PUT
// updates) that this mapping is accurate without per-route overrides.
func annotationsForMethod(method string) *sdk.ToolAnnotations {
	destructive := true
	switch method {
	case "GET":
		return &sdk.ToolAnnotations{ReadOnlyHint: true}
	case "DELETE":
		return &sdk.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true}
	case "PUT":
		return &sdk.ToolAnnotations{IdempotentHint: true}
	}
	return nil
}

func errorResult(text string) *sdk.CallToolResult {
	return &sdk.CallToolResult{
		Content: []sdk.Content{&sdk.TextContent{Text: text}},
		IsError: true,
	}
}
