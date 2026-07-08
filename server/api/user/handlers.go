// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"encoding/json"
	"net/http"

	"github.com/achetronic/magec/server/store"
)

// ClientInfoResponse is returned when a client is authenticated.
type ClientInfoResponse struct {
	Paired        bool           `json:"paired" example:"true"`
	Name          string         `json:"name,omitempty" example:"my-tablet"`
	DefaultAgent  string         `json:"defaultAgent,omitempty" example:"magec"`
	AllowedAgents []AgentSummary `json:"allowedAgents,omitempty"`
}

// ClientInfoUnpairedResponse is returned when no auth token is provided.
type ClientInfoUnpairedResponse struct {
	Paired bool `json:"paired" example:"false"`
}

// AgentSummary is a minimal agent descriptor.
type AgentSummary struct {
	ID            string         `json:"id" example:"magec"`
	Name          string         `json:"name" example:"Magec"`
	Type          string         `json:"type" example:"agent"`
	ResponseAgent bool           `json:"responseAgent,omitempty"`
	Agents        []AgentSummary `json:"agents,omitempty"`
}

// HealthResponse is the response from the health endpoint.
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// SpeechRequest is the body for the TTS speech proxy.
type SpeechRequest struct {
	Input string `json:"input" example:"Hello world"`
}

// ErrorResponse is returned for all error responses.
type ErrorResponse struct {
	Error string `json:"error" example:"resource not found"`
}

// Handler provides the user API endpoints.
type Handler struct {
	store *store.Store
}

// New creates a new user API handler.
func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

// Health checks if the server is running.
// @Summary      Health check
// @Description  Returns 200 if the server is healthy
// @Tags         system
// @Produce      plain
// @Success      200  {string}  string  "ok"
// @Router       /health [get]
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// ClientInfo returns pairing and agent info for the authenticated client.
// @Summary      Client info
// @Description  Returns client pairing status, name, default agent, and allowed agents. Requires Bearer token via Authorization header.
// @Tags         client
// @Produce      json
// @Success      200  {object}  ClientInfoResponse          "Authenticated client info"
// @Header       200  {string}  X-Client-ID                 "Set by auth middleware"
// @Failure      404  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /client/info [get]
func (h *Handler) ClientInfo(w http.ResponseWriter, r *http.Request) {
	clientID := r.Header.Get("X-Client-ID")
	if clientID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ClientInfoUnpairedResponse{Paired: false})
		return
	}
	cl, ok := h.store.GetClient(clientID)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "client not found"})
		return
	}
	agents := h.store.ListAgents()
	flows := h.store.ListFlows()
	allowedDetails := make([]AgentSummary, 0, len(cl.AllowedAgents))
	agentMap := make(map[string]store.AgentDefinition, len(agents))
	for _, a := range agents {
		agentMap[a.ID] = a
	}
	for _, id := range cl.AllowedAgents {
		if a, ok := agentMap[id]; ok {
			allowedDetails = append(allowedDetails, AgentSummary{ID: a.ID, Name: a.Name, Type: "agent"})
			continue
		}
		for _, f := range flows {
			if f.ID == id {
				responseSet := make(map[string]bool)
				for _, rid := range f.ResponseAgentIDs() {
					responseSet[rid] = true
				}
				var nested []AgentSummary
				for _, aid := range f.AgentIDs() {
					if na, ok := agentMap[aid]; ok {
						nested = append(nested, AgentSummary{ID: na.ID, Name: na.Name, Type: "agent", ResponseAgent: responseSet[aid]})
					}
				}
				allowedDetails = append(allowedDetails, AgentSummary{ID: f.ID, Name: f.Name, Type: "flow", Agents: nested})
				break
			}
		}
	}
	defaultAgent := ""
	if len(cl.AllowedAgents) > 0 {
		defaultAgent = cl.AllowedAgents[0]
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ClientInfoResponse{
		Paired:        true,
		Name:          cl.Name,
		DefaultAgent:  defaultAgent,
		AllowedAgents: allowedDetails,
	})
}

