package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

// listBackends returns all backends.
// @Summary      List backends
// @Description  Returns all configured AI backends
// @Tags         backends
// @Produce      json
// @Success      200  {array}  store.BackendDefinition
// @Router       /backends [get]
func (h *Handler) listBackends(w http.ResponseWriter, r *http.Request) {
	backends := h.store.ListBackends()
	writeJSON(w, http.StatusOK, backends)
}

// getBackend returns a single backend by name.
// @Summary      Get backend
// @Description  Returns a backend by its unique name
// @Tags         backends
// @Produce      json
// @Param        name  path      string  true  "Backend name"
// @Success      200   {object}  store.BackendDefinition
// @Failure      404   {object}  ErrorResponse
// @Router       /backends/{name} [get]
func (h *Handler) getBackend(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	b, ok := h.store.GetBackend(name)
	if !ok {
		writeError(w, http.StatusNotFound, "backend not found")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// createBackend creates a new backend.
// @Summary      Create backend
// @Description  Creates a new AI backend
// @Tags         backends
// @Accept       json
// @Produce      json
// @Param        body  body      store.BackendDefinition  true  "Backend definition"
// @Success      201   {object}  store.BackendDefinition
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Router       /backends [post]
func (h *Handler) createBackend(w http.ResponseWriter, r *http.Request) {
	var b store.BackendDefinition
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if b.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if b.Type == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}
	if err := h.store.CreateBackend(b); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

// updateBackend updates an existing backend.
// @Summary      Update backend
// @Description  Updates a backend by name
// @Tags         backends
// @Accept       json
// @Produce      json
// @Param        name  path      string                   true  "Backend name"
// @Param        body  body      store.BackendDefinition  true  "Backend definition"
// @Success      200   {object}  store.BackendDefinition
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /backends/{name} [put]
func (h *Handler) updateBackend(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	var b store.BackendDefinition
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if b.Name != "" && b.Name != name {
		if err := h.store.RenameBackend(name, b.Name); err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		name = b.Name
	}
	if err := h.store.UpdateBackend(name, b); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	b.Name = name
	writeJSON(w, http.StatusOK, b)
}

// deleteBackend deletes a backend.
// @Summary      Delete backend
// @Description  Deletes a backend by name
// @Tags         backends
// @Param        name  path  string  true  "Backend name"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /backends/{name} [delete]
func (h *Handler) deleteBackend(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := h.store.DeleteBackend(name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
