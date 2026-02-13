package store

import "github.com/google/uuid"

// generateID returns a new random UUID v4 string (e.g. "550e8400-e29b-41d4-a716-446655440000").
func generateID() string {
	return uuid.New().String()
}

// AgentDefinition represents a single agent's full configuration in the store.
type AgentDefinition struct {
	ID           string     `json:"id" yaml:"id"`
	Name         string     `json:"name" yaml:"name"`
	Description  string     `json:"description,omitempty" yaml:"description,omitempty"`
	SystemPrompt string     `json:"systemPrompt,omitempty" yaml:"systemPrompt,omitempty"`
	OutputKey    string     `json:"outputKey,omitempty" yaml:"outputKey,omitempty"`
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

// Command represents a reusable prompt that can be invoked against an agent
// via triggers (cron jobs, webhooks) or other automation.
type Command struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Prompt      string `json:"prompt" yaml:"prompt"`
}

// TriggerType identifies the kind of trigger.
const (
	TriggerTypeCron    = "cron"
	TriggerTypeWebhook = "webhook"
)

// Trigger represents an automation that executes a command against an agent.
// Type determines which config block is used.
type Trigger struct {
	ID          string          `json:"id" yaml:"id"`
	Name        string          `json:"name" yaml:"name"`
	Description string          `json:"description,omitempty" yaml:"description,omitempty"`
	Type        string          `json:"type" yaml:"type"`
	Enabled     bool            `json:"enabled" yaml:"enabled"`
	AgentID     string          `json:"agentId,omitempty" yaml:"agentId,omitempty"`
	CommandID   string          `json:"commandId,omitempty" yaml:"commandId,omitempty"`
	ClientID    string          `json:"clientId,omitempty" yaml:"clientId,omitempty"`
	Cron        *CronConfig     `json:"cron,omitempty" yaml:"cron,omitempty"`
	Webhook     *WebhookConfig  `json:"webhook,omitempty" yaml:"webhook,omitempty"`
}

// CronConfig holds the schedule for a cron trigger.
type CronConfig struct {
	Schedule string `json:"schedule" yaml:"schedule"`
}

// WebhookConfig holds settings for a webhook trigger.
// When Passthrough is true, the prompt comes from the request body
// and CommandID/AgentID on the Trigger may be empty.
type WebhookConfig struct {
	Passthrough bool   `json:"passthrough" yaml:"passthrough"`
	Secret      string `json:"secret,omitempty" yaml:"secret,omitempty"`
}

// FlowStepType identifies the kind of node inside a flow.
const (
	FlowStepAgent      = "agent"
	FlowStepSequential = "sequential"
	FlowStepParallel   = "parallel"
	FlowStepLoop       = "loop"
)

// FlowStep is a recursive node in a flow tree.
// Leaf nodes have Type "agent" and reference an AgentDefinition by ID.
// Container nodes have Type "sequential", "parallel", or "loop" and hold
// child steps. Loop nodes additionally specify MaxIterations.
type FlowStep struct {
	Type          string     `json:"type"`
	AgentID       string     `json:"agentId,omitempty"`
	MaxIterations uint       `json:"maxIterations,omitempty"`
	Steps         []FlowStep `json:"steps,omitempty"`
}

// FlowDefinition represents a multi-agent workflow stored as a recursive tree
// of steps that maps directly to ADK workflow agents.
type FlowDefinition struct {
	ID          string   `json:"id" yaml:"id"`
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Root        FlowStep `json:"root" yaml:"root"`
}

// StoreData is the top-level structure persisted to disk.
type StoreData struct {
	Backends        []BackendDefinition `json:"backends"`
	MemoryProviders []MemoryProvider    `json:"memoryProviders"`
	MCPServers      []MCPServer         `json:"mcpServers"`
	Agents          []AgentDefinition   `json:"agents"`
	CronJobs        []CronJob           `json:"cronJobs"`
	Clients         []ClientDefinition  `json:"clients"`
	Flows           []FlowDefinition    `json:"flows"`
	Commands        []Command           `json:"commands"`
	Triggers        []Trigger           `json:"triggers"`
}

// CronJob is the legacy type kept for data migration. New code uses Trigger.
type CronJob struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Schedule    string `json:"schedule" yaml:"schedule"`
	AgentID     string `json:"agentId" yaml:"agentId"`
	Prompt      string `json:"prompt" yaml:"prompt"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled     bool   `json:"enabled" yaml:"enabled"`
}
