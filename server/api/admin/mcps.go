package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

// listMCPServers returns all MCP servers.
// @Summary      List MCP servers
// @Description  Returns all configured MCP (Model Context Protocol) servers
// @Tags         mcps
// @Produce      json
// @Success      200  {array}  store.MCPServer
// @Router       /mcps [get]
func (h *Handler) listMCPServers(w http.ResponseWriter, r *http.Request) {
	mcps := h.store.ListMCPServers()
	writeJSON(w, http.StatusOK, mcps)
}

// getMCPServer returns a single MCP server by ID.
// @Summary      Get MCP server
// @Description  Returns an MCP server by its unique ID
// @Tags         mcps
// @Produce      json
// @Param        id    path      string  true  "MCP Server ID"
// @Success      200   {object}  store.MCPServer
// @Failure      404   {object}  ErrorResponse
// @Router       /mcps/{id} [get]
func (h *Handler) getMCPServer(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	m, ok := h.store.GetMCPServer(id)
	if !ok {
		writeError(w, http.StatusNotFound, "MCP server not found")
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// createMCPServer creates a new MCP server.
// @Summary      Create MCP server
// @Description  Creates a new MCP server configuration
// @Tags         mcps
// @Accept       json
// @Produce      json
// @Param        body  body      store.MCPServer  true  "MCP server definition"
// @Success      201   {object}  store.MCPServer
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Router       /mcps [post]
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

// updateMCPServer updates an existing MCP server.
// @Summary      Update MCP server
// @Description  Updates an MCP server by ID
// @Tags         mcps
// @Accept       json
// @Produce      json
// @Param        id    path      string           true  "MCP Server ID"
// @Param        body  body      store.MCPServer  true  "MCP server definition"
// @Success      200   {object}  store.MCPServer
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /mcps/{id} [put]
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

// deleteMCPServer deletes an MCP server.
// @Summary      Delete MCP server
// @Description  Deletes an MCP server by ID
// @Tags         mcps
// @Param        id  path  string  true  "MCP Server ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /mcps/{id} [delete]
func (h *Handler) deleteMCPServer(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteMCPServer(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
