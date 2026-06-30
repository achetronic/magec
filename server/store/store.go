// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

// Store manages agent, backend, and MCP configurations with JSON persistence.
// It maintains two copies of the data:
//   - data: expanded (env vars resolved, secrets decrypted) — used at runtime
//   - rawData: unexpanded (original ${VAR} references, secrets encrypted) — written to disk
//
// Mutations from the API update both copies identically. Data loaded from disk
// with ${VAR} references is preserved in rawData so that persist() never writes
// expanded secrets or API keys to the store file.
type Store struct {
	mu            sync.RWMutex
	data          StoreData
	rawData       StoreData
	filePath      string
	encryptionKey string

	// brokenSkillIDs holds the IDs of skills whose store.json
	// entry carries fields outside the canonical {id, slug} shape
	// (decision #29). Filled in at loadFromDisk time and consulted
	// by the Skill accessors so a degraded skill is invisible to
	// every downstream consumer (admin API, agent toolset wiring)
	// until the operator re-uploads it.
	//
	// We do NOT auto-migrate, auto-repair, or hide the broken
	// entries from disk: the operator gets a single Warn-level log
	// per broken skill at startup with a clear instruction, edits
	// store.json by hand, and re-uploads through the admin UI.
	//
	// The value is a short human reason (e.g. "legacy fields
	// present: instructions, name, references") used in the
	// startup log line.
	brokenSkillIDs map[string]string

	changeMu   sync.Mutex
	changeSubs []chan struct{}
}

// New creates a new Store. If filePath is non-empty and the file exists, it loads from it.
// The encryptionKey is used to encrypt/decrypt secret values at rest. If empty, secrets are stored in cleartext.
func New(filePath string, encryptionKey string) (*Store, error) {
	defaults := StoreData{
		Backends:        []BackendDefinition{},
		MemoryProviders: []MemoryProvider{},
		MCPServers:      []MCPServer{},
		Skills:          []Skill{},
		Agents:          []AgentDefinition{},
		Clients:         []ClientDefinition{},
		Flows:           []FlowDefinition{},
		Commands:        []Command{},
		Secrets:         []Secret{},
	}
	s := &Store{
		filePath:      filePath,
		encryptionKey: encryptionKey,
		data:          defaults,
		rawData:       defaults,
	}

	if filePath != "" {
		if _, err := os.Stat(filePath); err == nil {
			if err := s.loadFromDisk(); err != nil {
				return nil, fmt.Errorf("failed to load store from %s: %w", filePath, err)
			}
		}
	}

	return s, nil
}

// OnChange returns a channel that receives a signal whenever the store is mutated.
// Multiple subscribers are supported. The channel is buffered (size 1) so a slow
// consumer won't block writers — at most one pending notification is kept.
func (s *Store) OnChange() <-chan struct{} {
	ch := make(chan struct{}, 1)
	s.changeMu.Lock()
	s.changeSubs = append(s.changeSubs, ch)
	s.changeMu.Unlock()
	return ch
}

// DataDir returns the directory containing the store file and related data.
func (s *Store) DataDir() string {
	return filepath.Dir(s.filePath)
}

// Data returns a copy of the current (expanded) store data for runtime use.
func (s *Store) Data() StoreData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

// RawData returns a copy of the unexpanded store data (with $VAR references
// intact). Safe for API responses — never contains resolved secrets.
func (s *Store) RawData() StoreData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rawData
}

// Reload re-reads the store file from disk, replacing all in-memory data.
func (s *Store) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadFromDisk()
}

// expandStruct takes any struct, marshals it to JSON, expands environment
// variables, and unmarshals back. This mirrors what loadFromDisk does for the
// entire store file, but scoped to a single entity.
func expandStruct[T any](v T) T {
	data, err := json.Marshal(v)
	if err != nil {
		return v
	}
	expanded := os.ExpandEnv(string(data))
	var out T
	if err := json.Unmarshal([]byte(expanded), &out); err != nil {
		return v
	}
	return out
}

// --- Settings ---

// GetSettings returns the current global settings.
func (s *Store) GetSettings() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Settings
}

// UpdateSettings replaces the global settings and persists.
func (s *Store) UpdateSettings(settings Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Settings = settings
	s.rawData.Settings = settings
	return s.persist()
}

// ResolveTemporaryDir returns the absolute path that callers must use when
// writing transient files. It is the single source of truth for that path:
// any subsystem that needs a workdir-style location (e.g. the export_artifact
// tool) must call this method and obey what it returns. When the operator
// has not configured Settings.TemporaryDir, it falls back to os.TempDir().
// Callers must not duplicate this fallback elsewhere.
func (s *Store) ResolveTemporaryDir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.data.Settings.TemporaryDir != "" {
		return s.data.Settings.TemporaryDir
	}
	return os.TempDir()
}

