package clients

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

// Register adds a client provider to the global registry. Called from init()
// in each provider package (e.g. clients/telegram, clients/direct).
func Register(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := providers[p.Type()]; exists {
		panic(fmt.Sprintf("clients: provider type %q already registered", p.Type()))
	}
	providers[p.Type()] = p
}

// Get returns the provider for the given type, or nil if not registered.
func Get(providerType string) Provider {
	mu.RLock()
	defer mu.RUnlock()
	return providers[providerType]
}

// All returns every registered provider, sorted alphabetically by type.
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

// ValidType returns true if a provider is registered for the given type string.
func ValidType(providerType string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := providers[providerType]
	return ok
}

// ValidateConfig validates a config block against the provider's JSON Schema.
func ValidateConfig(providerType string, configBlock map[string]interface{}) error {
	p := Get(providerType)
	if p == nil {
		return nil
	}
	return schema.Validate(p.ConfigSchema(), configBlock)
}
