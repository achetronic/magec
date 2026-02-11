package client

import (
	"fmt"
	"sort"
	"sync"
)

var (
	mu        sync.RWMutex
	providers = map[string]Provider{}
)

func Register(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := providers[p.Type()]; exists {
		panic(fmt.Sprintf("client: provider type %q already registered", p.Type()))
	}
	providers[p.Type()] = p
}

func Get(providerType string) Provider {
	mu.RLock()
	defer mu.RUnlock()
	return providers[providerType]
}

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

func ValidType(providerType string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := providers[providerType]
	return ok
}

func ValidateRequired(providerType string, configBlock map[string]interface{}) error {
	p := Get(providerType)
	if p == nil {
		return nil
	}
	for _, f := range p.ConfigFields() {
		if !f.Required {
			continue
		}
		val, ok := configBlock[f.Key]
		if !ok {
			return fmt.Errorf("%s is required for client type %s", f.Key, providerType)
		}
		if s, isStr := val.(string); isStr && s == "" {
			return fmt.Errorf("%s is required for client type %s", f.Key, providerType)
		}
	}
	return nil
}