// --- Backends ---

func (s *Store) ListBackends() []BackendDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]BackendDefinition, len(s.data.Backends))
	copy(result, s.data.Backends)
	return result
}

func (s *Store) GetBackend(id string) (BackendDefinition, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.data.Backends {
		if b.ID == id {
			return b, true
		}
	}
	return BackendDefinition{}, false
}

func (s *Store) ListRawBackends() []BackendDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]BackendDefinition, len(s.rawData.Backends))
	copy(result, s.rawData.Backends)
	return result
}

func (s *Store) GetRawBackend(id string) (BackendDefinition, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.rawData.Backends {
		if b.ID == id {
			return b, true
		}
	}
	return BackendDefinition{}, false
}

func (s *Store) CreateBackend(b BackendDefinition) (BackendDefinition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b.ID = generateID()
	s.data.Backends = append(s.data.Backends, expandStruct(b))
	s.rawData.Backends = append(s.rawData.Backends, b)
	return b, s.persist()
}

func (s *Store) UpdateBackend(id string, b BackendDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Backends {
		if existing.ID == id {
			b.ID = id
			s.data.Backends[i] = expandStruct(b)
			s.rawData.Backends[i] = b
			return s.persist()
		}
	}
	return fmt.Errorf("backend %q not found", id)
}

func (s *Store) DeleteBackend(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Backends {
		if existing.ID == id {
			s.data.Backends = append(s.data.Backends[:i], s.data.Backends[i+1:]...)
			s.rawData.Backends = append(s.rawData.Backends[:i], s.rawData.Backends[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("backend %q not found", id)
}

// --- Memory Providers ---

// --- Memory Providers ---

func (s *Store) ListMemoryProviders() []MemoryProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]MemoryProvider, len(s.data.MemoryProviders))
	copy(result, s.data.MemoryProviders)
	return result
}

func (s *Store) GetMemoryProvider(id string) (MemoryProvider, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, m := range s.data.MemoryProviders {
		if m.ID == id {
			return m, true
		}
	}
	return MemoryProvider{}, false
}

// ListRawMemoryProviders returns providers with original $VAR references
// intact (from rawData). Safe for API responses.
func (s *Store) ListRawMemoryProviders() []MemoryProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]MemoryProvider, len(s.rawData.MemoryProviders))
	copy(result, s.rawData.MemoryProviders)
	return result
}

// GetRawMemoryProvider returns a single provider with original $VAR references
// intact (from rawData). Safe for API responses.
func (s *Store) GetRawMemoryProvider(id string) (MemoryProvider, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, m := range s.rawData.MemoryProviders {
		if m.ID == id {
			return m, true
		}
	}
	return MemoryProvider{}, false
}

func (s *Store) CreateMemoryProvider(m MemoryProvider) (MemoryProvider, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m.ID = generateID()
	s.data.MemoryProviders = append(s.data.MemoryProviders, expandStruct(m))
	s.rawData.MemoryProviders = append(s.rawData.MemoryProviders, m)
	return m, s.persist()
}

func (s *Store) UpdateMemoryProvider(id string, m MemoryProvider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.MemoryProviders {
		if existing.ID == id {
			m.ID = id
			s.data.MemoryProviders[i] = expandStruct(m)
			s.rawData.MemoryProviders[i] = m
			return s.persist()
		}
	}
	return fmt.Errorf("memory provider %q not found", id)
}

func (s *Store) DeleteMemoryProvider(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.MemoryProviders {
		if existing.ID == id {
			s.data.MemoryProviders = append(s.data.MemoryProviders[:i], s.data.MemoryProviders[i+1:]...)
			s.rawData.MemoryProviders = append(s.rawData.MemoryProviders[:i], s.rawData.MemoryProviders[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("memory provider %q not found", id)
}

// --- MCP Servers (global) ---

func (s *Store) ListMCPServers() []MCPServer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]MCPServer, len(s.data.MCPServers))
	copy(result, s.data.MCPServers)
	return result
}

func (s *Store) GetMCPServer(id string) (MCPServer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, m := range s.data.MCPServers {
		if m.ID == id {
			return m, true
		}
	}
	return MCPServer{}, false
}

func (s *Store) ListRawMCPServers() []MCPServer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]MCPServer, len(s.rawData.MCPServers))
	copy(result, s.rawData.MCPServers)
	return result
}

func (s *Store) GetRawMCPServer(id string) (MCPServer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, m := range s.rawData.MCPServers {
		if m.ID == id {
			return m, true
		}
	}
	return MCPServer{}, false
}

