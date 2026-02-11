package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

// listDevices returns all devices.
// @Summary      List devices
// @Description  Returns all configured access point devices
// @Tags         devices
// @Produce      json
// @Success      200  {array}  store.Device
// @Router       /devices [get]
func (h *Handler) listDevices(w http.ResponseWriter, r *http.Request) {
	devices := h.store.ListDevices()
	writeJSON(w, http.StatusOK, devices)
}

// getDevice returns a single device by name.
// @Summary      Get device
// @Description  Returns a device by its unique name
// @Tags         devices
// @Produce      json
// @Param        name  path      string  true  "Device name"
// @Success      200   {object}  store.Device
// @Failure      404   {object}  ErrorResponse
// @Router       /devices/{name} [get]
func (h *Handler) getDevice(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	d, ok := h.store.GetDevice(name)
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	writeJSON(w, http.StatusOK, d)
}

// createDevice creates a new device with an auto-generated token.
// @Summary      Create device
// @Description  Creates a new device. A unique auth token is generated automatically.
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        body  body      store.Device  true  "Device definition (token is auto-generated)"
// @Success      201   {object}  store.Device
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Router       /devices [post]
func (h *Handler) createDevice(w http.ResponseWriter, r *http.Request) {
	var d store.Device
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if d.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if d.DefaultAgent == "" {
		writeError(w, http.StatusBadRequest, "defaultAgent is required")
		return
	}
	created, err := h.store.CreateDevice(d)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// updateDevice updates an existing device.
// @Summary      Update device
// @Description  Updates a device by name. Token is preserved.
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        name  path      string        true  "Device name"
// @Param        body  body      store.Device  true  "Device definition"
// @Success      200   {object}  store.Device
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /devices/{name} [put]
func (h *Handler) updateDevice(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	var d store.Device
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if d.Name != "" && d.Name != name {
		if err := h.store.RenameDevice(name, d.Name); err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		name = d.Name
	}
	if err := h.store.UpdateDevice(name, d); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetDevice(name)
	writeJSON(w, http.StatusOK, updated)
}

// deleteDevice deletes a device.
// @Summary      Delete device
// @Description  Deletes a device by name, revoking its access token
// @Tags         devices
// @Param        name  path  string  true  "Device name"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /devices/{name} [delete]
func (h *Handler) deleteDevice(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := h.store.DeleteDevice(name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// regenerateDeviceToken generates a new auth token for a device.
// @Summary      Regenerate device token
// @Description  Generates a new authentication token for a device, invalidating the previous one
// @Tags         devices
// @Produce      json
// @Param        name  path      string  true  "Device name"
// @Success      200   {object}  store.Device
// @Failure      404   {object}  ErrorResponse
// @Router       /devices/{name}/regenerate-token [post]
func (h *Handler) regenerateDeviceToken(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	device, err := h.store.RegenerateDeviceToken(name)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, device)
}
