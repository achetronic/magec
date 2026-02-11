package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

// Store manages agent, backend, and MCP configurations with JSON persistence.
type Store struct {
	mu       sync.RWMutex
	data     StoreData
	filePath string

	changeMu    sync.Mutex
	changeSubs  []chan struct{}
}

// New creates a new Store. If filePath is non-empty and the file exists, it loads from it.
func New(filePath string) (*Store, error) {
	s := &Store{
		filePath: filePath,
		data: StoreData{
			Backends:        []BackendDefinition{},
			MemoryProviders: []MemoryProvider{},
			MCPServers:      []MCPServer{},
			Agents:          []AgentDefinition{},
			CronJobs:        []CronJob{},
			Devices:         []Device{},
		},
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

// Data returns a copy of the current store data.
func (s *Store) Data() StoreData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

// --- Backends ---

func (s *Store) ListBackends() []BackendDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]BackendDefinition, len(s.data.Backends))
	copy(result, s.data.Backends)
	return result
}

func (s *Store) GetBackend(name string) (BackendDefinition, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.data.Backends {
		if b.Name == name {
			return b, true
		}
	}
	return BackendDefinition{}, false
}

func (s *Store) CreateBackend(b BackendDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.data.Backends {
		if existing.Name == b.Name {
			return fmt.Errorf("backend %q already exists", b.Name)
		}
	}
	s.data.Backends = append(s.data.Backends, b)
	return s.persist()
}

func (s *Store) UpdateBackend(name string, b BackendDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Backends {
		if existing.Name == name {
			b.Name = name
			s.data.Backends[i] = b
			return s.persist()
		}
	}
	return fmt.Errorf("backend %q not found", name)
}