func (s *Store) CreateMCPServer(m MCPServer) (MCPServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m.ID = generateID()
	s.data.MCPServers = append(s.data.MCPServers, expandStruct(m))
	s.rawData.MCPServers = append(s.rawData.MCPServers, m)
	return m, s.persist()
}

func (s *Store) UpdateMCPServer(id string, m MCPServer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.MCPServers {
		if existing.ID == id {
			m.ID = id
			s.data.MCPServers[i] = expandStruct(m)
			s.rawData.MCPServers[i] = m
			return s.persist()
		}
	}
	return fmt.Errorf("MCP server %q not found", id)
}

func (s *Store) DeleteMCPServer(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.MCPServers {
		if existing.ID == id {
			s.data.MCPServers = append(s.data.MCPServers[:i], s.data.MCPServers[i+1:]...)
			s.rawData.MCPServers = append(s.rawData.MCPServers[:i], s.rawData.MCPServers[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("MCP server %q not found", id)
}

// --- Agents ---

func (s *Store) ListAgents() []AgentDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]AgentDefinition, len(s.data.Agents))
	copy(result, s.data.Agents)
	return result
}

func (s *Store) GetAgent(id string) (AgentDefinition, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.data.Agents {
		if a.ID == id {
			return a, true
		}
	}
	return AgentDefinition{}, false
}

func (s *Store) ListRawAgents() []AgentDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]AgentDefinition, len(s.rawData.Agents))
	copy(result, s.rawData.Agents)
	return result
}

func (s *Store) GetRawAgent(id string) (AgentDefinition, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.rawData.Agents {
		if a.ID == id {
			return a, true
		}
	}
	return AgentDefinition{}, false
}

func (s *Store) CreateAgent(a AgentDefinition) (AgentDefinition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	a.ID = generateID()
	if a.MCPServers == nil {
		a.MCPServers = []string{}
	}
	if a.Skills == nil {
		a.Skills = []string{}
	}
	s.data.Agents = append(s.data.Agents, expandStruct(a))
	s.rawData.Agents = append(s.rawData.Agents, a)
	return a, s.persist()
}

func (s *Store) UpdateAgent(id string, a AgentDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Agents {
		if existing.ID == id {
			a.ID = id
			if a.MCPServers == nil {
				a.MCPServers = []string{}
			}
			if a.Skills == nil {
				a.Skills = []string{}
			}
			s.data.Agents[i] = expandStruct(a)
			s.rawData.Agents[i] = a
			return s.persist()
		}
	}
	return fmt.Errorf("agent %q not found", id)
}

func (s *Store) DeleteAgent(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Agents {
		if existing.ID == id {
			s.data.Agents = append(s.data.Agents[:i], s.data.Agents[i+1:]...)
			s.rawData.Agents = append(s.rawData.Agents[:i], s.rawData.Agents[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("agent %q not found", id)
}

// --- Agent MCP linking ---

// LinkAgentMCP adds an MCP server reference to an agent.
func (s *Store) LinkAgentMCP(agentID, mcpID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mcpExists := false
	for _, m := range s.data.MCPServers {
		if m.ID == mcpID {
			mcpExists = true
			break
		}
	}
	if !mcpExists {
		return fmt.Errorf("MCP server %q not found", mcpID)
	}

	for i, a := range s.data.Agents {
		if a.ID == agentID {
			if slices.Contains(a.MCPServers, mcpID) {
				return fmt.Errorf("MCP %q already linked to agent %q", mcpID, agentID)
			}
			s.data.Agents[i].MCPServers = append(s.data.Agents[i].MCPServers, mcpID)
			s.rawData.Agents[i].MCPServers = append(s.rawData.Agents[i].MCPServers, mcpID)
			return s.persist()
		}
	}
	return fmt.Errorf("agent %q not found", agentID)
}

// UnlinkAgentMCP removes an MCP server reference from an agent.
func (s *Store) UnlinkAgentMCP(agentID, mcpID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, a := range s.data.Agents {
		if a.ID == agentID {
			idx := slices.Index(a.MCPServers, mcpID)
			if idx == -1 {
				return fmt.Errorf("MCP %q not linked to agent %q", mcpID, agentID)
			}
			s.data.Agents[i].MCPServers = slices.Delete(a.MCPServers, idx, idx+1)
			s.rawData.Agents[i].MCPServers = slices.Delete(s.rawData.Agents[i].MCPServers, idx, idx+1)
			return s.persist()
		}
	}
	return fmt.Errorf("agent %q not found", agentID)
}

// ResolveAgentMCPs returns the full MCPServer definitions for an agent's linked MCPs.
func (s *Store) ResolveAgentMCPs(agentID string) ([]MCPServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var agentMCPIDs []string
	found := false
	for _, a := range s.data.Agents {
		if a.ID == agentID {
			agentMCPIDs = a.MCPServers
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("agent %q not found", agentID)
	}

	mcpMap := make(map[string]MCPServer, len(s.data.MCPServers))
	for _, m := range s.data.MCPServers {
		mcpMap[m.ID] = m
	}

	result := make([]MCPServer, 0, len(agentMCPIDs))
	for _, id := range agentMCPIDs {
		if m, ok := mcpMap[id]; ok {
			result = append(result, m)
		}
	}
	return result, nil
}

// ResolveRawAgentMCPs returns the unexpanded MCPServer definitions for an
// agent's linked MCPs. Safe for API responses.
func (s *Store) ResolveRawAgentMCPs(agentID string) ([]MCPServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var agentMCPIDs []string
	found := false
	for _, a := range s.rawData.Agents {
		if a.ID == agentID {
			agentMCPIDs = a.MCPServers
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("agent %q not found", agentID)
	}

	mcpMap := make(map[string]MCPServer, len(s.rawData.MCPServers))
	for _, m := range s.rawData.MCPServers {
		mcpMap[m.ID] = m
	}

	result := make([]MCPServer, 0, len(agentMCPIDs))
	for _, id := range agentMCPIDs {
		if m, ok := mcpMap[id]; ok {
			result = append(result, m)
		}
	}
	return result, nil
}

// --- Clients ---

// generateToken creates a random API token with the "mgc_" prefix.
func generateToken() string {
	b := make([]byte, 20)
	rand.Read(b)
	return "mgc_" + hex.EncodeToString(b)
}

func (s *Store) ListClients() []ClientDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]ClientDefinition, len(s.data.Clients))
	copy(result, s.data.Clients)
	return result
}

func (s *Store) GetClient(id string) (ClientDefinition, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.data.Clients {
		if c.ID == id {
			return c, true
		}
	}
	return ClientDefinition{}, false
}

func (s *Store) ListRawClients() []ClientDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]ClientDefinition, len(s.rawData.Clients))
	copy(result, s.rawData.Clients)
	return result
}

