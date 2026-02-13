package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

func (h *Handler) listCommands(w http.ResponseWriter, r *http.Request) {
	commands := h.store.ListCommands()
	writeJSON(w, http.StatusOK, commands)
}

func (h *Handler) getCommand(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c, ok := h.store.GetCommand(id)
	if !ok {
		writeError(w, http.StatusNotFound, "command not found")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *Handler) createCommand(w http.ResponseWriter, r *http.Request) {
	var c store.Command
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if c.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if c.Prompt == "" {
		writeError(w, http.StatusBadRequest, "prompt is required")
		return
	}
	created, err := h.store.CreateCommand(c)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) updateCommand(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var c store.Command
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := h.store.UpdateCommand(id, c); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetCommand(id)
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteCommand(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteCommand(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
