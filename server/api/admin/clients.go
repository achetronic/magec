package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/clients"
	"github.com/achetronic/magec/server/store"
)

// listClients returns all clients.
// @Summary      List clients
// @Description  Returns all configured clients (devices, Telegram bots, etc.)
// @Tags         clients
// @Produce      json
// @Success      200  {array}  store.ClientDefinition
// @Router       /clients [get]
func (h *Handler) listClients(w http.ResponseWriter, r *http.Request) {
	clients := h.store.ListClients()
	writeJSON(w, http.StatusOK, clients)
}

// getClient returns a single client by ID.
// @Summary      Get client
// @Description  Returns a client by its unique ID
// @Tags         clients
// @Produce      json
// @Param        id    path      string  true  "Client ID"
// @Success      200   {object}  store.ClientDefinition
// @Failure      404   {object}  ErrorResponse
// @Router       /clients/{id} [get]
func (h *Handler) getClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c, ok := h.store.GetClient(id)
	if !ok {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// createClient creates a new client with an auto-generated token.
// @Summary      Create client
// @Description  Creates a new client. A unique auth token is generated automatically.
// @Tags         clients
// @Accept       json
// @Produce      json
// @Param        body  body      store.ClientDefinition  true  "Client definition (token is auto-generated)"
// @Success      201   {object}  store.ClientDefinition
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Router       /clients [post]
func (h *Handler) createClient(w http.ResponseWriter, r *http.Request) {
	var c store.ClientDefinition
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if c.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if c.Type == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}
	if !clients.ValidType(c.Type) {
		writeError(w, http.StatusBadRequest, "unsupported client type: "+c.Type)
		return
	}
	if err := validateClientConfig(c); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.store.CreateClient(c)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// updateClient updates an existing client.
// @Summary      Update client
// @Description  Updates a client by ID. Token and ID are preserved.
// @Tags         clients
// @Accept       json
// @Produce      json
// @Param        id    path      string                  true  "Client ID"
// @Param        body  body      store.ClientDefinition  true  "Client definition"
// @Success      200   {object}  store.ClientDefinition
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /clients/{id} [put]
func (h *Handler) updateClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var c store.ClientDefinition
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if c.Type != "" {
		if err := validateClientConfig(c); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if err := h.store.UpdateClient(id, c); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetClient(id)
	writeJSON(w, http.StatusOK, updated)
}

// deleteClient deletes a client.
// @Summary      Delete client
// @Description  Deletes a client by ID, revoking its access token
// @Tags         clients
// @Param        id  path  string  true  "Client ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /clients/{id} [delete]
func (h *Handler) deleteClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteClient(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// regenerateClientToken generates a new auth token for a client.
// @Summary      Regenerate client token
// @Description  Generates a new authentication token for a client, invalidating the previous one
// @Tags         clients
// @Produce      json
// @Param        id    path      string  true  "Client ID"
// @Success      200   {object}  store.ClientDefinition
// @Failure      404   {object}  ErrorResponse
// @Router       /clients/{id}/regenerate-token [post]
func (h *Handler) regenerateClientToken(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	cl, err := h.store.RegenerateClientToken(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cl)
}

// ClientTypeInfo represents a registered client type with its JSON Schema.
type ClientTypeInfo struct {
	Type         string        `json:"type" example:"telegram"`
	DisplayName  string        `json:"displayName" example:"Telegram"`
	ConfigSchema clients.Schema `json:"configSchema"`
}

// listClientTypes returns all registered client types with field specs.
// @Summary      List client types
// @Description  Returns registered client types with config field specifications for dynamic form rendering
// @Tags         clients
// @Produce      json
// @Success      200  {array}  ClientTypeInfo
// @Router       /clients/types [get]
func (h *Handler) listClientTypes(w http.ResponseWriter, r *http.Request) {
	var types []ClientTypeInfo
	for _, p := range clients.All() {
		types = append(types, ClientTypeInfo{
			Type:         p.Type(),
			DisplayName:  p.DisplayName(),
			ConfigSchema: p.ConfigSchema(),
		})
	}
	writeJSON(w, http.StatusOK, types)
}

func validateClientConfig(c store.ClientDefinition) error {
	raw, err := json.Marshal(c.Config)
	if err != nil {
		return nil
	}
	var full map[string]map[string]interface{}
	if err := json.Unmarshal(raw, &full); err != nil {
		return nil
	}
	return clients.ValidateConfig(c.Type, full[c.Type])
}
