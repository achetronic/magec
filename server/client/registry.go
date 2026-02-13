package client

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

// Register adds a client provider to the global registry. Called from init()
// in each provider package (e.g. client/telegram, client/direct).
func Register(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := providers[p.Type()]; exists {
		panic(fmt.Sprintf("client: provider type %q already registered", p.Type()))
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

// validateRequired walks the schema and checks required fields are present.
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

	// Handle oneOf: find the matching branch based on const values and validate its required fields.
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
			return fmt.Errorf("%s is required for client type %s", label, providerType)
		}
		if s, isStr := val.(string); isStr && s == "" {
			label := key
			if propSchema, ok := props[key].(Schema); ok {
				if title, ok := propSchema["title"].(string); ok {
					label = title
				}
			}
			return fmt.Errorf("%s is required for client type %s", label, providerType)
		}
	}
	return nil
}

// matchOneOf finds the first oneOf branch where all const constraints match the data.
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

// jsonEqual compares two values by their JSON representation.
func jsonEqual(a, b interface{}) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}
