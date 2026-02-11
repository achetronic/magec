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
// @Description  Returns all configured memory providers (session and long-term)
// @Tags         memory
// @Produce      json
// @Success      200  {array}  store.MemoryProvider
// @Router       /memory [get]
func (h *Handler) listMemoryProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.store.ListMemoryProviders()
	writeJSON(w, http.StatusOK, providers)
}

// getMemoryProvider returns a single memory provider by name.
// @Summary      Get memory provider
// @Description  Returns a memory provider by its unique name
// @Tags         memory
// @Produce      json
// @Param        name  path      string  true  "Provider name"
// @Success      200   {object}  store.MemoryProvider
// @Failure      404   {object}  ErrorResponse
// @Router       /memory/{name} [get]
func (h *Handler) getMemoryProvider(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	m, ok := h.store.GetMemoryProvider(name)
	if !ok {
		writeError(w, http.StatusNotFound, "memory provider not found")
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// createMemoryProvider creates a new memory provider.
// @Summary      Create memory provider
// @Description  Creates a new memory provider. Type must be registered and support the given category.
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
	if err := h.store.CreateMemoryProvider(m); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

// updateMemoryProvider updates an existing memory provider.
// @Summary      Update memory provider
// @Description  Updates a memory provider by name
// @Tags         memory
// @Accept       json
// @Produce      json
// @Param        name  path      string                true  "Provider name"
// @Param        body  body      store.MemoryProvider  true  "Memory provider definition"
// @Success      200   {object}  store.MemoryProvider
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /memory/{name} [put]
func (h *Handler) updateMemoryProvider(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	var m store.MemoryProvider
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if m.Name != "" && m.Name != name {
		if err := h.store.RenameMemoryProvider(name, m.Name); err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		name = m.Name
	}
	if err := h.store.UpdateMemoryProvider(name, m); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	m.Name = name
	writeJSON(w, http.StatusOK, m)
}

// deleteMemoryProvider deletes a memory provider.
// @Summary      Delete memory provider
// @Description  Deletes a memory provider by name
// @Tags         memory
// @Param        name  path  string  true  "Provider name"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /memory/{name} [delete]
func (h *Handler) deleteMemoryProvider(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := h.store.DeleteMemoryProvider(name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// checkMemoryProviderHealth performs a real-time connection test.
// @Summary      Health check
// @Description  Pings the memory provider with a 5-second timeout to verify connectivity
// @Tags         memory
// @Produce      json
// @Param        name  path      string  true  "Provider name"
// @Success      200   {object}  memory.HealthResult
// @Failure      404   {object}  ErrorResponse
// @Router       /memory/{name}/health [get]
func (h *Handler) checkMemoryProviderHealth(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	m, ok := h.store.GetMemoryProvider(name)
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

// MemoryTypeInfo represents a registered memory provider type with its capabilities.
type MemoryTypeInfo struct {
	Type        string             `json:"type" example:"redis"`
	DisplayName string             `json:"displayName" example:"Redis"`
	Categories  []string           `json:"categories" example:"session"`
	Fields      []memory.FieldSpec `json:"fields"`
}

// listMemoryTypes returns all registered provider types with field specs.
// @Summary      List memory provider types
// @Description  Returns registered provider types with supported categories and config field specifications for dynamic form rendering
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
			Type:        p.Type(),
			DisplayName: p.DisplayName(),
			Categories:  cats,
			Fields:      p.ConfigFields(),
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
