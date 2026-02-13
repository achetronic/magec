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

// Schema is a JSON Schema object represented as a plain map so providers can
// define arbitrarily complex schemas (oneOf, if/then, etc.) without the
// framework imposing structural limits.
type Schema = map[string]interface{}

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

	// ConfigSchema returns the JSON Schema describing this provider's configuration.
	// The admin UI renders form inputs dynamically from this schema.
	// Each property corresponds to a key in store.MemoryProvider.Config.
	ConfigSchema() Schema

	// Ping tests the connection using the given config fields.
	Ping(ctx context.Context, config map[string]interface{}) HealthResult
}
