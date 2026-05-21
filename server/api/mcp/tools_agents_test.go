package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/store"
)

func TestAgentCRUD_LifecycleAndMCPLinks(t *testing.T) {
	h := newTestHandler(t)
	ctx := context.Background()

	_, _, err := h.createAgent(ctx, &sdk.CallToolRequest{}, createAgentInput{Agent: store.AgentDefinition{Name: ""}})
	if err == nil || !IsValidation(err) {
		t.Fatalf("expected validation error for empty name, got %v", err)
	}

	_, agent, err := h.createAgent(ctx, &sdk.CallToolRequest{}, createAgentInput{Agent: store.AgentDefinition{Name: "weather"}})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	if agent.ID == "" {
		t.Fatal("agent id is empty")
	}

	_, mcp, err := h.createMCPServer(ctx, &sdk.CallToolRequest{}, createMCPInput{Server: store.MCPServer{Name: "weather-mcp", Type: "http", Endpoint: "http://example.com"}})
	if err != nil {
		t.Fatalf("create mcp server: %v", err)
	}

	if _, _, err := h.linkAgentMCP(ctx, &sdk.CallToolRequest{}, agentMCPLinkInput{AgentID: agent.ID, MCPID: mcp.ID}); err != nil {
		t.Fatalf("link agent mcp: %v", err)
	}

	_, list, err := h.listAgentMCPs(ctx, &sdk.CallToolRequest{}, idInput{ID: agent.ID})
	if err != nil {
		t.Fatalf("list agent mcps: %v", err)
	}
	if len(list.MCPServers) != 1 || list.MCPServers[0].ID != mcp.ID {
		t.Fatalf("unexpected agent mcps: %+v", list)
	}

	if _, _, err := h.unlinkAgentMCP(ctx, &sdk.CallToolRequest{}, agentMCPLinkInput{AgentID: agent.ID, MCPID: mcp.ID}); err != nil {
		t.Fatalf("unlink agent mcp: %v", err)
	}

	if _, _, err := h.deleteAgent(ctx, &sdk.CallToolRequest{}, idInput{ID: agent.ID}); err != nil {
		t.Fatalf("delete agent: %v", err)
	}
	if got := len(h.store.ListRawAgents()); got != 0 {
		t.Fatalf("expected 0 agents after delete, got %d", got)
	}
}
