package admin

import (
	"encoding/json"
	"net/http"

	"github.com/achetronic/magec/server/store"
)

// getSettings returns the global runtime settings.
// @Summary      Get settings
// @Description  Returns the global settings (session provider, long-term memory provider).
// @Tags         settings
// @Produce      json
// @Success      200  {object}  store.Settings
// @Router       /settings [get]
func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.store.GetSettings())
}

// updateSettings replaces the global runtime settings.
// @Summary      Update settings
// @Description  Replaces the global settings. Changes take effect on next agent rebuild.
// @Tags         settings
// @Accept       json
// @Produce      json
// @Param        body  body  store.Settings  true  "New settings"
// @Success      200  {object}  store.Settings
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /settings [put]
func (h *Handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	var settings store.Settings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.store.UpdateSettings(settings); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, settings)
}
