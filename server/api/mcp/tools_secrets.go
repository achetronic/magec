package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/api/admin"
	"github.com/achetronic/magec/server/store"
)

type listSecretsOutput struct {
	Secrets []admin.SecretResponse `json:"secrets"`
}

func (h *Handler) listSecrets(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, listSecretsOutput, error) {
	stored := h.store.ListSecrets()
	out := make([]admin.SecretResponse, len(stored))
	for i, s := range stored {
		out[i] = admin.SecretToResponse(s)
	}
	return nil, listSecretsOutput{Secrets: out}, nil
}

func (h *Handler) getSecret(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, admin.SecretResponse, error) {
	s, ok := h.store.GetSecret(in.ID)
	if !ok {
		return nil, admin.SecretResponse{}, fmt.Errorf("get secret: %w", errValidation("not found: "+in.ID))
	}
	return nil, admin.SecretToResponse(s), nil
}

type createSecretInput struct {
	Name        string `json:"name" jsonschema:"human-readable secret name"`
	Key         string `json:"key" jsonschema:"environment variable name (e.g. OPENAI_API_KEY)"`
	Value       string `json:"value" jsonschema:"secret value (never returned in subsequent reads)"`
	Description string `json:"description,omitempty"`
}

func (h *Handler) createSecret(_ context.Context, _ *sdk.CallToolRequest, in createSecretInput) (*sdk.CallToolResult, admin.SecretResponse, error) {
	if in.Name == "" {
		return nil, admin.SecretResponse{}, fmt.Errorf("create secret: %w", errValidation("name is required"))
	}
	if in.Key == "" {
		return nil, admin.SecretResponse{}, fmt.Errorf("create secret: %w", errValidation("key is required"))
	}
	if in.Value == "" {
		return nil, admin.SecretResponse{}, fmt.Errorf("create secret: %w", errValidation("value is required"))
	}
	s, err := h.store.CreateSecret(store.Secret{
		Name:        in.Name,
		Key:         in.Key,
		Value:       in.Value,
		Description: in.Description,
	})
	if err != nil {
		return nil, admin.SecretResponse{}, fmt.Errorf("create secret: %w", err)
	}
	return nil, admin.SecretToResponse(s), nil
}

type updateSecretInput struct {
	ID          string `json:"id" jsonschema:"secret id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Value       string `json:"value,omitempty" jsonschema:"new value (omit to keep the existing one)"`
	Description string `json:"description,omitempty"`
}

func (h *Handler) updateSecret(_ context.Context, _ *sdk.CallToolRequest, in updateSecretInput) (*sdk.CallToolResult, admin.SecretResponse, error) {
	if err := h.store.UpdateSecret(in.ID, store.Secret{
		Name:        in.Name,
		Key:         in.Key,
		Value:       in.Value,
		Description: in.Description,
	}); err != nil {
		return nil, admin.SecretResponse{}, fmt.Errorf("update secret: %w", err)
	}
	updated, _ := h.store.GetSecret(in.ID)
	return nil, admin.SecretToResponse(updated), nil
}

func (h *Handler) deleteSecret(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, emptyOutput, error) {
	if err := h.store.DeleteSecret(in.ID); err != nil {
		return nil, emptyOutput{}, fmt.Errorf("delete secret: %w", err)
	}
	return nil, okOutput, nil
}

func (h *Handler) registerSecretTools() {
	destructive := true
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_list_secrets", Title: "List secrets",
		Description: "List every secret (id, name, key, description). Values are never returned.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.listSecrets)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_get_secret", Title: "Get secret",
		Description: "Return one secret by id (without the value).",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.getSecret)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_create_secret", Title: "Create secret",
		Description: "Create a new secret. Value is stored encrypted at rest if server.encryptionKey is configured.",
	}, h.createSecret)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_update_secret", Title: "Update secret",
		Description: "Replace a secret by id. Leave value empty to keep the existing one (other fields are still rewritten).",
		Annotations: &sdk.ToolAnnotations{IdempotentHint: true},
	}, h.updateSecret)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_delete_secret", Title: "Delete secret",
		Description: "Delete a secret by id. Unsets the corresponding env var.",
		Annotations: &sdk.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, h.deleteSecret)
	h.toolCount++
}
