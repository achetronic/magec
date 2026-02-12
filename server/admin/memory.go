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

func (h *Handler) listMemoryProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.store.ListMemoryProviders()
	writeJSON(w, http.StatusOK, providers)
}

func (h *Handler) getMemoryProvider(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	m, ok := h.store.GetMemoryProvider(id)
	if !ok {
		writeError(w, http.StatusNotFound, "memory provider not found")
		return
	}
	writeJSON(w, http.StatusOK, m)
}

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

func (h *Handler) deleteMemoryProvider(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteMemoryProvider(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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

type MemoryTypeInfo struct {
	Type        string             `json:"type" example:"redis"`
	DisplayName string             `json:"displayName" example:"Redis"`
	Categories  []string           `json:"categories" example:"session"`
	Fields      []memory.FieldSpec `json:"fields"`
}

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
