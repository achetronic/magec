package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/api/admin"
	"github.com/achetronic/magec/server/clients"
	"github.com/achetronic/magec/server/store"
)

type listClientsOutput struct {
	Clients []store.ClientDefinition `json:"clients"`
}

func (h *Handler) listClients(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, listClientsOutput, error) {
	return nil, listClientsOutput{Clients: h.store.ListRawClients()}, nil
}

func (h *Handler) getClient(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, store.ClientDefinition, error) {
	c, ok := h.store.GetRawClient(in.ID)
	if !ok {
		return nil, store.ClientDefinition{}, fmt.Errorf("get client: %w", errValidation("not found: "+in.ID))
	}
	return nil, c, nil
}

type createClientInput struct {
	Client store.ClientDefinition `json:"client" jsonschema:"client definition (token is auto-generated)"`
}

func (h *Handler) createClient(_ context.Context, _ *sdk.CallToolRequest, in createClientInput) (*sdk.CallToolResult, store.ClientDefinition, error) {
	c := in.Client
	if c.Name == "" {
		return nil, store.ClientDefinition{}, fmt.Errorf("create client: %w", errValidation("name is required"))
	}
	if c.Type == "" {
		return nil, store.ClientDefinition{}, fmt.Errorf("create client: %w", errValidation("type is required"))
	}
	if !clients.ValidType(c.Type) {
		return nil, store.ClientDefinition{}, fmt.Errorf("create client: %w", errValidation("unsupported client type: "+c.Type))
	}
	if err := admin.ValidateClientConfig(c); err != nil {
		return nil, store.ClientDefinition{}, fmt.Errorf("create client: %w", err)
	}
	created, err := h.store.CreateClient(c)
	if err != nil {
		return nil, store.ClientDefinition{}, fmt.Errorf("create client: %w", err)
	}
	return nil, created, nil
}

type updateClientInput struct {
	ID     string                 `json:"id" jsonschema:"client id"`
	Client store.ClientDefinition `json:"client" jsonschema:"new client definition"`
}

func (h *Handler) updateClient(_ context.Context, _ *sdk.CallToolRequest, in updateClientInput) (*sdk.CallToolResult, store.ClientDefinition, error) {
	if in.Client.Type != "" {
		if err := admin.ValidateClientConfig(in.Client); err != nil {
			return nil, store.ClientDefinition{}, fmt.Errorf("update client: %w", err)
		}
	}
	if err := h.store.UpdateClient(in.ID, in.Client); err != nil {
		return nil, store.ClientDefinition{}, fmt.Errorf("update client: %w", err)
	}
	updated, _ := h.store.GetRawClient(in.ID)
	return nil, updated, nil
}

func (h *Handler) deleteClient(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, emptyOutput, error) {
	if err := h.store.DeleteClient(in.ID); err != nil {
		return nil, emptyOutput{}, fmt.Errorf("delete client: %w", err)
	}
	return nil, okOutput, nil
}

func (h *Handler) regenerateClientToken(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, store.ClientDefinition, error) {
	cl, err := h.store.RegenerateClientToken(in.ID)
	if err != nil {
		return nil, store.ClientDefinition{}, fmt.Errorf("regenerate client token: %w", err)
	}
	return nil, cl, nil
}

type clientTypeInfo struct {
	Type         string         `json:"type"`
	DisplayName  string         `json:"displayName"`
	ConfigSchema clients.Schema `json:"configSchema"`
}

type listClientTypesOutput struct {
	Types []clientTypeInfo `json:"types"`
}

func (h *Handler) listClientTypes(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, listClientTypesOutput, error) {
	var types []clientTypeInfo
	for _, p := range clients.All() {
		types = append(types, clientTypeInfo{
			Type:         p.Type(),
			DisplayName:  p.DisplayName(),
			ConfigSchema: p.ConfigSchema(),
		})
	}
	return nil, listClientTypesOutput{Types: types}, nil
}

func (h *Handler) registerClientTools() {
	destructive := true
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_list_clients", Title: "List clients",
		Description: "List every configured client (telegram, slack, discord, cron, webhook, direct).",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.listClients)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_get_client", Title: "Get client",
		Description: "Return one client by id (including the auth token).",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.getClient)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_create_client", Title: "Create client",
		Description: "Create a new client. Type must be registered (magec_list_client_types) and config must validate against the type's schema.",
	}, h.createClient)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_update_client", Title: "Update client",
		Description: "Replace the client identified by id. Token and id are preserved.",
		Annotations: &sdk.ToolAnnotations{IdempotentHint: true},
	}, h.updateClient)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_delete_client", Title: "Delete client",
		Description: "Delete a client by id, revoking its token.",
		Annotations: &sdk.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, h.deleteClient)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_regenerate_client_token", Title: "Regenerate client token",
		Description: "Generate a fresh auth token for a client, invalidating the previous one.",
	}, h.regenerateClientToken)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_list_client_types", Title: "List client types",
		Description: "List registered client types with their JSON schemas.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.listClientTypes)
	h.toolCount++
}
