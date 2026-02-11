package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

// listAgents returns all agents.
// @Summary      List agents
// @Description  Returns all configured agents
// @Tags         agents
// @Produce      json
// @Success      200  {array}  store.AgentDefinition
// @Router       /agents [get]
func (h *Handler) listAgents(w http.ResponseWriter, r *http.Request) {
	agents := h.store.ListAgents()
	writeJSON(w, http.StatusOK, agents)
}

// getAgent returns a single agent by ID.
// @Summary      Get agent
// @Description  Returns an agent by its unique ID
// @Tags         agents
// @Produce      json
// @Param        id   path      string  true  "Agent ID"
// @Success      200  {object}  store.AgentDefinition
// @Failure      404  {object}  ErrorResponse
// @Router       /agents/{id} [get]
func (h *Handler) getAgent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	a, ok := h.store.GetAgent(id)
	if !ok {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}
	writeJSON(w, http.StatusOK, a)
}

// createAgent creates a new agent.
// @Summary      Create agent
// @Description  Creates a new agent definition
// @Tags         agents
// @Accept       json
// @Produce      json
// @Param        body  body      store.AgentDefinition  true  "Agent definition"
// @Success      201   {object}  store.AgentDefinition
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Router       /agents [post]
func (h *Handler) createAgent(w http.ResponseWriter, r *http.Request) {
	var a store.AgentDefinition
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if a.ID == "" {
		writeError(w, http.StatusBadRequest, "id is required")
		return
	}
	if a.Name == "" {
		a.Name = a.ID
	}
	if err := h.store.CreateAgent(a); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

// updateAgent updates an existing agent.
// @Summary      Update agent
// @Description  Updates an agent by ID
// @Tags         agents
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "Agent ID"
// @Param        body  body      store.AgentDefinition  true  "Agent definition"
// @Success      200   {object}  store.AgentDefinition
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /agents/{id} [put]
func (h *Handler) updateAgent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var a store.AgentDefinition
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if a.ID != "" && a.ID != id {
		if err := h.store.RenameAgent(id, a.ID); err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		id = a.ID
	}
	if err := h.store.UpdateAgent(id, a); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	a.ID = id
	writeJSON(w, http.StatusOK, a)
}

// deleteAgent deletes an agent.
// @Summary      Delete agent
// @Description  Deletes an agent by ID
// @Tags         agents
// @Param        id  path  string  true  "Agent ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /agents/{id} [delete]
func (h *Handler) deleteAgent(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteAgent(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// listAgentMCPs returns the resolved MCP servers linked to an agent.
// @Summary      List agent MCPs
// @Description  Returns the MCP servers linked to a specific agent
// @Tags         agents
// @Produce      json
// @Param        id   path      string  true  "Agent ID"
// @Success      200  {array}   store.MCPServer
// @Failure      404  {object}  ErrorResponse
// @Router       /agents/{id}/mcps [get]
func (h *Handler) listAgentMCPs(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	mcps, err := h.store.ResolveAgentMCPs(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, mcps)
}

// linkAgentMCP links an MCP server to an agent.
// @Summary      Link MCP to agent
// @Description  Associates a global MCP server with an agent
// @Tags         agents
// @Param        id    path  string  true  "Agent ID"
// @Param        name  path  string  true  "MCP server name"
// @Success      204
// @Failure      409  {object}  ErrorResponse
// @Router       /agents/{id}/mcps/{name} [put]
func (h *Handler) linkAgentMCP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["id"]
	mcpName := vars["name"]
	if err := h.store.LinkAgentMCP(agentID, mcpName); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// unlinkAgentMCP unlinks an MCP server from an agent.
// @Summary      Unlink MCP from agent
// @Description  Removes the association between an MCP server and an agent
// @Tags         agents
// @Param        id    path  string  true  "Agent ID"
// @Param        name  path  string  true  "MCP server name"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /agents/{id}/mcps/{name} [delete]
func (h *Handler) unlinkAgentMCP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["id"]
	mcpName := vars["name"]
	if err := h.store.UnlinkAgentMCP(agentID, mcpName); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
