package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/store"
)

type listAgentsOutput struct {
	Agents []store.AgentDefinition `json:"agents"`
}

func (h *Handler) listAgents(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, listAgentsOutput, error) {
	return nil, listAgentsOutput{Agents: h.store.ListRawAgents()}, nil
}

func (h *Handler) getAgent(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, store.AgentDefinition, error) {
	a, ok := h.store.GetRawAgent(in.ID)
	if !ok {
		return nil, store.AgentDefinition{}, fmt.Errorf("get agent: %w", errValidation("not found: "+in.ID))
	}
	return nil, a, nil
}

type createAgentInput struct {
	Agent store.AgentDefinition `json:"agent" jsonschema:"agent definition"`
}

func (h *Handler) createAgent(_ context.Context, _ *sdk.CallToolRequest, in createAgentInput) (*sdk.CallToolResult, store.AgentDefinition, error) {
	if in.Agent.Name == "" {
		return nil, store.AgentDefinition{}, fmt.Errorf("create agent: %w", errValidation("name is required"))
	}
	created, err := h.store.CreateAgent(in.Agent)
	if err != nil {
		return nil, store.AgentDefinition{}, fmt.Errorf("create agent: %w", err)
	}
	return nil, created, nil
}

type updateAgentInput struct {
	ID    string                `json:"id" jsonschema:"agent id"`
	Agent store.AgentDefinition `json:"agent" jsonschema:"new agent definition"`
}

func (h *Handler) updateAgent(_ context.Context, _ *sdk.CallToolRequest, in updateAgentInput) (*sdk.CallToolResult, store.AgentDefinition, error) {
	if err := h.store.UpdateAgent(in.ID, in.Agent); err != nil {
		return nil, store.AgentDefinition{}, fmt.Errorf("update agent: %w", err)
	}
	updated, _ := h.store.GetRawAgent(in.ID)
	return nil, updated, nil
}

func (h *Handler) deleteAgent(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, emptyOutput, error) {
	if err := h.store.DeleteAgent(in.ID); err != nil {
		return nil, emptyOutput{}, fmt.Errorf("delete agent: %w", err)
	}
	return nil, okOutput, nil
}

type listAgentMCPsOutput struct {
	MCPServers []store.MCPServer `json:"mcpServers"`
}

func (h *Handler) listAgentMCPs(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, listAgentMCPsOutput, error) {
	mcps, err := h.store.ResolveRawAgentMCPs(in.ID)
	if err != nil {
		return nil, listAgentMCPsOutput{}, fmt.Errorf("list agent mcps: %w", err)
	}
	return nil, listAgentMCPsOutput{MCPServers: mcps}, nil
}

func (h *Handler) linkAgentMCP(_ context.Context, _ *sdk.CallToolRequest, in agentMCPLinkInput) (*sdk.CallToolResult, emptyOutput, error) {
	if err := h.store.LinkAgentMCP(in.AgentID, in.MCPID); err != nil {
		return nil, emptyOutput{}, fmt.Errorf("link agent mcp: %w", err)
	}
	return nil, okOutput, nil
}

func (h *Handler) unlinkAgentMCP(_ context.Context, _ *sdk.CallToolRequest, in agentMCPLinkInput) (*sdk.CallToolResult, emptyOutput, error) {
	if err := h.store.UnlinkAgentMCP(in.AgentID, in.MCPID); err != nil {
		return nil, emptyOutput{}, fmt.Errorf("unlink agent mcp: %w", err)
	}
	return nil, okOutput, nil
}

func (h *Handler) registerAgentTools() {
	destructive := true
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_list_agents", Title: "List agents",
		Description: "List every agent definition.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.listAgents)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_get_agent", Title: "Get agent",
		Description: "Return one agent by id.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.getAgent)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_create_agent", Title: "Create agent",
		Description: "Create a new agent.",
	}, h.createAgent)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_update_agent", Title: "Update agent",
		Description: "Replace the agent identified by id.",
		Annotations: &sdk.ToolAnnotations{IdempotentHint: true},
	}, h.updateAgent)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_delete_agent", Title: "Delete agent",
		Description: "Delete an agent by id.",
		Annotations: &sdk.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, h.deleteAgent)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_list_agent_mcps", Title: "List agent MCP servers",
		Description: "Resolve the MCP servers linked to an agent.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.listAgentMCPs)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_link_agent_mcp", Title: "Link MCP to agent",
		Description: "Attach an MCP server to an agent.",
		Annotations: &sdk.ToolAnnotations{IdempotentHint: true},
	}, h.linkAgentMCP)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_unlink_agent_mcp", Title: "Unlink MCP from agent",
		Description: "Detach an MCP server from an agent.",
		Annotations: &sdk.ToolAnnotations{IdempotentHint: true},
	}, h.unlinkAgentMCP)
	h.toolCount++
}
