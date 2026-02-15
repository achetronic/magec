package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/adk/session"

	"github.com/achetronic/magec/server/store"
)

// ErrorResponse is returned for all error responses.
type ErrorResponse struct {
	Error string `json:"error" example:"resource not found"`
}

// Handler provides the admin API router.
type Handler struct {
	store          *store.Store
	conversations  *store.ConversationStore
	sessionService session.Service
	router         *mux.Router
}

// New creates a new admin API handler.
func New(s *store.Store) *Handler {
	h := &Handler{store: s}
	h.router = h.buildRouter()
	return h
}

// SetConversationStore injects the conversation store for the audit endpoints.
func (h *Handler) SetConversationStore(cs *store.ConversationStore) {
	h.conversations = cs
}

// ConversationStore returns the conversation store (used by external components
// that need to log conversations).
func (h *Handler) ConversationStore() *store.ConversationStore {
	return h.conversations
}

// SetSessionService injects the ADK session service for direct session operations.
func (h *Handler) SetSessionService(svc session.Service) {
	h.sessionService = svc
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *Handler) buildRouter() *mux.Router {
	r := mux.NewRouter()

	// Backends
	r.HandleFunc("/backends", h.listBackends).Methods("GET")
	r.HandleFunc("/backends", h.createBackend).Methods("POST")
	r.HandleFunc("/backends/{id}", h.getBackend).Methods("GET")
	r.HandleFunc("/backends/{id}", h.updateBackend).Methods("PUT")
	r.HandleFunc("/backends/{id}", h.deleteBackend).Methods("DELETE")

	// Memory Providers
	r.HandleFunc("/memory", h.listMemoryProviders).Methods("GET")
	r.HandleFunc("/memory", h.createMemoryProvider).Methods("POST")
	r.HandleFunc("/memory/types", h.listMemoryTypes).Methods("GET")
	r.HandleFunc("/memory/{id}", h.getMemoryProvider).Methods("GET")
	r.HandleFunc("/memory/{id}", h.updateMemoryProvider).Methods("PUT")
	r.HandleFunc("/memory/{id}", h.deleteMemoryProvider).Methods("DELETE")
	r.HandleFunc("/memory/{id}/health", h.checkMemoryProviderHealth).Methods("GET")

	// MCP Servers (global)
	r.HandleFunc("/mcps", h.listMCPServers).Methods("GET")
	r.HandleFunc("/mcps", h.createMCPServer).Methods("POST")
	r.HandleFunc("/mcps/{id}", h.getMCPServer).Methods("GET")
	r.HandleFunc("/mcps/{id}", h.updateMCPServer).Methods("PUT")
	r.HandleFunc("/mcps/{id}", h.deleteMCPServer).Methods("DELETE")

	// Agents
	r.HandleFunc("/agents", h.listAgents).Methods("GET")
	r.HandleFunc("/agents", h.createAgent).Methods("POST")
	r.HandleFunc("/agents/{id}", h.getAgent).Methods("GET")
	r.HandleFunc("/agents/{id}", h.updateAgent).Methods("PUT")
	r.HandleFunc("/agents/{id}", h.deleteAgent).Methods("DELETE")

	// Agent MCP linking
	r.HandleFunc("/agents/{id}/mcps", h.listAgentMCPs).Methods("GET")
	r.HandleFunc("/agents/{id}/mcps/{mcpId}", h.linkAgentMCP).Methods("PUT")
	r.HandleFunc("/agents/{id}/mcps/{mcpId}", h.unlinkAgentMCP).Methods("DELETE")

	// Clients
	r.HandleFunc("/clients", h.listClients).Methods("GET")
	r.HandleFunc("/clients", h.createClient).Methods("POST")
	r.HandleFunc("/clients/types", h.listClientTypes).Methods("GET")
	r.HandleFunc("/clients/{id}", h.getClient).Methods("GET")
	r.HandleFunc("/clients/{id}", h.updateClient).Methods("PUT")
	r.HandleFunc("/clients/{id}", h.deleteClient).Methods("DELETE")
	r.HandleFunc("/clients/{id}/regenerate-token", h.regenerateClientToken).Methods("POST")

	// Commands
	r.HandleFunc("/commands", h.listCommands).Methods("GET")
	r.HandleFunc("/commands", h.createCommand).Methods("POST")
	r.HandleFunc("/commands/{id}", h.getCommand).Methods("GET")
	r.HandleFunc("/commands/{id}", h.updateCommand).Methods("PUT")
	r.HandleFunc("/commands/{id}", h.deleteCommand).Methods("DELETE")

	// Flows
	r.HandleFunc("/flows", h.listFlows).Methods("GET")
	r.HandleFunc("/flows", h.createFlow).Methods("POST")
	r.HandleFunc("/flows/{id}", h.getFlow).Methods("GET")
	r.HandleFunc("/flows/{id}", h.updateFlow).Methods("PUT")
	r.HandleFunc("/flows/{id}", h.deleteFlow).Methods("DELETE")

	// Settings
	r.HandleFunc("/settings", h.getSettings).Methods("GET")
	r.HandleFunc("/settings", h.updateSettings).Methods("PUT")

	// Conversations (audit)
	r.HandleFunc("/conversations", h.listConversations).Methods("GET")
	r.HandleFunc("/conversations/stats", h.conversationStats).Methods("GET")
	r.HandleFunc("/conversations/clear", h.clearConversations).Methods("DELETE")
	r.HandleFunc("/conversations/{id}", h.getConversation).Methods("GET")
	r.HandleFunc("/conversations/{id}", h.deleteConversation).Methods("DELETE")
	r.HandleFunc("/conversations/{id}/pair", h.findPerspectivePair).Methods("GET")
	r.HandleFunc("/conversations/{id}/summary", h.updateConversationSummary).Methods("PUT")
	r.HandleFunc("/conversations/{id}/reset-session", h.resetConversationSession).Methods("POST")

	return r
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
