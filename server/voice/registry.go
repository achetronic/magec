// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package voice

import (
	"fmt"
	"sort"
	"sync"
)

var (
	mu        sync.RWMutex
	providers = map[string]Provider{}
)

// Register adds a voice provider to the global registry. Called from init()
// in each provider package (e.g. voice/openai, voice/gemini).
func Register(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	if _, exists := providers[p.Type()]; exists {
		panic(fmt.Sprintf("voice: provider type %q already registered", p.Type()))
	}
	providers[p.Type()] = p
}

// Get returns the provider for the given backend type, or nil if not registered.
func Get(backendType string) Provider {
	mu.RLock()
	defer mu.RUnlock()
	return providers[backendType]
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