func (s *Store) GetRawClient(id string) (ClientDefinition, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.rawData.Clients {
		if c.ID == id {
			return c, true
		}
	}
	return ClientDefinition{}, false
}

// GetClientByToken looks up a client by its API token. Used by the auth middleware.
func (s *Store) GetClientByToken(token string) (ClientDefinition, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.data.Clients {
		if c.Token == token {
			return c, true
		}
	}
	return ClientDefinition{}, false
}

func (s *Store) CreateClient(c ClientDefinition) (ClientDefinition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c.ID = generateID()
	c.Token = generateToken()
	if c.AllowedAgents == nil {
		c.AllowedAgents = []string{}
	}
	s.data.Clients = append(s.data.Clients, expandStruct(c))
	s.rawData.Clients = append(s.rawData.Clients, c)
	return c, s.persist()
}

func (s *Store) UpdateClient(id string, c ClientDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Clients {
		if existing.ID == id {
			c.ID = id
			c.Token = existing.Token
			if c.AllowedAgents == nil {
				c.AllowedAgents = []string{}
			}
			s.data.Clients[i] = expandStruct(c)
			s.rawData.Clients[i] = c
			return s.persist()
		}
	}
	return fmt.Errorf("client %q not found", id)
}

func (s *Store) DeleteClient(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Clients {
		if existing.ID == id {
			s.data.Clients = append(s.data.Clients[:i], s.data.Clients[i+1:]...)
			s.rawData.Clients = append(s.rawData.Clients[:i], s.rawData.Clients[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("client %q not found", id)
}

// RegenerateClientToken replaces a client's API token with a new random one.
func (s *Store) RegenerateClientToken(id string) (ClientDefinition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Clients {
		if existing.ID == id {
			newToken := generateToken()
			s.data.Clients[i].Token = newToken
			s.rawData.Clients[i].Token = newToken
			return s.rawData.Clients[i], s.persist()
		}
	}
	return ClientDefinition{}, fmt.Errorf("client %q not found", id)
}

// --- Flows ---

func (s *Store) ListFlows() []FlowDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]FlowDefinition, len(s.data.Flows))
	copy(result, s.data.Flows)
	return result
}

func (s *Store) GetFlow(id string) (FlowDefinition, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, f := range s.data.Flows {
		if f.ID == id {
			return f, true
		}
	}
	return FlowDefinition{}, false
}

func (s *Store) ListRawFlows() []FlowDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]FlowDefinition, len(s.rawData.Flows))
	copy(result, s.rawData.Flows)
	return result
}

func (s *Store) GetRawFlow(id string) (FlowDefinition, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, f := range s.rawData.Flows {
		if f.ID == id {
			return f, true
		}
	}
	return FlowDefinition{}, false
}

