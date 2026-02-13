package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

// listBackends returns all backends.
// @Summary      List backends
// @Description  Returns all configured LLM/TTS/transcription backends
// @Tags         backends
// @Produce      json
// @Success      200  {array}  store.BackendDefinition
// @Router       /backends [get]
func (h *Handler) listBackends(w http.ResponseWriter, r *http.Request) {
	backends := h.store.ListBackends()
	writeJSON(w, http.StatusOK, backends)
}

// getBackend returns a single backend by ID.
// @Summary      Get backend
// @Description  Returns a backend by its unique ID
// @Tags         backends
// @Produce      json
// @Param        id    path      string  true  "Backend ID"
// @Success      200   {object}  store.BackendDefinition
// @Failure      404   {object}  ErrorResponse
// @Router       /backends/{id} [get]
func (h *Handler) getBackend(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	b, ok := h.store.GetBackend(id)
	if !ok {
		writeError(w, http.StatusNotFound, "backend not found")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// createBackend creates a new backend.
// @Summary      Create backend
// @Description  Creates a new LLM/TTS/transcription backend
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
	created, err := h.store.CreateBackend(b)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// updateBackend updates an existing backend.
// @Summary      Update backend
// @Description  Updates a backend by ID
// @Tags         backends
// @Accept       json
// @Produce      json
// @Param        id    path      string                   true  "Backend ID"
// @Param        body  body      store.BackendDefinition  true  "Backend definition"
// @Success      200   {object}  store.BackendDefinition
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /backends/{id} [put]
func (h *Handler) updateBackend(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var b store.BackendDefinition
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := h.store.UpdateBackend(id, b); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetBackend(id)
	writeJSON(w, http.StatusOK, updated)
}

// deleteBackend deletes a backend.
// @Summary      Delete backend
// @Description  Deletes a backend by ID
// @Tags         backends
// @Param        id  path  string  true  "Backend ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /backends/{id} [delete]
func (h *Handler) deleteBackend(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteBackend(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
