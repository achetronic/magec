package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

// SecretResponse is the public representation of a secret (value is never exposed).
type SecretResponse struct {
	ID          string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string `json:"name" example:"OpenAI API Key"`
	Key         string `json:"key" example:"OPENAI_API_KEY"`
	Description string `json:"description,omitempty" example:"Production key for GPT-4o"`
}

// SecretCreateRequest is the payload for creating a secret.
type SecretCreateRequest struct {
	Name        string `json:"name" example:"OpenAI API Key"`
	Key         string `json:"key" example:"OPENAI_API_KEY"`
	Value       string `json:"value" example:"sk-..."`
	Description string `json:"description,omitempty" example:"Production key for GPT-4o"`
}

// SecretUpdateRequest is the payload for updating a secret. Value is optional — omit to keep current.
type SecretUpdateRequest struct {
	Name        string `json:"name" example:"OpenAI API Key"`
	Key         string `json:"key" example:"OPENAI_API_KEY"`
	Value       string `json:"value,omitempty" example:"sk-..."`
	Description string `json:"description,omitempty" example:"Production key for GPT-4o"`
}

func secretResponse(s store.Secret) SecretResponse {
	return SecretResponse{
		ID:          s.ID,
		Name:        s.Name,
		Key:         s.Key,
		Description: s.Description,
	}
}

// listSecrets returns all secrets (values are never exposed).
// @Summary      List secrets
// @Description  Returns all configured secrets without their values
// @Tags         secrets
// @Produce      json
// @Security     AdminAuth
// @Success      200  {array}  SecretResponse
// @Router       /secrets [get]
func (h *Handler) listSecrets(w http.ResponseWriter, r *http.Request) {
	secrets := h.store.ListSecrets()
	result := make([]SecretResponse, len(secrets))
	for i, s := range secrets {
		result[i] = secretResponse(s)
	}
	writeJSON(w, http.StatusOK, result)
}

// getSecret returns a single secret by ID (value is never exposed).
// @Summary      Get secret
// @Description  Returns a secret by its unique ID without the value
// @Tags         secrets
// @Produce      json
// @Security     AdminAuth
// @Param        id    path      string  true  "Secret ID"
// @Success      200   {object}  SecretResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /secrets/{id} [get]
func (h *Handler) getSecret(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	s, ok := h.store.GetSecret(id)
	if !ok {
		writeError(w, http.StatusNotFound, "secret not found")
		return
	}
	writeJSON(w, http.StatusOK, secretResponse(s))
}

// createSecret creates a new secret.
// @Summary      Create secret
// @Description  Creates a new secret. The value is encrypted at rest when adminPassword is configured.
// @Tags         secrets
// @Accept       json
// @Produce      json
// @Security     AdminAuth
// @Param        body  body      SecretCreateRequest  true  "Secret definition"
// @Success      201   {object}  SecretResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Router       /secrets [post]
func (h *Handler) createSecret(w http.ResponseWriter, r *http.Request) {
	var s store.Secret
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if s.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if s.Key == "" {
		writeError(w, http.StatusBadRequest, "key is required")
		return
	}
	if s.Value == "" {
		writeError(w, http.StatusBadRequest, "value is required")
		return
	}
	created, err := h.store.CreateSecret(s)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, secretResponse(created))
}

// updateSecret updates an existing secret.
// @Summary      Update secret
// @Description  Updates a secret by ID. Omit value to keep the current one.
// @Tags         secrets
// @Accept       json
// @Produce      json
// @Security     AdminAuth
// @Param        id    path      string               true  "Secret ID"
// @Param        body  body      SecretUpdateRequest   true  "Secret definition"
// @Success      200   {object}  SecretResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /secrets/{id} [put]
func (h *Handler) updateSecret(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var s store.Secret
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if s.Value == "" {
		existing, ok := h.store.GetSecret(id)
		if ok {
			s.Value = existing.Value
		}
	}
	if err := h.store.UpdateSecret(id, s); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetSecret(id)
	writeJSON(w, http.StatusOK, secretResponse(updated))
}

// deleteSecret deletes a secret.
// @Summary      Delete secret
// @Description  Deletes a secret by ID
// @Tags         secrets
// @Security     AdminAuth
// @Param        id  path  string  true  "Secret ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /secrets/{id} [delete]
func (h *Handler) deleteSecret(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteSecret(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
