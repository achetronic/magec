package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/store"
)

type listBackendsOutput struct {
	Backends []store.BackendDefinition `json:"backends"`
}

func (h *Handler) listBackends(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, listBackendsOutput, error) {
	return nil, listBackendsOutput{Backends: h.store.ListRawBackends()}, nil
}

func (h *Handler) getBackend(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, store.BackendDefinition, error) {
	b, ok := h.store.GetRawBackend(in.ID)
	if !ok {
		return nil, store.BackendDefinition{}, fmt.Errorf("get backend: %w", errValidation("backend not found: "+in.ID))
	}
	return nil, b, nil
}

type createBackendInput struct {
	Definition store.BackendDefinition `json:"definition" jsonschema:"backend definition (id is assigned by the server)"`
}

func (h *Handler) createBackend(_ context.Context, _ *sdk.CallToolRequest, in createBackendInput) (*sdk.CallToolResult, store.BackendDefinition, error) {
	if in.Definition.Name == "" {
		return nil, store.BackendDefinition{}, fmt.Errorf("create backend: %w", errValidation("name is required"))
	}
	if in.Definition.Type == "" {
		return nil, store.BackendDefinition{}, fmt.Errorf("create backend: %w", errValidation("type is required"))
	}
	created, err := h.store.CreateBackend(in.Definition)
	if err != nil {
		return nil, store.BackendDefinition{}, fmt.Errorf("create backend: %w", err)
	}
	return nil, created, nil
}

type updateBackendInput struct {
	ID         string                  `json:"id" jsonschema:"backend id"`
	Definition store.BackendDefinition `json:"definition" jsonschema:"new backend definition (id is taken from the path)"`
}

func (h *Handler) updateBackend(_ context.Context, _ *sdk.CallToolRequest, in updateBackendInput) (*sdk.CallToolResult, store.BackendDefinition, error) {
	if err := h.store.UpdateBackend(in.ID, in.Definition); err != nil {
		return nil, store.BackendDefinition{}, fmt.Errorf("update backend: %w", err)
	}
	updated, _ := h.store.GetRawBackend(in.ID)
	return nil, updated, nil
}

func (h *Handler) deleteBackend(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, emptyOutput, error) {
	if err := h.store.DeleteBackend(in.ID); err != nil {
		return nil, emptyOutput{}, fmt.Errorf("delete backend: %w", err)
	}
	return nil, okOutput, nil
}

func (h *Handler) registerBackendTools() {
	destructive := true
	sdk.AddTool(h.server, &sdk.Tool{
		Name:        "magec_list_backends",
		Title:       "List backends",
		Description: "List every configured LLM/TTS/transcription backend.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.listBackends)
	h.toolCount++

	sdk.AddTool(h.server, &sdk.Tool{
		Name:        "magec_get_backend",
		Title:       "Get backend",
		Description: "Return one backend by id.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.getBackend)
	h.toolCount++

	sdk.AddTool(h.server, &sdk.Tool{
		Name:        "magec_create_backend",
		Title:       "Create backend",
		Description: "Create a new backend. The server assigns the id.",
	}, h.createBackend)
	h.toolCount++

	sdk.AddTool(h.server, &sdk.Tool{
		Name:        "magec_update_backend",
		Title:       "Update backend",
		Description: "Replace the backend identified by id with the provided definition.",
		Annotations: &sdk.ToolAnnotations{IdempotentHint: true},
	}, h.updateBackend)
	h.toolCount++

	sdk.AddTool(h.server, &sdk.Tool{
		Name:        "magec_delete_backend",
		Title:       "Delete backend",
		Description: "Delete a backend by id.",
		Annotations: &sdk.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, h.deleteBackend)
	h.toolCount++
}
