package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/memory"
	"github.com/achetronic/magec/server/store"
)

// listMemoryProviders returns all memory providers.
// @Summary      List memory providers
// @Description  Returns all configured memory providers (Redis, Postgres, etc.)
// @Tags         memory
// @Produce      json
// @Success      200  {array}  store.MemoryProvider
// @Router       /memory [get]
func (h *Handler) listMemoryProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.store.ListMemoryProviders()
	writeJSON(w, http.StatusOK, providers)
}

// getMemoryProvider returns a single memory provider by ID.
// @Summary      Get memory provider
// @Description  Returns a memory provider by its unique ID
// @Tags         memory
// @Produce      json
// @Param        id    path      string  true  "Memory Provider ID"
// @Success      200   {object}  store.MemoryProvider
// @Failure      404   {object}  ErrorResponse
// @Router       /memory/{id} [get]
func (h *Handler) getMemoryProvider(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	m, ok := h.store.GetMemoryProvider(id)
	if !ok {
		writeError(w, http.StatusNotFound, "memory provider not found")
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// createMemoryProvider creates a new memory provider.
// @Summary      Create memory provider
// @Description  Creates a new memory provider (session, semantic, etc.)
// @Tags         memory
// @Accept       json
// @Produce      json
// @Param        body  body      store.MemoryProvider  true  "Memory provider definition"
// @Success      201   {object}  store.MemoryProvider
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Router       /memory [post]
func (h *Handler) createMemoryProvider(w http.ResponseWriter, r *http.Request) {
	var m store.MemoryProvider
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if m.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if m.Type == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}
	if m.Category == "" {
		writeError(w, http.StatusBadRequest, "category is required")
		return
	}
	if !memory.ValidType(m.Type) {
		writeError(w, http.StatusBadRequest, "unsupported provider type: "+m.Type)
		return
	}
	if !memory.ValidTypeForCategory(m.Type, memory.Category(m.Category)) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("provider type %q does not support category %q", m.Type, m.Category))
		return
	}
	created, err := h.store.CreateMemoryProvider(m)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// updateMemoryProvider updates an existing memory provider.
// @Summary      Update memory provider
// @Description  Updates a memory provider by ID
// @Tags         memory
// @Accept       json
// @Produce      json
// @Param        id    path      string                true  "Memory Provider ID"
// @Param        body  body      store.MemoryProvider  true  "Memory provider definition"
// @Success      200   {object}  store.MemoryProvider
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /memory/{id} [put]
func (h *Handler) updateMemoryProvider(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var m store.MemoryProvider
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := h.store.UpdateMemoryProvider(id, m); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetMemoryProvider(id)
	writeJSON(w, http.StatusOK, updated)
}

// deleteMemoryProvider deletes a memory provider.
// @Summary      Delete memory provider
// @Description  Deletes a memory provider by ID
// @Tags         memory
// @Param        id  path  string  true  "Memory Provider ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /memory/{id} [delete]
func (h *Handler) deleteMemoryProvider(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteMemoryProvider(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// checkMemoryProviderHealth checks connectivity of a memory provider.
// @Summary      Check memory provider health
// @Description  Pings the memory provider to verify connectivity
// @Tags         memory
// @Produce      json
// @Param        id    path      string  true  "Memory Provider ID"
// @Success      200   {object}  memory.HealthResult
// @Failure      404   {object}  ErrorResponse
// @Router       /memory/{id}/health [get]
func (h *Handler) checkMemoryProviderHealth(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	m, ok := h.store.GetMemoryProvider(id)
	if !ok {
		writeError(w, http.StatusNotFound, "memory provider not found")
		return
	}
	provider := memory.Get(m.Type)
	if provider == nil {
		writeJSON(w, http.StatusOK, memory.HealthResult{
			Healthy: false,
			Detail:  "unsupported provider type: " + m.Type,
		})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	result := provider.Ping(ctx, memoryProviderToMap(m))
	writeJSON(w, http.StatusOK, result)
}

// MemoryTypeInfo represents a registered memory provider type with its JSON Schema.
type MemoryTypeInfo struct {
	Type         string        `json:"type" example:"redis"`
	DisplayName  string        `json:"displayName" example:"Redis"`
	Categories   []string      `json:"categories" example:"session"`
	ConfigSchema memory.Schema `json:"configSchema"`
}

// listMemoryTypes returns all registered memory provider types.
// @Summary      List memory types
// @Description  Returns registered memory provider types with config schemas for dynamic form rendering
// @Tags         memory
// @Produce      json
// @Success      200  {array}  MemoryTypeInfo
// @Router       /memory/types [get]
func (h *Handler) listMemoryTypes(w http.ResponseWriter, r *http.Request) {
	var types []MemoryTypeInfo
	for _, p := range memory.All() {
		cats := make([]string, len(p.SupportedCategories()))
		for i, c := range p.SupportedCategories() {
			cats[i] = string(c)
		}
		types = append(types, MemoryTypeInfo{
			Type:         p.Type(),
			DisplayName:  p.DisplayName(),
			Categories:   cats,
			ConfigSchema: p.ConfigSchema(),
		})
	}
	writeJSON(w, http.StatusOK, types)
}

func memoryProviderToMap(m store.MemoryProvider) map[string]interface{} {
	if m.Config == nil {
		return map[string]interface{}{}
	}
	return m.Config
}
