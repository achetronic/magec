package memory

import (
	"fmt"
	"sort"
	"sync"

	"github.com/achetronic/magec/server/schema"
)

var (
	mu        sync.RWMutex
	providers = map[string]Provider{}
)

// Register adds a provider to the global registry.
// Call this from an init() function in each provider package.
// Panics if a provider with the same type is already registered.
func Register(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := providers[p.Type()]; exists {
		panic(fmt.Sprintf("memory: provider type %q already registered", p.Type()))
	}
	providers[p.Type()] = p
}

// Get returns the provider for the given type, or nil if not registered.
func Get(providerType string) Provider {
	mu.RLock()
	defer mu.RUnlock()
	return providers[providerType]
}

// All returns every registered provider, sorted by type name.
func All() []Provider {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]Provider, 0, len(providers))
	for _, p := range providers {
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Type() < result[j].Type()
	})
	return result
}

// SupportsCategory returns all registered providers that support the given category.
func SupportsCategory(cat Category) []Provider {
	mu.RLock()
	defer mu.RUnlock()
	var result []Provider
	for _, p := range providers {
		for _, c := range p.SupportedCategories() {
			if c == cat {
				result = append(result, p)
				break
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Type() < result[j].Type()
	})
	return result
}

// ValidType returns true if the given type string is registered.
func ValidType(providerType string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := providers[providerType]
	return ok
}

// ValidTypeForCategory returns true if the type is registered and supports the category.
func ValidTypeForCategory(providerType string, cat Category) bool {
	mu.RLock()
	defer mu.RUnlock()
	p, ok := providers[providerType]
	if !ok {
		return false
	}
	for _, c := range p.SupportedCategories() {
		if c == cat {
			return true
		}
	}
	return false
}

// ValidateConfig validates a config block against the provider's JSON Schema.
func ValidateConfig(providerType string, configBlock map[string]interface{}) error {
	p := Get(providerType)
	if p == nil {
		return nil
	}
	return schema.Validate(p.ConfigSchema(), configBlock)
}
