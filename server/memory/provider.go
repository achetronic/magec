package memory

import (
	"context"
)

// Category classifies what role a memory provider instance serves.
// A single provider type (e.g. Redis) may support both categories
// with different configurations, but each configured instance
// serves exactly one role.
type Category string

const (
	CategorySession  Category = "session"
	CategoryLongTerm Category = "longterm"
)

// HealthResult holds the result of a connection health check.
type HealthResult struct {
	Healthy bool   `json:"healthy"`
	Detail  string `json:"detail"`
}

// FieldSpec describes a single configuration field for a provider.
// The admin UI uses this to dynamically render forms — no hardcoded
// fields per provider type. Adding a new provider with new fields
// requires zero UI changes.
type FieldSpec struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Required    bool   `json:"required,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Default     string `json:"default,omitempty"`
}

// Provider defines what every memory provider type must implement.
// To add a new provider (e.g. Memcached, Qdrant, Milvus):
//  1. Create a new package under server/memory/<name>/
//  2. Implement the Provider interface
//  3. Call memory.Register() in an init() function
//
// See server/memory/redis/ and server/memory/postgres/ for examples.
type Provider interface {
	// Type returns the unique identifier for this provider (e.g. "redis", "postgres").
	Type() string

	// DisplayName returns a human-readable name for the admin UI.
	DisplayName() string

	// SupportedCategories returns the memory roles this provider type can fill.
	SupportedCategories() []Category

	// ConfigFields returns the field specifications for this provider's config.
	// The admin UI renders form inputs dynamically from these specs.
	// Each field corresponds to a key in store.MemoryProvider.Config.
	ConfigFields() []FieldSpec

	// Ping tests the connection using the given config fields.
	Ping(ctx context.Context, config map[string]interface{}) HealthResult
}
