package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

// ErrorResponse is returned for all error responses.
type ErrorResponse struct {
	Error string `json:"error" example:"resource not found"`
}

// Handler provides the admin API router.
type Handler struct {
	store  *store.Store
	router *mux.Router
}

// New creates a new admin API handler.
func New(s *store.Store) *Handler {
	h := &Handler{store: s}
	h.router = h.buildRouter()
	return h
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

	// Cron Jobs
	r.HandleFunc("/crons", h.listCronJobs).Methods("GET")
	r.HandleFunc("/crons", h.createCronJob).Methods("POST")
	r.HandleFunc("/crons/{id}", h.getCronJob).Methods("GET")
	r.HandleFunc("/crons/{id}", h.updateCronJob).Methods("PUT")
	r.HandleFunc("/crons/{id}", h.deleteCronJob).Methods("DELETE")

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
