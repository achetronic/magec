package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

func (h *Handler) listBackends(w http.ResponseWriter, r *http.Request) {
	backends := h.store.ListBackends()
	writeJSON(w, http.StatusOK, backends)
}

func (h *Handler) getBackend(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	b, ok := h.store.GetBackend(id)
	if !ok {
		writeError(w, http.StatusNotFound, "backend not found")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

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

func (h *Handler) deleteBackend(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteBackend(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
