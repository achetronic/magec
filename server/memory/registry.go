package memory

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
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
// It checks that all required properties are present and non-empty.
func ValidateConfig(providerType string, configBlock map[string]interface{}) error {
	p := Get(providerType)
	if p == nil {
		return nil
	}
	schema := p.ConfigSchema()
	if schema == nil {
		return nil
	}
	return validateRequired(providerType, schema, configBlock)
}

func validateRequired(providerType string, schema Schema, data map[string]interface{}) error {
	requiredRaw, ok := schema["required"]
	if !ok {
		return nil
	}
	required, ok := requiredRaw.([]string)
	if !ok {
		reqIface, ok := requiredRaw.([]interface{})
		if !ok {
			return nil
		}
		for _, v := range reqIface {
			if s, ok := v.(string); ok {
				required = append(required, s)
			}
		}
	}

	propsRaw, ok := schema["properties"]
	if !ok {
		return nil
	}
	props, ok := propsRaw.(Schema)
	if !ok {
		return nil
	}

	if oneOfRaw, ok := schema["oneOf"]; ok {
		if branches, ok := oneOfRaw.([]Schema); ok {
			branch := matchOneOf(branches, data)
			if branch != nil {
				if err := validateRequired(providerType, branch, data); err != nil {
					return err
				}
			}
		}
	}

	for _, key := range required {
		val, exists := data[key]
		if !exists {
			label := key
			if propSchema, ok := props[key].(Schema); ok {
				if title, ok := propSchema["title"].(string); ok {
					label = title
				}
			}
			return fmt.Errorf("%s is required for memory provider type %s", label, providerType)
		}
		if s, isStr := val.(string); isStr && s == "" {
			label := key
			if propSchema, ok := props[key].(Schema); ok {
				if title, ok := propSchema["title"].(string); ok {
					label = title
				}
			}
			return fmt.Errorf("%s is required for memory provider type %s", label, providerType)
		}
	}
	return nil
}

func matchOneOf(branches []Schema, data map[string]interface{}) Schema {
	for _, branch := range branches {
		propsRaw, ok := branch["properties"]
		if !ok {
			continue
		}
		props, ok := propsRaw.(Schema)
		if !ok {
			continue
		}
		match := true
		for key, propRaw := range props {
			propSchema, ok := propRaw.(Schema)
			if !ok {
				continue
			}
			constVal, hasConst := propSchema["const"]
			if !hasConst {
				continue
			}
			dataVal, exists := data[key]
			if !exists {
				match = false
				break
			}
			if !jsonEqual(constVal, dataVal) {
				match = false
				break
			}
		}
		if match {
			return branch
		}
	}
	return nil
}

func jsonEqual(a, b interface{}) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}
