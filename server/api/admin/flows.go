package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/agent/flowgraph"
	"github.com/achetronic/magec/server/store"
)

// listFlows returns all flows.
// @Summary      List flows
// @Description  Returns all configured agent orchestration flows
// @Tags         flows
// @Produce      json
// @Success      200  {array}  store.FlowDefinition
// @Security     AdminAuth
// @Router       /flows [get]
func (h *Handler) listFlows(w http.ResponseWriter, r *http.Request) {
	flows := h.store.ListRawFlows()
	writeJSON(w, http.StatusOK, flows)
}

// getFlow returns a single flow by ID.
// @Summary      Get flow
// @Description  Returns a flow by its unique ID
// @Tags         flows
// @Produce      json
// @Param        id    path      string  true  "Flow ID"
// @Success      200   {object}  store.FlowDefinition
// @Failure      404   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /flows/{id} [get]
func (h *Handler) getFlow(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	f, ok := h.store.GetRawFlow(id)
	if !ok {
		writeError(w, http.StatusNotFound, "flow not found")
		return
	}
	writeJSON(w, http.StatusOK, f)
}

// createFlow creates a new flow.
// @Summary      Create flow
// @Description  Creates a new agent orchestration flow defined as a directed graph of nodes and edges
// @Tags         flows
// @Accept       json
// @Produce      json
// @Param        body  body      store.FlowDefinition  true  "Flow definition (nodes + edges graph)"
// @Success      201   {object}  store.FlowDefinition
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /flows [post]
func (h *Handler) createFlow(w http.ResponseWriter, r *http.Request) {
	var f store.FlowDefinition
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if f.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if err := flowgraph.Validate(&f); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.validateFlowSecrets(&f); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.store.CreateFlow(f)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// updateFlow updates an existing flow.
// @Summary      Update flow
// @Description  Updates a flow by ID
// @Tags         flows
// @Accept       json
// @Produce      json
// @Param        id    path      string                true  "Flow ID"
// @Param        body  body      store.FlowDefinition  true  "Flow definition"
// @Success      200   {object}  store.FlowDefinition
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /flows/{id} [put]
func (h *Handler) updateFlow(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var f store.FlowDefinition
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := flowgraph.Validate(&f); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.validateFlowSecrets(&f); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.UpdateFlow(id, f); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetRawFlow(id)
	writeJSON(w, http.StatusOK, updated)
}

// deleteFlow deletes a flow.
// @Summary      Delete flow
// @Description  Deletes a flow by ID
// @Tags         flows
// @Param        id     path   string   true   "Flow ID"
// @Param        force  query  boolean  false  "Scrub all references to this entity and delete anyway"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Failure      409  {object}  ReferencesResponse  "Flow is referenced by other entities"
// @Security     AdminAuth
// @Router       /flows/{id} [delete]
func (h *Handler) deleteFlow(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if !h.deleteGuard(w, r, id) {
		return
	}
	if err := h.store.DeleteFlow(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