func (s *Store) DeleteBackend(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Backends {
		if existing.Name == name {
			s.data.Backends = append(s.data.Backends[:i], s.data.Backends[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("backend %q not found", name)
}

// --- Memory Providers ---

func (s *Store) ListMemoryProviders() []MemoryProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]MemoryProvider, len(s.data.MemoryProviders))
	copy(result, s.data.MemoryProviders)
	return result
}

func (s *Store) GetMemoryProvider(name string) (MemoryProvider, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, m := range s.data.MemoryProviders {
		if m.Name == name {
			return m, true
		}
	}
	return MemoryProvider{}, false
}

func (s *Store) CreateMemoryProvider(m MemoryProvider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.data.MemoryProviders {
		if existing.Name == m.Name {
			return fmt.Errorf("memory provider %q already exists", m.Name)
		}
	}
	s.data.MemoryProviders = append(s.data.MemoryProviders, m)
	return s.persist()
}

func (s *Store) UpdateMemoryProvider(name string, m MemoryProvider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.MemoryProviders {
		if existing.Name == name {
			m.Name = name
			s.data.MemoryProviders[i] = m
			return s.persist()
		}
	}
	return fmt.Errorf("memory provider %q not found", name)
}

func (s *Store) DeleteMemoryProvider(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.MemoryProviders {
		if existing.Name == name {
			s.data.MemoryProviders = append(s.data.MemoryProviders[:i], s.data.MemoryProviders[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("memory provider %q not found", name)
}

// --- MCP Servers (global) ---

func (s *Store) ListMCPServers() []MCPServer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]MCPServer, len(s.data.MCPServers))
	copy(result, s.data.MCPServers)
	return result
}

func (s *Store) GetMCPServer(name string) (MCPServer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, m := range s.data.MCPServers {
		if m.Name == name {
			return m, true
		}
	}
	return MCPServer{}, false
}

func (s *Store) CreateMCPServer(m MCPServer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.data.MCPServers {
		if existing.Name == m.Name {
			return fmt.Errorf("MCP server %q already exists", m.Name)
		}
	}
	s.data.MCPServers = append(s.data.MCPServers, m)
	return s.persist()
}

func (s *Store) UpdateMCPServer(name string, m MCPServer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.MCPServers {
		if existing.Name == name {
			m.Name = name
			s.data.MCPServers[i] = m
			return s.persist()
		}
	}
	return fmt.Errorf("MCP server %q not found", name)
}

func (s *Store) DeleteMCPServer(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.MCPServers {
		if existing.Name == name {
			s.data.MCPServers = append(s.data.MCPServers[:i], s.data.MCPServers[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("MCP server %q not found", name)
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

func (s *Store) CreateAgent(a AgentDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.data.Agents {
		if existing.ID == a.ID {
			return fmt.Errorf("agent %q already exists", a.ID)
		}
	}
	if a.MCPServers == nil {
		a.MCPServers = []string{}
	}
	s.data.Agents = append(s.data.Agents, a)
	return s.persist()
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
			s.data.Agents[i] = a
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
			return s.persist()
		}
	}
	return fmt.Errorf("agent %q not found", id)
}

// --- Agent MCP linking ---

// LinkAgentMCP adds an MCP server reference to an agent.
func (s *Store) LinkAgentMCP(agentID, mcpName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mcpExists := false
	for _, m := range s.data.MCPServers {
		if m.Name == mcpName {
			mcpExists = true
			break
		}
	}
	if !mcpExists {
		return fmt.Errorf("MCP server %q not found", mcpName)
	}

	for i, a := range s.data.Agents {
		if a.ID == agentID {
			if slices.Contains(a.MCPServers, mcpName) {
				return fmt.Errorf("MCP %q already linked to agent %q", mcpName, agentID)
			}
			s.data.Agents[i].MCPServers = append(s.data.Agents[i].MCPServers, mcpName)
			return s.persist()
		}
	}
	return fmt.Errorf("agent %q not found", agentID)
}

// UnlinkAgentMCP removes an MCP server reference from an agent.
func (s *Store) UnlinkAgentMCP(agentID, mcpName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, a := range s.data.Agents {
		if a.ID == agentID {
			idx := slices.Index(a.MCPServers, mcpName)
			if idx == -1 {
				return fmt.Errorf("MCP %q not linked to agent %q", mcpName, agentID)
			}
			s.data.Agents[i].MCPServers = slices.Delete(a.MCPServers, idx, idx+1)
			return s.persist()
		}
	}
	return fmt.Errorf("agent %q not found", agentID)
}

// ResolveAgentMCPs returns the full MCPServer definitions for an agent's linked MCPs.
func (s *Store) ResolveAgentMCPs(agentID string) ([]MCPServer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var agentMCPNames []string
	found := false
	for _, a := range s.data.Agents {
		if a.ID == agentID {
			agentMCPNames = a.MCPServers
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("agent %q not found", agentID)
	}

	mcpMap := make(map[string]MCPServer, len(s.data.MCPServers))
	for _, m := range s.data.MCPServers {
		mcpMap[m.Name] = m
	}

	result := make([]MCPServer, 0, len(agentMCPNames))
	for _, name := range agentMCPNames {
		if m, ok := mcpMap[name]; ok {
			result = append(result, m)
		}
	}
	return result, nil
}

// --- Cron Jobs ---

func (s *Store) ListCronJobs() []CronJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]CronJob, len(s.data.CronJobs))
	copy(result, s.data.CronJobs)
	return result
}

func (s *Store) GetCronJob(name string) (CronJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.data.CronJobs {
		if c.Name == name {
			return c, true
		}
	}
	return CronJob{}, false
}

func (s *Store) CreateCronJob(c CronJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.data.CronJobs {
		if existing.Name == c.Name {
			return fmt.Errorf("cron job %q already exists", c.Name)
		}
	}
	s.data.CronJobs = append(s.data.CronJobs, c)
	return s.persist()
}

func (s *Store) UpdateCronJob(name string, c CronJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.CronJobs {
		if existing.Name == name {
			c.Name = name
			s.data.CronJobs[i] = c
			return s.persist()
		}
	}
	return fmt.Errorf("cron job %q not found", name)
}

func (s *Store) DeleteCronJob(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.CronJobs {
		if existing.Name == name {
			s.data.CronJobs = append(s.data.CronJobs[:i], s.data.CronJobs[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("cron job %q not found", name)
}

// --- Devices ---

func generateToken() string {
	b := make([]byte, 20)
	rand.Read(b)
	return "mgc_" + hex.EncodeToString(b)
}

func (s *Store) ListDevices() []Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Device, len(s.data.Devices))
	copy(result, s.data.Devices)
	return result
}

func (s *Store) GetDevice(name string) (Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.data.Devices {
		if d.Name == name {
			return d, true
		}
	}
	return Device{}, false
}

func (s *Store) GetDeviceByToken(token string) (Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, d := range s.data.Devices {
		if d.Token == token {
			return d, true
		}
	}
	return Device{}, false
}

func (s *Store) CreateDevice(d Device) (Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.data.Devices {
		if existing.Name == d.Name {
			return Device{}, fmt.Errorf("device %q already exists", d.Name)
		}
	}
	d.Token = generateToken()
	if d.AllowedAgents == nil {
		d.AllowedAgents = []string{}
	}
	s.data.Devices = append(s.data.Devices, d)
	return d, s.persist()
}

func (s *Store) UpdateDevice(name string, d Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Devices {
		if existing.Name == name {
			d.Name = name
			d.Token = existing.Token
			if d.AllowedAgents == nil {
				d.AllowedAgents = []string{}
			}
			s.data.Devices[i] = d
			return s.persist()
		}
	}
	return fmt.Errorf("device %q not found", name)
}

func (s *Store) DeleteDevice(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Devices {
		if existing.Name == name {
			s.data.Devices = append(s.data.Devices[:i], s.data.Devices[i+1:]...)
			return s.persist()
		}
	}
	return fmt.Errorf("device %q not found", name)
}

func (s *Store) RegenerateDeviceToken(name string) (Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.data.Devices {
		if existing.Name == name {
			s.data.Devices[i].Token = generateToken()
			return s.data.Devices[i], s.persist()
		}
	}
	return Device{}, fmt.Errorf("device %q not found", name)
}

// --- Rename with cascade ---

// RenameBackend renames a backend and updates all references in agents and memory providers.
func (s *Store) RenameBackend(oldName, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for i, b := range s.data.Backends {
		if b.Name == oldName {
			s.data.Backends[i].Name = newName
			found = true
		} else if b.Name == newName {
			return fmt.Errorf("backend %q already exists", newName)
		}
	}
	if !found {
		return fmt.Errorf("backend %q not found", oldName)
	}

	for i := range s.data.Agents {
		if s.data.Agents[i].LLM.Backend == oldName {
			s.data.Agents[i].LLM.Backend = newName
		}
		if s.data.Agents[i].Transcription.Backend == oldName {
			s.data.Agents[i].Transcription.Backend = newName
		}
		if s.data.Agents[i].TTS.Backend == oldName {
			s.data.Agents[i].TTS.Backend = newName
		}
	}

	for i := range s.data.MemoryProviders {
		if s.data.MemoryProviders[i].Embedding != nil && s.data.MemoryProviders[i].Embedding.Backend == oldName {
			s.data.MemoryProviders[i].Embedding.Backend = newName
		}
	}

	return s.persist()
}

// RenameMemoryProvider renames a memory provider and updates all agent references.
func (s *Store) RenameMemoryProvider(oldName, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for i, m := range s.data.MemoryProviders {
		if m.Name == oldName {
			s.data.MemoryProviders[i].Name = newName
			found = true
		} else if m.Name == newName {
			return fmt.Errorf("memory provider %q already exists", newName)
		}
	}
	if !found {
		return fmt.Errorf("memory provider %q not found", oldName)
	}

	for i := range s.data.Agents {
		if s.data.Agents[i].Memory.Session == oldName {
			s.data.Agents[i].Memory.Session = newName
		}
		if s.data.Agents[i].Memory.LongTerm == oldName {
			s.data.Agents[i].Memory.LongTerm = newName
		}
	}

	return s.persist()
}

// RenameMCPServer renames an MCP server and updates all agent references.
func (s *Store) RenameMCPServer(oldName, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for i, m := range s.data.MCPServers {
		if m.Name == oldName {
			s.data.MCPServers[i].Name = newName
			found = true
		} else if m.Name == newName {
			return fmt.Errorf("MCP server %q already exists", newName)
		}
	}
	if !found {
		return fmt.Errorf("MCP server %q not found", oldName)
	}

	for i := range s.data.Agents {
		for j, name := range s.data.Agents[i].MCPServers {
			if name == oldName {
				s.data.Agents[i].MCPServers[j] = newName
			}
		}
	}

	return s.persist()
}

// RenameAgent renames an agent and updates all references in devices and cron jobs.
func (s *Store) RenameAgent(oldID, newID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for i, a := range s.data.Agents {
		if a.ID == oldID {
			s.data.Agents[i].ID = newID
			found = true
		} else if a.ID == newID {
			return fmt.Errorf("agent %q already exists", newID)
		}
	}
	if !found {
		return fmt.Errorf("agent %q not found", oldID)
	}

	for i := range s.data.Devices {
		if s.data.Devices[i].DefaultAgent == oldID {
			s.data.Devices[i].DefaultAgent = newID
		}
		for j, id := range s.data.Devices[i].AllowedAgents {
			if id == oldID {
				s.data.Devices[i].AllowedAgents[j] = newID
			}
		}
	}

	for i := range s.data.CronJobs {
		if s.data.CronJobs[i].AgentID == oldID {
			s.data.CronJobs[i].AgentID = newID
		}
	}

	return s.persist()
}

// RenameDevice renames a device (no cascading references).
func (s *Store) RenameDevice(oldName, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for i, d := range s.data.Devices {
		if d.Name == oldName {
			s.data.Devices[i].Name = newName
			found = true
		} else if d.Name == newName {
			return fmt.Errorf("device %q already exists", newName)
		}
	}
	if !found {
		return fmt.Errorf("device %q not found", oldName)
	}

	return s.persist()
}

// RenameCronJob renames a cron job (no cascading references).
func (s *Store) RenameCronJob(oldName, newName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for i, c := range s.data.CronJobs {
		if c.Name == oldName {
			s.data.CronJobs[i].Name = newName
			found = true
		} else if c.Name == newName {
			return fmt.Errorf("cron job %q already exists", newName)
		}
	}
	if !found {
		return fmt.Errorf("cron job %q not found", oldName)
	}

	return s.persist()
}

// --- Persistence ---

func (s *Store) persist() error {
	if s.filePath == "" {
		return nil
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal store data: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write store file: %w", err)
	}

	s.notifyChange()
	return nil
}

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

func (s *Store) loadFromDisk() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var storeData StoreData
	if err := json.Unmarshal(data, &storeData); err != nil {
		return err
	}

	if storeData.Backends == nil {
		storeData.Backends = []BackendDefinition{}
	}
	if storeData.MemoryProviders == nil {
		storeData.MemoryProviders = []MemoryProvider{}
	}
	if storeData.MCPServers == nil {
		storeData.MCPServers = []MCPServer{}
	}
	if storeData.Agents == nil {
		storeData.Agents = []AgentDefinition{}
	}
	if storeData.CronJobs == nil {
		storeData.CronJobs = []CronJob{}
	}
	if storeData.Devices == nil {
		storeData.Devices = []Device{}
	}

	s.data = storeData
	return nil
}
