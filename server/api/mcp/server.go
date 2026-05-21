package mcp

// registerAll wires every tool group on the MCP server. Each helper
// increments h.toolCount so the startup log can report the catalogue size
// without re-introspecting the SDK.
func (h *Handler) registerAll() {
	h.registerBackendTools()
	h.registerMemoryTools()
	h.registerMCPServerTools()
	h.registerAgentTools()
	h.registerClientTools()
	h.registerCommandTools()
	h.registerFlowTools()
	h.registerSkillTools()
	h.registerSettingsTools()
	h.registerSecretTools()
	h.registerConversationTools()
	h.registerVoiceTools()
}