func (s *Store) CreateFlow(f FlowDefinition) (FlowDefinition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	f.ID = generateID()
	s.data.Flows = append(s.data.Flows, expandStruct(f))
	s.rawData.Flows = append(s.rawData.Flows, f)
	return f, s.persist()
}

func (s *Store) UpdateFlow(id string, f FlowDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Flows {
		if existing.ID == id {
			f.ID = id
			s.data.Flows[i] = expandStruct(f)
			s.rawData.Flows[i] = f
			return s.persist()
		}
	}
	return fmt.Errorf("flow %q not found", id)
}

func (s *Store) DeleteFlow(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Flows {
		if existing.ID == id {
			s.data.Flows = append(s.data.Flows[:i], s.data.Flows[i+1:]...)
			s.rawData.Flows = append(s.rawData.Flows[:i], s.rawData.Flows[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("flow %q not found", id)
}

// --- Skills ---
//
// Skills are tracked in the store with the bare minimum needed to keep
// agent links stable across UI renames: an immutable UUID and the
// on-disk slug (= directory name under data/skills/, also the SKILL.md
// frontmatter `name`). Everything else (name, description, instructions,
// resources) lives on disk inside data/skills/{slug}/ and is read at
// admin-API GET time. See decision #29.
//
// Skills whose store.json entry still carries legacy fields
// (`instructions`, `references`, `name`, `description`) are filtered
// out of every read path here. They show up nowhere — not in the
// admin UI list, not in the runtime agent toolset — until the
// operator removes the legacy entry from store.json by hand and
// re-uploads the skill via the admin UI. The startup log already
// pinpointed the offending IDs by then.

// skillIsBrokenLocked is a cheap predicate over the broken-skill
// set populated by loadFromDisk. The caller must already hold s.mu
// (read or write); we don't re-lock here to avoid lock churn in
// tight loops like ListSkills.
func (s *Store) skillIsBrokenLocked(id string) bool {
	if s.brokenSkillIDs == nil {
		return false
	}
	_, broken := s.brokenSkillIDs[id]
	return broken
}

func (s *Store) ListSkills() []Skill {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Skill, 0, len(s.data.Skills))
	for _, sk := range s.data.Skills {
		if s.skillIsBrokenLocked(sk.ID) {
			continue
		}
		result = append(result, sk)
	}
	return result
}

func (s *Store) GetSkill(id string) (Skill, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.skillIsBrokenLocked(id) {
		return Skill{}, false
	}
	for _, sk := range s.data.Skills {
		if sk.ID == id {
			return sk, true
		}
	}
	return Skill{}, false
}

// GetSkillBySlug returns the skill whose on-disk directory matches slug.
// The admin upload handler uses it to detect re-uploads (same slug ->
// existing skill) versus brand-new skills. Broken entries are hidden
// here too — if the operator re-uploads a slug that collided with a
// legacy entry, the upload behaves as a clean create.
func (s *Store) GetSkillBySlug(slug string) (Skill, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sk := range s.data.Skills {
		if sk.Slug == slug && !s.skillIsBrokenLocked(sk.ID) {
			return sk, true
		}
	}
	return Skill{}, false
}

func (s *Store) ListRawSkills() []Skill {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Skill, 0, len(s.rawData.Skills))
	for _, sk := range s.rawData.Skills {
		if s.skillIsBrokenLocked(sk.ID) {
			continue
		}
		result = append(result, sk)
	}
	return result
}

func (s *Store) GetRawSkill(id string) (Skill, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.skillIsBrokenLocked(id) {
		return Skill{}, false
	}
	for _, sk := range s.rawData.Skills {
		if sk.ID == id {
			return sk, true
		}
	}
	return Skill{}, false
}

// CreateSkill registers a new skill record. Slug must be unique and is
// validated by the caller (the upload handler) — this method only
// guarantees the in-memory invariant.
func (s *Store) CreateSkill(sk Skill) (Skill, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sk.Slug == "" {
		return Skill{}, fmt.Errorf("skill slug is required")
	}
	for _, existing := range s.data.Skills {
		if existing.Slug == sk.Slug {
			return Skill{}, fmt.Errorf("skill slug %q already exists", sk.Slug)
		}
	}

	sk.ID = generateID()
	s.data.Skills = append(s.data.Skills, expandStruct(sk))
	s.rawData.Skills = append(s.rawData.Skills, sk)
	return sk, s.persist()
}

// UpdateSkill replaces the slug of an existing skill (keeping its ID).
// The admin layer calls this after re-uploading a package whose
// frontmatter `name` differs from the previous slug — it has already
// renamed the directory on disk before calling.
func (s *Store) UpdateSkill(id string, sk Skill) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sk.Slug == "" {
		return fmt.Errorf("skill slug is required")
	}
	for _, existing := range s.data.Skills {
		if existing.ID != id && existing.Slug == sk.Slug {
			return fmt.Errorf("skill slug %q already in use", sk.Slug)
		}
	}
	for i, existing := range s.data.Skills {
		if existing.ID == id {
			sk.ID = id
			s.data.Skills[i] = expandStruct(sk)
			s.rawData.Skills[i] = sk
			return s.persist()
		}
	}
	return fmt.Errorf("skill %q not found", id)
}

// DeleteSkill removes the store record AND the on-disk directory that
// backs it. Failure to remove the directory is logged but not returned
// — the store record is the authoritative ledger and we don't want to
// strand it on a transient filesystem error.
func (s *Store) DeleteSkill(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Skills {
		if existing.ID == id {
			slug := existing.Slug
			s.data.Skills = append(s.data.Skills[:i], s.data.Skills[i+1:]...)
			s.rawData.Skills = append(s.rawData.Skills[:i], s.rawData.Skills[i+1:]...)
			if err := s.persist(); err != nil {
				return err
			}
			if slug != "" {
				_ = os.RemoveAll(s.SkillDir(slug))
			}
			return nil
		}
	}
	return fmt.Errorf("skill %q not found", id)
}

// SkillsDir returns the root directory that holds every skill package.
// Callers typically wrap it with os.DirFS to feed ADK's skilltoolset.
func (s *Store) SkillsDir() string {
	return filepath.Join(filepath.Dir(s.filePath), "skills")
}

// SkillDir returns the absolute on-disk directory for a single skill,
// keyed by slug. The directory layout is the one ADK's skilltoolset
// expects: SKILL.md at the root, optional references/, assets/,
// scripts/ subtrees.
func (s *Store) SkillDir(slug string) string {
	return filepath.Join(s.SkillsDir(), slug)
}

// --- Commands ---

func (s *Store) ListCommands() []Command {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Command, len(s.data.Commands))
	copy(result, s.data.Commands)
	return result
}

