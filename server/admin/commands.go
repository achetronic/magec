package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

// listCommands returns all commands.
// @Summary      List commands
// @Description  Returns all configured reusable prompt commands
// @Tags         commands
// @Produce      json
// @Success      200  {array}  store.Command
// @Router       /commands [get]
func (h *Handler) listCommands(w http.ResponseWriter, r *http.Request) {
	commands := h.store.ListCommands()
	writeJSON(w, http.StatusOK, commands)
}

// getCommand returns a single command by ID.
// @Summary      Get command
// @Description  Returns a command by its unique ID
// @Tags         commands
// @Produce      json
// @Param        id    path      string  true  "Command ID"
// @Success      200   {object}  store.Command
// @Failure      404   {object}  ErrorResponse
// @Router       /commands/{id} [get]
func (h *Handler) getCommand(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c, ok := h.store.GetCommand(id)
	if !ok {
		writeError(w, http.StatusNotFound, "command not found")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// createCommand creates a new command.
// @Summary      Create command
// @Description  Creates a new reusable prompt command
// @Tags         commands
// @Accept       json
// @Produce      json
// @Param        body  body      store.Command  true  "Command definition"
// @Success      201   {object}  store.Command
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Router       /commands [post]
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

// updateCommand updates an existing command.
// @Summary      Update command
// @Description  Updates a command by ID
// @Tags         commands
// @Accept       json
// @Produce      json
// @Param        id    path      string         true  "Command ID"
// @Param        body  body      store.Command  true  "Command definition"
// @Success      200   {object}  store.Command
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /commands/{id} [put]
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

// deleteCommand deletes a command.
// @Summary      Delete command
// @Description  Deletes a command by ID
// @Tags         commands
// @Param        id  path  string  true  "Command ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /commands/{id} [delete]
func (h *Handler) deleteCommand(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteCommand(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