// Speech proxies a TTS request to the agent's configured backend.
// @Summary      Text-to-Speech
// @Description  Proxies a TTS request to the speech backend configured for the given agent. Returns audio data.
// @Tags         voice
// @Accept       json
// @Produce      application/octet-stream
// @Param        agentId  path      string         true  "Agent ID"
// @Param        body     body      SpeechRequest  true  "Speech request with input text"
// @Success      200      {file}    binary         "Audio data"
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      502      {object}  ErrorResponse
// @Failure      503      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /voice/{agentId}/speech [post]
func (h *Handler) Speech(w http.ResponseWriter, r *http.Request) {}

// Transcription proxies an STT request to the agent's configured backend.
// @Summary      Speech-to-Text
// @Description  Proxies a transcription request to the STT backend configured for the given agent. Accepts multipart audio.
// @Tags         voice
// @Accept       multipart/form-data
// @Produce      json
// @Param        agentId  path      string  true   "Agent ID"
// @Param        file     formData  file    true   "Audio file to transcribe"
// @Success      200      {object}  object  "Transcription result"
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      502      {object}  ErrorResponse
// @Failure      503      {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /voice/{agentId}/transcription [post]
func (h *Handler) Transcription(w http.ResponseWriter, r *http.Request) {}

// VoiceEvents serves the WebSocket connection for real-time voice events (wake word detection, VAD).
// @Summary      Voice events WebSocket
// @Description  WebSocket endpoint for real-time voice events including wake word detection and voice activity detection (VAD).
// @Tags         voice
// @Success      101  {string}  string  "Switching Protocols"
// @Router       /voice/events [get]
func (h *Handler) VoiceEvents(w http.ResponseWriter, r *http.Request) {}

// WebhookRequest is the JSON body for incoming webhook calls.
type WebhookRequest struct {
	Prompt string `json:"prompt,omitempty" example:"Summarize today's news"`
}

// WebhookResponse is the response from a webhook execution.
type WebhookResponse struct {
	OK       bool   `json:"ok" example:"true"`
	Response string `json:"response,omitempty" example:"Here is the summary..."`
	Error    string `json:"error,omitempty" example:""`
}

// Webhook executes a webhook client's configured command against its allowed agents.
// @Summary      Execute webhook
// @Description  Fires a webhook client. For fixed-command webhooks, the command is executed as configured. For passthrough webhooks, the prompt must be provided in the request body. The client's token is required for authentication. The command runs against all agents in the client's allowedAgents list.
// @Tags         webhooks
// @Accept       json
// @Produce      json
// @Param        id    path      string          true  "Webhook client ID"
// @Param        body  body      WebhookRequest  false "Request body (required for passthrough webhooks)"
// @Success      200   {object}  WebhookResponse
// @Failure      400   {object}  WebhookResponse
// @Failure      401   {object}  WebhookResponse
// @Failure      403   {object}  WebhookResponse
// @Failure      404   {object}  WebhookResponse
// @Failure      500   {object}  WebhookResponse
// @Security     BearerAuth
// @Router       /webhooks/{id} [post]
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {}

// EphemeralArtifact serves an artifact's raw bytes when called with a valid
// signed token issued by the get_artifact_url tool.
// @Summary      Download an artifact via signed URL
// @Description  Streams the raw bytes of a session artifact when the supplied token is a non-expired HMAC-SHA256 signature minted by the `get_artifact_url` tool. The token is the only credential: this endpoint bypasses client Bearer authentication. Binary artifacts are served with their original MIME type (defaulting to application/octet-stream); text artifacts come back as text/plain. Both responses set Content-Disposition so curl/wget pick a sensible local filename. The endpoint returns 503 when the server has no signing secret configured (server.encryptionKey unset), 401 for invalid or expired tokens, 404 when the artifact does not exist, and 410 when the artifact exists but has empty content.
// @Tags         ephemeral
// @Produce      application/octet-stream
// @Produce      text/plain
// @Param        token  path      string  true  "Signed token issued by get_artifact_url"
// @Success      200    {file}    binary  "Artifact bytes"
// @Failure      401    {object}  ErrorResponse
// @Failure      404    {object}  ErrorResponse
// @Failure      410    {object}  ErrorResponse
// @Failure      503    {object}  ErrorResponse
// @Router       /ephemeral/artifacts/{token} [get]
func (h *Handler) EphemeralArtifact(w http.ResponseWriter, r *http.Request) {}
