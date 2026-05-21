package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/store"
)

func (h *Handler) getSettings(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, store.Settings, error) {
	return nil, h.store.GetSettings(), nil
}

type updateSettingsInput struct {
	Settings store.Settings `json:"settings" jsonschema:"new global settings"`
}

func (h *Handler) updateSettings(_ context.Context, _ *sdk.CallToolRequest, in updateSettingsInput) (*sdk.CallToolResult, store.Settings, error) {
	if err := h.store.UpdateSettings(in.Settings); err != nil {
		return nil, store.Settings{}, fmt.Errorf("update settings: %w", err)
	}
	return nil, h.store.GetSettings(), nil
}

func (h *Handler) registerSettingsTools() {
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_get_settings", Title: "Get settings",
		Description: "Return the global runtime settings (session/long-term memory provider selection, temporary dir).",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.getSettings)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_update_settings", Title: "Update settings",
		Description: "Replace the global runtime settings.",
		Annotations: &sdk.ToolAnnotations{IdempotentHint: true},
	}, h.updateSettings)
	h.toolCount++
}
