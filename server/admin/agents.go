package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

func (h *Handler) listAgents(w http.ResponseWriter, r *http.Request) {
	agents := h.store.ListAgents()
	writeJSON(w, http.StatusOK, agents)
}

func (h *Handler) getAgent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	a, ok := h.store.GetAgent(id)
	if !ok {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) createAgent(w http.ResponseWriter, r *http.Request) {
	var a store.AgentDefinition
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if a.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	created, err := h.store.CreateAgent(a)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) updateAgent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var a store.AgentDefinition
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := h.store.UpdateAgent(id, a); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetAgent(id)
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteAgent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteAgent(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listAgentMCPs(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	mcps, err := h.store.ResolveAgentMCPs(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, mcps)
}

func (h *Handler) linkAgentMCP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["id"]
	mcpID := vars["mcpId"]
	if err := h.store.LinkAgentMCP(agentID, mcpID); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) unlinkAgentMCP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["id"]
	mcpID := vars["mcpId"]
	if err := h.store.UnlinkAgentMCP(agentID, mcpID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
