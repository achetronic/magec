package mcp

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/memory"
	"github.com/achetronic/magec/server/store"
)

type listMemoryOutput struct {
	Providers []store.MemoryProvider `json:"providers"`
}

func (h *Handler) listMemoryProviders(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, listMemoryOutput, error) {
	return nil, listMemoryOutput{Providers: h.store.ListRawMemoryProviders()}, nil
}

func (h *Handler) getMemoryProvider(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, store.MemoryProvider, error) {
	m, ok := h.store.GetRawMemoryProvider(in.ID)
	if !ok {
		return nil, store.MemoryProvider{}, fmt.Errorf("get memory provider: %w", errValidation("not found: "+in.ID))
	}
	return nil, m, nil
}

type createMemoryInput struct {
	Provider store.MemoryProvider `json:"provider" jsonschema:"memory provider definition"`
}

func (h *Handler) createMemoryProvider(_ context.Context, _ *sdk.CallToolRequest, in createMemoryInput) (*sdk.CallToolResult, store.MemoryProvider, error) {
	m := in.Provider
	if m.Name == "" {
		return nil, store.MemoryProvider{}, fmt.Errorf("create memory provider: %w", errValidation("name is required"))
	}
	if m.Type == "" {
		return nil, store.MemoryProvider{}, fmt.Errorf("create memory provider: %w", errValidation("type is required"))
	}
	if m.Category == "" {
		return nil, store.MemoryProvider{}, fmt.Errorf("create memory provider: %w", errValidation("category is required"))
	}
	if !memory.ValidType(m.Type) {
		return nil, store.MemoryProvider{}, fmt.Errorf("create memory provider: %w", errValidation("unsupported provider type: "+m.Type))
	}
	if !memory.ValidTypeForCategory(m.Type, memory.Category(m.Category)) {
		return nil, store.MemoryProvider{}, fmt.Errorf("create memory provider: %w", errValidation(fmt.Sprintf("provider type %q does not support category %q", m.Type, m.Category)))
	}
	created, err := h.store.CreateMemoryProvider(m)
	if err != nil {
		return nil, store.MemoryProvider{}, fmt.Errorf("create memory provider: %w", err)
	}
	return nil, created, nil
}

type updateMemoryInput struct {
	ID       string               `json:"id" jsonschema:"memory provider id"`
	Provider store.MemoryProvider `json:"provider" jsonschema:"new memory provider definition"`
}

func (h *Handler) updateMemoryProvider(_ context.Context, _ *sdk.CallToolRequest, in updateMemoryInput) (*sdk.CallToolResult, store.MemoryProvider, error) {
	if err := h.store.UpdateMemoryProvider(in.ID, in.Provider); err != nil {
		return nil, store.MemoryProvider{}, fmt.Errorf("update memory provider: %w", err)
	}
	updated, _ := h.store.GetRawMemoryProvider(in.ID)
	return nil, updated, nil
}

func (h *Handler) deleteMemoryProvider(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, emptyOutput, error) {
	if err := h.store.DeleteMemoryProvider(in.ID); err != nil {
		return nil, emptyOutput{}, fmt.Errorf("delete memory provider: %w", err)
	}
	return nil, okOutput, nil
}

func (h *Handler) checkMemoryHealth(ctx context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, memory.HealthResult, error) {
	m, ok := h.store.GetMemoryProvider(in.ID)
	if !ok {
		return nil, memory.HealthResult{}, fmt.Errorf("check memory health: %w", errValidation("not found: "+in.ID))
	}
	provider := memory.Get(m.Type)
	if provider == nil {
		return nil, memory.HealthResult{Healthy: false, Detail: "unsupported provider type: " + m.Type}, nil
	}
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cfg := m.Config
	if cfg == nil {
		cfg = map[string]interface{}{}
	}
	return nil, provider.Ping(checkCtx, cfg), nil
}

type memoryTypeInfo struct {
	Type         string        `json:"type"`
	DisplayName  string        `json:"displayName"`
	Categories   []string      `json:"categories"`
	ConfigSchema memory.Schema `json:"configSchema"`
}

type listMemoryTypesOutput struct {
	Types []memoryTypeInfo `json:"types"`
}

func (h *Handler) listMemoryTypes(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, listMemoryTypesOutput, error) {
	var types []memoryTypeInfo
	for _, p := range memory.All() {
		cats := make([]string, len(p.SupportedCategories()))
		for i, c := range p.SupportedCategories() {
			cats[i] = string(c)
		}
		types = append(types, memoryTypeInfo{
			Type:         p.Type(),
			DisplayName:  p.DisplayName(),
			Categories:   cats,
			ConfigSchema: p.ConfigSchema(),
		})
	}
	return nil, listMemoryTypesOutput{Types: types}, nil
}

func (h *Handler) registerMemoryTools() {
	destructive := true
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_list_memory_providers", Title: "List memory providers",
		Description: "List every configured memory provider (Redis session, Postgres long-term, etc.).",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.listMemoryProviders)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_get_memory_provider", Title: "Get memory provider",
		Description: "Return one memory provider by id.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.getMemoryProvider)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_create_memory_provider", Title: "Create memory provider",
		Description: "Create a new memory provider. Type must be registered (see magec_list_memory_types) and compatible with the category.",
	}, h.createMemoryProvider)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_update_memory_provider", Title: "Update memory provider",
		Description: "Replace the memory provider identified by id.",
		Annotations: &sdk.ToolAnnotations{IdempotentHint: true},
	}, h.updateMemoryProvider)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_delete_memory_provider", Title: "Delete memory provider",
		Description: "Delete a memory provider by id.",
		Annotations: &sdk.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, h.deleteMemoryProvider)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_check_memory_health", Title: "Check memory provider health",
		Description: "Ping the memory provider's backing service (5s timeout). Returns connectivity status.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.checkMemoryHealth)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_list_memory_types", Title: "List memory provider types",
		Description: "List registered memory provider types with their JSON schemas.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.listMemoryTypes)
	h.toolCount++
}
