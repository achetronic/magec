package userapi

import (
	"encoding/json"
	"net/http"

	"github.com/achetronic/magec/server/store"
)

// DeviceInfoResponse is returned when a client is authenticated.
type DeviceInfoResponse struct {
	Paired        bool              `json:"paired" example:"true"`
	Name          string            `json:"name,omitempty" example:"my-tablet"`
	DefaultAgent  string            `json:"defaultAgent,omitempty" example:"magec"`
	AllowedAgents []AgentSummary    `json:"allowedAgents,omitempty"`
}

// DeviceInfoUnpairedResponse is returned when no auth token is provided.
type DeviceInfoUnpairedResponse struct {
	Paired bool `json:"paired" example:"false"`
}

// AgentSummary is a minimal agent descriptor.
type AgentSummary struct {
	ID   string `json:"id" example:"magec"`
	Name string `json:"name" example:"Magec"`
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

// DeviceInfo returns pairing and agent info for the authenticated client.
// @Summary      Device info
// @Description  Returns client pairing status, name, default agent, and allowed agents. Requires Bearer token via Authorization header.
// @Tags         device
// @Produce      json
// @Success      200  {object}  DeviceInfoResponse          "Authenticated client info"
// @Header       200  {string}  X-Client-Name               "Set by auth middleware"
// @Failure      404  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /device/info [get]
func (h *Handler) DeviceInfo(w http.ResponseWriter, r *http.Request) {
	clientName := r.Header.Get("X-Client-Name")
	if clientName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DeviceInfoUnpairedResponse{Paired: false})
		return
	}
	cl, ok := h.store.GetClient(clientName)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "client not found"})
		return
	}
	agents := h.store.ListAgents()
	allowedDetails := make([]AgentSummary, 0, len(cl.AllowedAgents))
	for _, agentID := range cl.AllowedAgents {
		for _, a := range agents {
			if a.ID == agentID {
				allowedDetails = append(allowedDetails, AgentSummary{ID: a.ID, Name: a.Name})
				break
			}
		}
	}
	defaultAgent := ""
	if len(cl.AllowedAgents) > 0 {
		defaultAgent = cl.AllowedAgents[0]
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DeviceInfoResponse{
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