func (s *Store) GetCommand(id string) (Command, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.data.Commands {
		if c.ID == id {
			return c, true
		}
	}
	return Command{}, false
}

func (s *Store) ListRawCommands() []Command {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Command, len(s.rawData.Commands))
	copy(result, s.rawData.Commands)
	return result
}

func (s *Store) GetRawCommand(id string) (Command, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.rawData.Commands {
		if c.ID == id {
			return c, true
		}
	}
	return Command{}, false
}

func (s *Store) CreateCommand(c Command) (Command, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	c.ID = generateID()
	s.data.Commands = append(s.data.Commands, expandStruct(c))
	s.rawData.Commands = append(s.rawData.Commands, c)
	return c, s.persist()
}

func (s *Store) UpdateCommand(id string, c Command) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Commands {
		if existing.ID == id {
			c.ID = id
			s.data.Commands[i] = expandStruct(c)
			s.rawData.Commands[i] = c
			return s.persist()
		}
	}
	return fmt.Errorf("command %q not found", id)
}

func (s *Store) DeleteCommand(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Commands {
		if existing.ID == id {
			s.data.Commands = append(s.data.Commands[:i], s.data.Commands[i+1:]...)
			s.rawData.Commands = append(s.rawData.Commands[:i], s.rawData.Commands[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("command %q not found", id)
}

// --- Persistence ---

// --- Secrets ---

func (s *Store) ListSecrets() []Secret {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Secret, len(s.data.Secrets))
	copy(result, s.data.Secrets)
	return result
}

func (s *Store) GetSecret(id string) (Secret, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sec := range s.data.Secrets {
		if sec.ID == id {
			return sec, true
		}
	}
	return Secret{}, false
}

func (s *Store) CreateSecret(sec Secret) (Secret, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.data.Secrets {
		if existing.Key == sec.Key {
			return Secret{}, fmt.Errorf("secret with key %q already exists", sec.Key)
		}
	}

	sec.ID = generateID()

	rawSec := sec
	if s.encryptionKey != "" && sec.Value != "" && !isEncrypted(sec.Value) {
		enc, err := encryptValue(sec.Value, s.encryptionKey)
		if err != nil {
			return Secret{}, fmt.Errorf("encrypt: %w", err)
		}
		rawSec.Value = enc
	}

	if sec.Key != "" && sec.Value != "" {
		os.Setenv(sec.Key, sec.Value)
	}

	s.rawData.Secrets = append(s.rawData.Secrets, rawSec)
	s.reExpandDataLocked()
	return sec, s.persist()
}

func (s *Store) UpdateSecret(id string, sec Secret) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.data.Secrets {
		if existing.Key == sec.Key && existing.ID != id {
			return fmt.Errorf("secret with key %q already exists", sec.Key)
		}
	}

	for i, existing := range s.data.Secrets {
		if existing.ID == id {
			sec.ID = id
			if sec.Value == "" {
				sec.Value = existing.Value
			}

			rawSec := sec
			if s.encryptionKey != "" && rawSec.Value != "" && !isEncrypted(rawSec.Value) {
				enc, err := encryptValue(rawSec.Value, s.encryptionKey)
				if err != nil {
					return fmt.Errorf("encrypt: %w", err)
				}
				rawSec.Value = enc
			}

			if sec.Key != "" && sec.Value != "" {
				os.Setenv(sec.Key, sec.Value)
			}
			if existing.Key != "" && existing.Key != sec.Key {
				os.Unsetenv(existing.Key)
			}

			s.rawData.Secrets[i] = rawSec
			s.reExpandDataLocked()
			return s.persist()
		}
	}
	return fmt.Errorf("secret %q not found", id)
}

func (s *Store) DeleteSecret(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Secrets {
		if existing.ID == id {
			if existing.Key != "" {
				os.Unsetenv(existing.Key)
			}
			s.rawData.Secrets = append(s.rawData.Secrets[:i], s.rawData.Secrets[i+1:]...)
			s.reExpandDataLocked()
			return s.persist()
		}
	}
	return fmt.Errorf("secret %q not found", id)
}

// reExpandDataLocked re-evaluates all environment variables across the store's
// raw data and updates the expanded data representation. Must be called while holding s.mu.Lock().
func (s *Store) reExpandDataLocked() {
	data, err := json.Marshal(s.rawData)
	if err != nil {
		return
	}
	expanded := os.ExpandEnv(string(data))
	var storeData StoreData
	if err := json.Unmarshal([]byte(expanded), &storeData); err != nil {
		return
	}
	initSlices(&storeData)
	s.data = storeData
}

// persist writes the current store data to disk as formatted JSON and
// notifies all change subscribers.
func (s *Store) persist() error {
	if s.filePath == "" {
		return nil
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	data, err := json.MarshalIndent(s.rawData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal store data: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write store file: %w", err)
	}

	s.notifyChange()
	return nil
}

// notifyChange sends a non-blocking signal to all OnChange subscribers.
func (s *Store) notifyChange() {
	s.changeMu.Lock()
	defer s.changeMu.Unlock()
	for _, ch := range s.changeSubs {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// loadFromDisk reads the store file, extracts secrets and injects them as
// environment variables, then expands all env vars and unmarshals the final data.
// This two-pass approach lets secrets reference each other or be used in other fields.
func (s *Store) loadFromDisk() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	data = migrateTTSConfig(data)

	// Detect skills whose store.json entry carries fields outside
	// the canonical {id, slug} shape. We do not migrate them: the
	// operator removes the legacy entry from store.json and re-
	// uploads the skill through the admin UI (decision #29). The
	// detection runs on the raw bytes so we can inspect the legacy
	// fields before the strict struct unmarshal silently drops
	// them.
	s.brokenSkillIDs = detectBrokenSkills(data)
	for id, reason := range s.brokenSkillIDs {
		slog.Warn("skill in legacy format and will be ignored — re-upload through the admin UI",
			"id", id, "reason", reason,
			"action", "remove the entry from data/store.json and re-upload the skill via Skills → Upload Skill")
	}

	var raw StoreData
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for _, sec := range raw.Secrets {
		if sec.Key != "" && sec.Value != "" {
			val := sec.Value
			if s.encryptionKey != "" && isEncrypted(val) {
				decrypted, err := decryptValue(val, s.encryptionKey)
				if err != nil {
					return fmt.Errorf("failed to decrypt secret %q: %w", sec.Key, err)
				}
				val = decrypted
			}
			os.Setenv(sec.Key, val)
		}
	}

	expanded := os.ExpandEnv(string(data))

	var storeData StoreData
	if err := json.Unmarshal([]byte(expanded), &storeData); err != nil {
		return err
	}

	initSlices(&storeData)
	initSlices(&raw)

	s.data = storeData
	s.rawData = raw

	return nil
}

func initSlices(sd *StoreData) {
	if sd.Backends == nil {
		sd.Backends = []BackendDefinition{}
	}
	if sd.MemoryProviders == nil {
		sd.MemoryProviders = []MemoryProvider{}
	}
	if sd.MCPServers == nil {
		sd.MCPServers = []MCPServer{}
	}
	if sd.Skills == nil {
		sd.Skills = []Skill{}
	}
	if sd.Agents == nil {
		sd.Agents = []AgentDefinition{}
	}
	if sd.Clients == nil {
		sd.Clients = []ClientDefinition{}
	}
	if sd.Flows == nil {
		sd.Flows = []FlowDefinition{}
	}
	if sd.Commands == nil {
		sd.Commands = []Command{}
	}
	if sd.Secrets == nil {
		sd.Secrets = []Secret{}
	}
}

// migrateTTSConfig migrates old store formats to the typed config pattern:
// - tts.speed (top-level) → tts.config.openai.speed
// - tts.config.languageCode / temperature / stylePrompt (flat) → tts.config.gemini.*
// Operates on raw JSON bytes before unmarshal so no data is lost.
//
// TEMPORARY: introduced in v0.18.0. Remove after v0.20.0 — by then all
// installations will have been migrated on first load.
func migrateTTSConfig(data []byte) []byte {
	var store map[string]interface{}
	if err := json.Unmarshal(data, &store); err != nil {
		return data
	}

	agents, ok := store["agents"].([]interface{})
	if !ok {
		return data
	}

	changed := false
	for _, a := range agents {
		agent, ok := a.(map[string]interface{})
		if !ok {
			continue
		}
		tts, ok := agent["tts"].(map[string]interface{})
		if !ok {
			continue
		}

		if speed, ok := tts["speed"]; ok {
			delete(tts, "speed")
			cfg, _ := tts["config"].(map[string]interface{})
			if cfg == nil {
				cfg = map[string]interface{}{}
			}
			openai, _ := cfg["openai"].(map[string]interface{})
			if openai == nil {
				openai = map[string]interface{}{}
			}
			if _, exists := openai["speed"]; !exists {
				openai["speed"] = speed
			}
			cfg["openai"] = openai
			tts["config"] = cfg
			changed = true
		}

		cfg, _ := tts["config"].(map[string]interface{})
		if cfg == nil {
			continue
		}
		gemini := map[string]interface{}{}
		for _, key := range []string{"languageCode", "temperature", "stylePrompt"} {
			if v, ok := cfg[key]; ok {
				gemini[key] = v
				delete(cfg, key)
				changed = true
			}
		}
		if len(gemini) > 0 {
			existing, _ := cfg["gemini"].(map[string]interface{})
			if existing == nil {
				cfg["gemini"] = gemini
			} else {
				for k, v := range gemini {
					if _, exists := existing[k]; !exists {
						existing[k] = v
					}
				}
			}
		}
	}

	if !changed {
		return data
	}

	out, err := json.Marshal(store)
	if err != nil {
		return data
	}
	return out
}

// allowedSkillKeys is the canonical store-side shape of a skill
// entry as of decision #29. Any other key inside an entry tells us
// the operator is on the legacy format and the skill must be
// quarantined until they re-upload it.
var allowedSkillKeys = map[string]struct{}{
	"id":   {},
	"slug": {},
}

// detectBrokenSkills walks the raw store.json bytes BEFORE the
// strict struct unmarshal and reports the IDs of any skill entry
// that carries fields the new schema doesn't model. The returned
// map is keyed by skill ID and the value is a short human reason
// — both are surfaced through a single Warn-level log per broken
// skill at startup so the operator knows exactly what to fix.
//
// We do this on raw bytes (not on the parsed StoreData) because
// the JSON decoder silently drops unknown fields, so by the time
// we have the struct we'd have lost the evidence that the entry
// was legacy. Returns an empty (non-nil) map on success and a
// nil-keyed empty map when the store has no skills section.
func detectBrokenSkills(raw []byte) map[string]string {
	out := map[string]string{}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return out
	}
	skillsAny, ok := doc["skills"].([]any)
	if !ok {
		return out
	}

	for _, item := range skillsAny {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id, _ := entry["id"].(string)
		if id == "" {
			continue
		}
		var legacyKeys []string
		for k, v := range entry {
			if _, allowed := allowedSkillKeys[k]; allowed {
				continue
			}
			if isEmptyJSONValue(v) {
				continue
			}
			legacyKeys = append(legacyKeys, k)
		}
		if len(legacyKeys) > 0 {
			slices.Sort(legacyKeys)
			out[id] = "legacy fields present: " + strings.Join(legacyKeys, ", ")
		}
	}
	return out
}

// isEmptyJSONValue treats nil, empty strings, empty arrays and
// empty objects as "not really there" so a skill JSON-marshalled
// with `omitempty` left behind doesn't trip the detector. The
// detector should only fire on entries that carry actual legacy
// data — instructions, references, descriptions etc.
func isEmptyJSONValue(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return x == ""
	case []any:
		return len(x) == 0
	case map[string]any:
		return len(x) == 0
	default:
		return false
	}
}

// IsSkillBroken reports whether a skill ID was flagged as
// degraded at load time. Used by the Skill accessors to filter
// the list/get responses, and by tests.
func (s *Store) IsSkillBroken(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, broken := s.brokenSkillIDs[id]
	return broken
}
