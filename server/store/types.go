package store

import (
	"crypto/rand"
	"fmt"
)

// generateID returns a random 16-byte hex string (128 bits) suitable for use
// as an immutable resource identifier.
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// AgentDefinition represents a single agent's full configuration in the store.
type AgentDefinition struct {
	ID           string     `json:"id" yaml:"id"`
	Name         string     `json:"name" yaml:"name"`
	Description  string     `json:"description,omitempty" yaml:"description,omitempty"`
	SystemPrompt string     `json:"systemPrompt,omitempty" yaml:"systemPrompt,omitempty"`
	LLM          BackendRef `json:"llm" yaml:"llm"`
	Transcription BackendRef `json:"transcription,omitempty" yaml:"transcription,omitempty"`
	TTS          TTSRef     `json:"tts,omitempty" yaml:"tts,omitempty"`
	Memory       MemoryRef  `json:"memory,omitempty" yaml:"memory,omitempty"`
	MCPServers   []string   `json:"mcpServers,omitempty" yaml:"mcpServers,omitempty"`
}

// BackendDefinition represents a reusable AI backend.
type BackendDefinition struct {
	ID     string `json:"id" yaml:"id"`
	Name   string `json:"name" yaml:"name"`
	Type   string `json:"type" yaml:"type"`
	URL    string `json:"url,omitempty" yaml:"url,omitempty"`
	APIKey string `json:"apiKey,omitempty" yaml:"apiKey,omitempty"`
}

// BackendRef holds a reference to a backend by ID + model.
type BackendRef struct {
	Backend string `json:"backend,omitempty" yaml:"backend,omitempty"`
	Model   string `json:"model,omitempty" yaml:"model,omitempty"`
}

// TTSRef holds TTS-specific configuration referencing a backend by ID.
type TTSRef struct {
	Backend string  `json:"backend,omitempty" yaml:"backend,omitempty"`
	Model   string  `json:"model,omitempty" yaml:"model,omitempty"`
	Voice   string  `json:"voice,omitempty" yaml:"voice,omitempty"`
	Speed   float64 `json:"speed,omitempty" yaml:"speed,omitempty"`
}

// MemoryRef holds references to memory providers by ID.
type MemoryRef struct {
	Session  string `json:"session,omitempty" yaml:"session,omitempty"`
	LongTerm string `json:"longTerm,omitempty" yaml:"longTerm,omitempty"`
}

// MemoryProvider represents a reusable memory backend (Redis, Postgres, etc.).
type MemoryProvider struct {
	ID        string                 `json:"id" yaml:"id"`
	Name      string                 `json:"name" yaml:"name"`
	Type      string                 `json:"type" yaml:"type"`
	Category  string                 `json:"category" yaml:"category"`
	Config    map[string]interface{} `json:"config,omitempty" yaml:"config,omitempty"`
	Embedding *BackendRef            `json:"embedding,omitempty" yaml:"embedding,omitempty"`
}

// MCPServer represents an MCP server configuration.
type MCPServer struct {
	ID           string            `json:"id" yaml:"id"`
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

// ClientDefinition represents an access point (voice-ui, Telegram, Discord, webhook, etc.).
// Type determines what platform-specific config is expected inside Config.
type ClientDefinition struct {
	ID            string       `json:"id" yaml:"id"`
	Name          string       `json:"name" yaml:"name"`
	Type          string       `json:"type" yaml:"type"`
	Token         string       `json:"token" yaml:"token"`
	AllowedAgents []string     `json:"allowedAgents" yaml:"allowedAgents"`
	Enabled       bool         `json:"enabled" yaml:"enabled"`
	Config        ClientConfig `json:"config" yaml:"config"`
}

// ClientConfig holds platform-specific configuration. Only the field matching
// the ClientDefinition.Type should be populated.
type ClientConfig struct {
	Telegram *TelegramClientConfig `json:"telegram,omitempty" yaml:"telegram,omitempty"`
	Discord  *DiscordClientConfig  `json:"discord,omitempty" yaml:"discord,omitempty"`
	Slack    *SlackClientConfig    `json:"slack,omitempty" yaml:"slack,omitempty"`
}

// TelegramClientConfig holds Telegram bot settings for a client.
type TelegramClientConfig struct {
	BotToken     string  `json:"botToken,omitempty" yaml:"botToken,omitempty"`
	AllowedUsers []int64 `json:"allowedUsers,omitempty" yaml:"allowedUsers,omitempty"`
	AllowedChats []int64 `json:"allowedChats,omitempty" yaml:"allowedChats,omitempty"`
	ResponseMode string  `json:"responseMode,omitempty" yaml:"responseMode,omitempty"`
}

// DiscordClientConfig holds Discord bot settings for a client.
type DiscordClientConfig struct {
	BotToken        string   `json:"botToken,omitempty" yaml:"botToken,omitempty"`
	GuildID         string   `json:"guildId,omitempty" yaml:"guildId,omitempty"`
	AllowedUsers    []string `json:"allowedUsers,omitempty" yaml:"allowedUsers,omitempty"`
	AllowedChannels []string `json:"allowedChannels,omitempty" yaml:"allowedChannels,omitempty"`
}

// SlackClientConfig holds Slack bot settings for a client.
type SlackClientConfig struct {
	BotToken        string   `json:"botToken,omitempty" yaml:"botToken,omitempty"`
	SigningSecret   string   `json:"signingSecret,omitempty" yaml:"signingSecret,omitempty"`
	AllowedUsers    []string `json:"allowedUsers,omitempty" yaml:"allowedUsers,omitempty"`
	AllowedChannels []string `json:"allowedChannels,omitempty" yaml:"allowedChannels,omitempty"`
}

// CronJob represents a scheduled task that sends a prompt to an agent.
type CronJob struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Schedule    string `json:"schedule" yaml:"schedule"`
	AgentID     string `json:"agentId" yaml:"agentId"`
	Prompt      string `json:"prompt" yaml:"prompt"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled     bool   `json:"enabled" yaml:"enabled"`
}

// StoreData is the top-level structure persisted to disk.
type StoreData struct {
	Backends        []BackendDefinition `json:"backends"`
	MemoryProviders []MemoryProvider    `json:"memoryProviders"`
	MCPServers      []MCPServer         `json:"mcpServers"`
	Agents          []AgentDefinition   `json:"agents"`
	CronJobs        []CronJob           `json:"cronJobs"`
	Clients         []ClientDefinition  `json:"clients"`
}
