package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

// listMCPServers returns all MCP servers.
// @Summary      List MCP servers
// @Description  Returns all configured MCP tool servers
// @Tags         mcps
// @Produce      json
// @Success      200  {array}  store.MCPServer
// @Router       /mcps [get]
func (h *Handler) listMCPServers(w http.ResponseWriter, r *http.Request) {
	mcps := h.store.ListMCPServers()
	writeJSON(w, http.StatusOK, mcps)
}

// getMCPServer returns a single MCP server by name.
// @Summary      Get MCP server
// @Description  Returns an MCP server by its unique name
// @Tags         mcps
// @Produce      json
// @Param        name  path      string  true  "MCP server name"
// @Success      200   {object}  store.MCPServer
// @Failure      404   {object}  ErrorResponse
// @Router       /mcps/{name} [get]
func (h *Handler) getMCPServer(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	m, ok := h.store.GetMCPServer(name)
	if !ok {
		writeError(w, http.StatusNotFound, "MCP server not found")
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// createMCPServer creates a new MCP server.
// @Summary      Create MCP server
// @Description  Creates a new MCP tool server
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
	if err := h.store.CreateMCPServer(m); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

// updateMCPServer updates an existing MCP server.
// @Summary      Update MCP server
// @Description  Updates an MCP server by name
// @Tags         mcps
// @Accept       json
// @Produce      json
// @Param        name  path      string           true  "MCP server name"
// @Param        body  body      store.MCPServer  true  "MCP server definition"
// @Success      200   {object}  store.MCPServer
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /mcps/{name} [put]
func (h *Handler) updateMCPServer(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	var m store.MCPServer
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if m.Name != "" && m.Name != name {
		if err := h.store.RenameMCPServer(name, m.Name); err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		name = m.Name
	}
	if err := h.store.UpdateMCPServer(name, m); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	m.Name = name
	writeJSON(w, http.StatusOK, m)
}

// deleteMCPServer deletes an MCP server.
// @Summary      Delete MCP server
// @Description  Deletes an MCP server by name
// @Tags         mcps
// @Param        name  path  string  true  "MCP server name"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /mcps/{name} [delete]
func (h *Handler) deleteMCPServer(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := h.store.DeleteMCPServer(name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
