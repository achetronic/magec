package mcp

// idInput is the canonical shape for tools that target a resource by ID.
type idInput struct {
	ID string `json:"id" jsonschema:"resource id"`
}

// agentMCPLinkInput targets the agent/{id}/mcps/{mcpId} link operations.
type agentMCPLinkInput struct {
	AgentID string `json:"agentId" jsonschema:"agent id"`
	MCPID   string `json:"mcpId" jsonschema:"mcp server id"`
}

// idsOutput is reused by tools that return a list of IDs.
type idsOutput struct {
	IDs []string `json:"ids"`
}

// emptyOutput is returned by tools that have no payload (delete, link, etc.).
type emptyOutput struct {
	OK bool `json:"ok"`
}

var okOutput = emptyOutput{OK: true}
