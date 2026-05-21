package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/store"
)

type listMCPServersOutput struct {
	Servers []store.MCPServer `json:"servers"`
}

func (h *Handler) listMCPServers(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, listMCPServersOutput, error) {
	return nil, listMCPServersOutput{Servers: h.store.ListRawMCPServers()}, nil
}

func (h *Handler) getMCPServer(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, store.MCPServer, error) {
	m, ok := h.store.GetRawMCPServer(in.ID)
	if !ok {
		return nil, store.MCPServer{}, fmt.Errorf("get mcp server: %w", errValidation("not found: "+in.ID))
	}
	return nil, m, nil
}

type createMCPInput struct {
	Server store.MCPServer `json:"server" jsonschema:"mcp server definition"`
}

func (h *Handler) createMCPServer(_ context.Context, _ *sdk.CallToolRequest, in createMCPInput) (*sdk.CallToolResult, store.MCPServer, error) {
	if in.Server.Name == "" {
		return nil, store.MCPServer{}, fmt.Errorf("create mcp server: %w", errValidation("name is required"))
	}
	created, err := h.store.CreateMCPServer(in.Server)
	if err != nil {
		return nil, store.MCPServer{}, fmt.Errorf("create mcp server: %w", err)
	}
	return nil, created, nil
}

type updateMCPInput struct {
	ID     string          `json:"id" jsonschema:"mcp server id"`
	Server store.MCPServer `json:"server" jsonschema:"new mcp server definition"`
}

func (h *Handler) updateMCPServer(_ context.Context, _ *sdk.CallToolRequest, in updateMCPInput) (*sdk.CallToolResult, store.MCPServer, error) {
	if err := h.store.UpdateMCPServer(in.ID, in.Server); err != nil {
		return nil, store.MCPServer{}, fmt.Errorf("update mcp server: %w", err)
	}
	updated, _ := h.store.GetRawMCPServer(in.ID)
	return nil, updated, nil
}

func (h *Handler) deleteMCPServer(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, emptyOutput, error) {
	if err := h.store.DeleteMCPServer(in.ID); err != nil {
		return nil, emptyOutput{}, fmt.Errorf("delete mcp server: %w", err)
	}
	return nil, okOutput, nil
}

func (h *Handler) registerMCPServerTools() {
	destructive := true
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_list_mcp_servers", Title: "List MCP servers",
		Description: "List every MCP server registered as a tool source. Not to be confused with the embedded MCP server you are currently talking to.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.listMCPServers)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_get_mcp_server", Title: "Get MCP server",
		Description: "Return one registered MCP server by id.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.getMCPServer)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_create_mcp_server", Title: "Create MCP server",
		Description: "Register an external MCP server. Use type=http with endpoint+headers, or type=stdio with command+args+env.",
	}, h.createMCPServer)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_update_mcp_server", Title: "Update MCP server",
		Description: "Replace the MCP server identified by id.",
		Annotations: &sdk.ToolAnnotations{IdempotentHint: true},
	}, h.updateMCPServer)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_delete_mcp_server", Title: "Delete MCP server",
		Description: "Delete a registered MCP server by id.",
		Annotations: &sdk.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, h.deleteMCPServer)
	h.toolCount++
}
