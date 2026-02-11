package store

// AgentDefinition represents a single agent's full configuration in the store.
type AgentDefinition struct {
	ID                 string         `json:"id" yaml:"id"`
	Name               string         `json:"name" yaml:"name"`
	Description        string         `json:"description,omitempty" yaml:"description,omitempty"`
	SystemPrompt       string         `json:"systemPrompt,omitempty" yaml:"systemPrompt,omitempty"`
	SystemPromptSuffix string         `json:"systemPromptSuffix,omitempty" yaml:"systemPromptSuffix,omitempty"`
	LLM                BackendRef     `json:"llm" yaml:"llm"`
	Transcription      BackendRef     `json:"transcription,omitempty" yaml:"transcription,omitempty"`
	TTS                TTSRef         `json:"tts,omitempty" yaml:"tts,omitempty"`
	Memory             MemoryRef      `json:"memory,omitempty" yaml:"memory,omitempty"`
	MCPServers         []string       `json:"mcpServers,omitempty" yaml:"mcpServers,omitempty"`
	Telegram           TelegramConfig `json:"telegram,omitempty" yaml:"telegram,omitempty"`
}

// BackendDefinition represents a reusable AI backend.
type BackendDefinition struct {
	Name   string `json:"name" yaml:"name"`
	Type   string `json:"type" yaml:"type"`
	URL    string `json:"url,omitempty" yaml:"url,omitempty"`
	APIKey string `json:"apiKey,omitempty" yaml:"apiKey,omitempty"`
}

// BackendRef holds a reference to a backend by name + model.
type BackendRef struct {
	Backend string `json:"backend,omitempty" yaml:"backend,omitempty"`
	Model   string `json:"model,omitempty" yaml:"model,omitempty"`
}

// TTSRef holds TTS-specific configuration referencing a backend.
type TTSRef struct {
	Backend string  `json:"backend,omitempty" yaml:"backend,omitempty"`
	Model   string  `json:"model,omitempty" yaml:"model,omitempty"`
	Voice   string  `json:"voice,omitempty" yaml:"voice,omitempty"`
	Speed   float64 `json:"speed,omitempty" yaml:"speed,omitempty"`
}

// MemoryRef holds references to memory providers by name.
type MemoryRef struct {
	Session  string `json:"session,omitempty" yaml:"session,omitempty"`
	LongTerm string `json:"longTerm,omitempty" yaml:"longTerm,omitempty"`
}

// MemoryProvider represents a reusable memory backend (Redis, Postgres, etc.).
// Type identifies the backend technology. Category defines the role this
// instance serves: "session" for short-lived state, "longterm" for persistent
// memory with embeddings. The same Type may support both categories.
//
// Config holds provider-specific connection details (address, password, etc.)
// as an opaque map — each provider type defines its own fields.
// Embedding is kept as a top-level field because it's a structural concept
// shared by all long-term providers, not a connection detail.
type MemoryProvider struct {
	Name      string                 `json:"name" yaml:"name"`
	Type      string                 `json:"type" yaml:"type"`
	Category  string                 `json:"category" yaml:"category"`
	Config    map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty"`
	Embedding *BackendRef            `json:"embedding,omitempty" yaml:"embedding,omitempty"`
}

// MCPServer represents an MCP server configuration.
type MCPServer struct {
	Name         string            `json:"name" yaml:"name"`
	Type         string            `json:"type,omitempty" yaml:"type,omitempty"`
	Endpoint     string            `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Headers      map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Command      string            `json:"command,omitempty" yaml:"command,omitempty"`
	Args         []string          `json:"args,omitempty" yaml:"args,omitempty"`
	Env          map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	WorkDir      string            `json:"workDir,omitempty" yaml:"workDir,omitempty"`
	SystemPrompt string            `json:"systemPrompt,omitempty" yaml:"systemPrompt,omitempty"`
}

// TelegramConfig holds Telegram bot settings for an agent.
type TelegramConfig struct {
	Enabled      bool    `json:"enabled" yaml:"enabled"`
	Token        string  `json:"token,omitempty" yaml:"token,omitempty"`
	AllowedUsers []int64 `json:"allowedUsers,omitempty" yaml:"allowedUsers,omitempty"`
	AllowedChats []int64 `json:"allowedChats,omitempty" yaml:"allowedChats,omitempty"`
	ResponseMode string  `json:"responseMode,omitempty" yaml:"responseMode,omitempty"`
}

// CronJob represents a scheduled task that sends a prompt to an agent.
type CronJob struct {
	Name        string `json:"name" yaml:"name"`
	Schedule    string `json:"schedule" yaml:"schedule"`
	AgentID     string `json:"agentId" yaml:"agentId"`
	Prompt      string `json:"prompt" yaml:"prompt"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled     bool   `json:"enabled" yaml:"enabled"`
}

// Device represents an access point (tablet, phone, kiosk) that connects to the voice-UI.
type Device struct {
	Name          string   `json:"name" yaml:"name"`
	Token         string   `json:"token" yaml:"token"`
	DefaultAgent  string   `json:"defaultAgent" yaml:"defaultAgent"`
	AllowedAgents []string `json:"allowedAgents" yaml:"allowedAgents"`
	Enabled       bool     `json:"enabled" yaml:"enabled"`
}

// StoreData is the top-level structure persisted to disk.
type StoreData struct {
	Backends        []BackendDefinition `json:"backends"`
	MemoryProviders []MemoryProvider    `json:"memoryProviders"`
	MCPServers      []MCPServer         `json:"mcpServers"`
	Agents          []AgentDefinition   `json:"agents"`
	CronJobs        []CronJob           `json:"cronJobs"`
	Devices         []Device            `json:"devices"`
}
