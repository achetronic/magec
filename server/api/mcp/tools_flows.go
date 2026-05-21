package mcp

import (
	"context"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/api/admin"
	"github.com/achetronic/magec/server/store"
)

// openObjectSchema is a permissive JSON Schema that accepts any object. We
// pin it on flow tools because store.FlowStep is self-referential (each step
// has Steps []FlowStep), and the SDK's reflection-based schema generator does
// not support cycles. Tool argument validation falls back to "any object" for
// these tools; the tool handler still receives a typed store.FlowDefinition
// via the SDK's JSON unmarshalling, so runtime behaviour is unchanged.
func openObjectSchema() *jsonschema.Schema {
	return &jsonschema.Schema{Type: "object"}
}

type listFlowsOutput struct {
	Flows []store.FlowDefinition `json:"flows"`
}

func (h *Handler) listFlows(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, listFlowsOutput, error) {
	return nil, listFlowsOutput{Flows: h.store.ListRawFlows()}, nil
}

func (h *Handler) getFlow(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, store.FlowDefinition, error) {
	f, ok := h.store.GetRawFlow(in.ID)
	if !ok {
		return nil, store.FlowDefinition{}, fmt.Errorf("get flow: %w", errValidation("not found: "+in.ID))
	}
	return nil, f, nil
}

type createFlowInput struct {
	Flow store.FlowDefinition `json:"flow" jsonschema:"flow definition"`
}

func (h *Handler) createFlow(_ context.Context, _ *sdk.CallToolRequest, in createFlowInput) (*sdk.CallToolResult, store.FlowDefinition, error) {
	if in.Flow.Name == "" {
		return nil, store.FlowDefinition{}, fmt.Errorf("create flow: %w", errValidation("name is required"))
	}
	if err := admin.ValidateFlowStep(&in.Flow.Root); err != nil {
		return nil, store.FlowDefinition{}, fmt.Errorf("create flow: %w", err)
	}
	created, err := h.store.CreateFlow(in.Flow)
	if err != nil {
		return nil, store.FlowDefinition{}, fmt.Errorf("create flow: %w", err)
	}
	return nil, created, nil
}

type updateFlowInput struct {
	ID   string               `json:"id" jsonschema:"flow id"`
	Flow store.FlowDefinition `json:"flow" jsonschema:"new flow definition"`
}

func (h *Handler) updateFlow(_ context.Context, _ *sdk.CallToolRequest, in updateFlowInput) (*sdk.CallToolResult, store.FlowDefinition, error) {
	if err := admin.ValidateFlowStep(&in.Flow.Root); err != nil {
		return nil, store.FlowDefinition{}, fmt.Errorf("update flow: %w", err)
	}
	if err := h.store.UpdateFlow(in.ID, in.Flow); err != nil {
		return nil, store.FlowDefinition{}, fmt.Errorf("update flow: %w", err)
	}
	updated, _ := h.store.GetRawFlow(in.ID)
	return nil, updated, nil
}

func (h *Handler) deleteFlow(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, emptyOutput, error) {
	if err := h.store.DeleteFlow(in.ID); err != nil {
		return nil, emptyOutput{}, fmt.Errorf("delete flow: %w", err)
	}
	return nil, okOutput, nil
}

func (h *Handler) registerFlowTools() {
	destructive := true
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_list_flows", Title: "List flows",
		Description:  "List every multi-agent flow.",
		Annotations:  &sdk.ToolAnnotations{ReadOnlyHint: true},
		OutputSchema: openObjectSchema(),
	}, h.listFlows)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_get_flow", Title: "Get flow",
		Description:  "Return one flow by id.",
		Annotations:  &sdk.ToolAnnotations{ReadOnlyHint: true},
		OutputSchema: openObjectSchema(),
	}, h.getFlow)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_create_flow", Title: "Create flow",
		Description:  "Create a new flow. The Root step tree is validated recursively. Loop steps must set at most one of exitLoop or exitWhen.",
		InputSchema:  openObjectSchema(),
		OutputSchema: openObjectSchema(),
	}, h.createFlow)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_update_flow", Title: "Update flow",
		Description:  "Replace the flow identified by id. The Root step tree is validated recursively.",
		Annotations:  &sdk.ToolAnnotations{IdempotentHint: true},
		InputSchema:  openObjectSchema(),
		OutputSchema: openObjectSchema(),
	}, h.updateFlow)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_delete_flow", Title: "Delete flow",
		Description: "Delete a flow by id.",
		Annotations: &sdk.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, h.deleteFlow)
	h.toolCount++
}
