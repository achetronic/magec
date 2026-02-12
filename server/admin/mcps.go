package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

func (h *Handler) listMCPServers(w http.ResponseWriter, r *http.Request) {
	mcps := h.store.ListMCPServers()
	writeJSON(w, http.StatusOK, mcps)
}

func (h *Handler) getMCPServer(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	m, ok := h.store.GetMCPServer(id)
	if !ok {
		writeError(w, http.StatusNotFound, "MCP server not found")
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (h *Handler) createMCPServer(w http.ResponseWriter, r *http.Request) {
	var m store.MCPServer
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if m.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	created, err := h.store.CreateMCPServer(m)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) updateMCPServer(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var m store.MCPServer
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := h.store.UpdateMCPServer(id, m); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetMCPServer(id)
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteMCPServer(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteMCPServer(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
