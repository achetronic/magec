// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package user

// A2A endpoint documentation stubs.
// These endpoints are served by the A2A handler, not by the user API handlers.
// The annotations here exist solely to include them in the Swagger spec.

// A2AAgentCard describes an A2A-enabled agent for discovery.
type A2AAgentCard struct {
	Name               string           `json:"name" example:"Home Assistant"`
	Description        string           `json:"description" example:"Smart home control agent with access to Home Assistant"`
	URL                string           `json:"url" example:"https://magec.example.com/api/v1/a2a/home-assistant"`
	Version            string           `json:"version" example:"1.0.0"`
	ProtocolVersion    string           `json:"protocolVersion" example:"0.2.5"`
	PreferredTransport string           `json:"preferredTransport" example:"JSONRPC"`
	DefaultInputModes  []string         `json:"defaultInputModes" example:"text/plain"`
	DefaultOutputModes []string         `json:"defaultOutputModes" example:"text/plain"`
	Capabilities       A2ACapabilities  `json:"capabilities"`
	Skills             []A2ASkill       `json:"skills"`
	SecuritySchemes    map[string]any   `json:"securitySchemes"`
	Security           []map[string]any `json:"security"`
}

// A2ACapabilities describes supported A2A protocol capabilities.
type A2ACapabilities struct {
	Streaming bool `json:"streaming" example:"true"`
}

// A2ASkill describes a single capability of an A2A agent.
type A2ASkill struct {
	ID          string   `json:"id" example:"home-assistant"`
	Name        string   `json:"name" example:"model"`
	Description string   `json:"description" example:"I am a helpful smart home assistant. I can control lights and thermostats."`
	Tags        []string `json:"tags" example:"llm"`
}

// A2AJSONRPCRequest is a JSON-RPC 2.0 request for A2A method invocation.
type A2AJSONRPCRequest struct {
	JSONRPC string `json:"jsonrpc" example:"2.0"`
	ID      int    `json:"id" example:"1"`
	Method  string `json:"method" example:"message/send"`
	Params  any    `json:"params"`
}

// A2AJSONRPCResponse is a JSON-RPC 2.0 response from an A2A invocation.
type A2AJSONRPCResponse struct {
	JSONRPC string `json:"jsonrpc" example:"2.0"`
	ID      int    `json:"id" example:"1"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

// A2AListCards returns all A2A-enabled agent cards.
// @Summary      List all A2A agent cards
// @Description  Returns agent cards for all agents that have A2A enabled. Each card describes the agent's identity, capabilities, skills, and security requirements following the A2A protocol specification.
// @Tags         a2a
// @Produce      json
// @Param        agent  query     string  false  "Filter by agent ID. When provided, returns a single card object instead of an array."  example(home-assistant)
// @Success      200    {array}   A2AAgentCard  "List of agent cards (or single card object when ?agent= is used)"
// @Router       /a2a/.well-known/agent-card.json [get]
func (h *Handler) A2AListCards() {}

// A2APerAgentCard returns the agent card for a specific agent.
// @Summary      Get agent card by ID
// @Description  Returns the A2A agent card for the specified agent. This is the per-agent discovery endpoint that A2A clients resolve relative to the agent's base URL.
// @Tags         a2a
// @Produce      json
// @Param        agentID  path      string  true  "Agent ID"  example(home-assistant)
// @Success      200      {object}  A2AAgentCard  "Agent card"
// @Failure      404      {object}  ErrorResponse  "Agent not found or A2A not enabled"
// @Router       /a2a/{agentID}/.well-known/agent-card.json [get]
func (h *Handler) A2APerAgentCard() {}

// A2AInvoke invokes an A2A agent via JSON-RPC 2.0.
// @Summary      Invoke A2A agent (JSON-RPC)
// @Description  Sends a JSON-RPC 2.0 request to the specified agent. Supported methods include `message/send` and `message/stream`. The request must include a valid Bearer token.
// @Tags         a2a
// @Accept       json
// @Produce      json
// @Param        agentID  path      string             true  "Agent ID"  example(home-assistant)
// @Param        body     body      A2AJSONRPCRequest  true  "JSON-RPC 2.0 request"
// @Success      200      {object}  A2AJSONRPCResponse  "JSON-RPC 2.0 response"
// @Failure      401      {object}  ErrorResponse  "Missing or invalid Bearer token"
// @Failure      404      {object}  ErrorResponse  "Agent not found or A2A not enabled"
// @Security     BearerAuth
// @Router       /a2a/{agentID} [post]
func (h *Handler) A2AInvoke() {}
