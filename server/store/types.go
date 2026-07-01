package store

import (
	"github.com/google/uuid"
)

// generateID returns a new random UUID v4 string (e.g. "550e8400-e29b-41d4-a716-446655440000").
func generateID() string {
	return uuid.New().String()
}

// AgentDefinition represents a single agent's full configuration in the store.
type AgentDefinition struct {
	ID            string              `json:"id" yaml:"id"`
	Name          string              `json:"name" yaml:"name"`
	Description   string              `json:"description,omitempty" yaml:"description,omitempty"`
	SystemPrompt  string              `json:"systemPrompt,omitempty" yaml:"systemPrompt,omitempty"`
	OutputKey     string              `json:"outputKey,omitempty" yaml:"outputKey,omitempty"`
	LLM           BackendRef          `json:"llm" yaml:"llm"`
	Transcription BackendRef          `json:"transcription,omitempty" yaml:"transcription,omitempty"`
	TTS           TTSRef              `json:"tts,omitempty" yaml:"tts,omitempty"`
	MCPServers    []string            `json:"mcpServers,omitempty" yaml:"mcpServers,omitempty"`
	Skills        []string            `json:"skills,omitempty" yaml:"skills,omitempty"`
	Tags          []string            `json:"tags,omitempty" yaml:"tags,omitempty"`
	ContextGuard  *ContextGuardConfig `json:"contextGuard,omitempty" yaml:"contextGuard,omitempty"`
	A2A           *A2AConfig          `json:"a2a,omitempty" yaml:"a2a,omitempty"`
}

// A2AConfig holds per-agent A2A (Agent-to-Agent) protocol settings.
// When Enabled is true the agent is discoverable and invocable via A2A.
type A2AConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

// BackendDefinition represents a reusable AI backend.
type BackendDefinition struct {
	ID      string            `json:"id" yaml:"id"`
	Name    string            `json:"name" yaml:"name"`
	Type    string            `json:"type" yaml:"type"`
	URL     string            `json:"url,omitempty" yaml:"url,omitempty"`
	APIKey  string            `json:"apiKey,omitempty" yaml:"apiKey,omitempty"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

// BackendRef holds a reference to a backend by ID + model.
type BackendRef struct {
	Backend string            `json:"backend,omitempty" yaml:"backend,omitempty"`
	Model   string            `json:"model,omitempty" yaml:"model,omitempty"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Config  STTConfig         `json:"config,omitempty" yaml:"config,omitempty"`
}

// TTSConfig holds provider-specific TTS configuration. Only the field matching
// the backend type should be populated.
type TTSConfig struct {
	OpenAI *OpenAITTSConfig `json:"openai,omitempty" yaml:"openai,omitempty"`
	Gemini *GeminiTTSConfig `json:"gemini,omitempty" yaml:"gemini,omitempty"`
}

// OpenAITTSConfig holds OpenAI-specific TTS settings.
type OpenAITTSConfig struct {
	Speed float64 `json:"speed,omitempty" yaml:"speed,omitempty"`
}

// GeminiTTSConfig holds Gemini-specific TTS settings.
type GeminiTTSConfig struct {
	LanguageCode string  `json:"languageCode,omitempty" yaml:"languageCode,omitempty"`
	Temperature  float64 `json:"temperature,omitempty" yaml:"temperature,omitempty"`
	StylePrompt  string  `json:"stylePrompt,omitempty" yaml:"stylePrompt,omitempty"`
}

// STTConfig holds provider-specific STT configuration. Only the field matching
// the backend type should be populated.
type STTConfig struct {
	OpenAI *OpenAISTTConfig `json:"openai,omitempty" yaml:"openai,omitempty"`
	Gemini *GeminiSTTConfig `json:"gemini,omitempty" yaml:"gemini,omitempty"`
}

// OpenAISTTConfig holds OpenAI-specific STT settings (currently empty, placeholder).
type OpenAISTTConfig struct{}

// GeminiSTTConfig holds Gemini-specific STT settings (currently empty, placeholder).
type GeminiSTTConfig struct{}

// TTSRef holds TTS-specific configuration referencing a backend by ID.
type TTSRef struct {
	Backend string    `json:"backend,omitempty" yaml:"backend,omitempty"`
	Model   string    `json:"model,omitempty" yaml:"model,omitempty"`
	Voice   string    `json:"voice,omitempty" yaml:"voice,omitempty"`
	Config  TTSConfig `json:"config,omitempty" yaml:"config,omitempty"`
}

// ContextGuardConfig holds per-agent context guard settings.
// When Enabled is true the plugin compacts conversation history using
// the selected Strategy. When Enabled is false (or the struct is nil)
// the plugin does nothing for this agent.
type ContextGuardConfig struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Strategy  string `json:"strategy,omitempty" yaml:"strategy,omitempty"`
	MaxTurns  int    `json:"maxTurns,omitempty" yaml:"maxTurns,omitempty"`
	MaxTokens int    `json:"maxTokens,omitempty" yaml:"maxTokens,omitempty"`
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
	Insecure     bool              `json:"insecure,omitempty" yaml:"insecure,omitempty"`
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
	Cron     *CronClientConfig     `json:"cron,omitempty" yaml:"cron,omitempty"`
	Webhook  *WebhookClientConfig  `json:"webhook,omitempty" yaml:"webhook,omitempty"`
}

// TelegramClientConfig holds Telegram bot settings for a client.
type TelegramClientConfig struct {
	BotToken     string  `json:"botToken,omitempty" yaml:"botToken,omitempty"`
	AllowedUsers []int64 `json:"allowedUsers,omitempty" yaml:"allowedUsers,omitempty"`
	AllowedChats []int64 `json:"allowedChats,omitempty" yaml:"allowedChats,omitempty"`
	ResponseMode string  `json:"responseMode,omitempty" yaml:"responseMode,omitempty"`
	DefaultAgent string  `json:"defaultAgent,omitempty" yaml:"defaultAgent,omitempty"`
}

// DiscordClientConfig holds Discord bot settings for a client.
// Uses the Discord Gateway (WebSocket) — no public URL needed.
type DiscordClientConfig struct {
	BotToken           string   `json:"botToken,omitempty" yaml:"botToken,omitempty"`
	AllowedUsers       []string `json:"allowedUsers,omitempty" yaml:"allowedUsers,omitempty"`
	AllowedChannels    []string `json:"allowedChannels,omitempty" yaml:"allowedChannels,omitempty"`
	ResponseMode       string   `json:"responseMode,omitempty" yaml:"responseMode,omitempty"`
	DefaultAgent       string   `json:"defaultAgent,omitempty" yaml:"defaultAgent,omitempty"`
	ThreadHistoryLimit int      `json:"threadHistoryLimit,omitempty" yaml:"threadHistoryLimit,omitempty"`
}

// SlackClientConfig holds Slack bot settings for a client.
// Uses Socket Mode (WebSocket) — no public URL needed.
type SlackClientConfig struct {
	BotToken           string   `json:"botToken,omitempty" yaml:"botToken,omitempty"`
	AppToken           string   `json:"appToken,omitempty" yaml:"appToken,omitempty"`
	AllowedUsers       []string `json:"allowedUsers,omitempty" yaml:"allowedUsers,omitempty"`
	AllowedChannels    []string `json:"allowedChannels,omitempty" yaml:"allowedChannels,omitempty"`
	ResponseMode       string   `json:"responseMode,omitempty" yaml:"responseMode,omitempty"`
	DefaultAgent       string   `json:"defaultAgent,omitempty" yaml:"defaultAgent,omitempty"`
	ThreadHistoryLimit int      `json:"threadHistoryLimit,omitempty" yaml:"threadHistoryLimit,omitempty"`
}

// CronClientConfig holds settings for a cron-type client.
type CronClientConfig struct {
	Schedule  string `json:"schedule" yaml:"schedule"`
	CommandID string `json:"commandId" yaml:"commandId"`
}

// WebhookClientConfig holds settings for a webhook-type client.
// Exactly one of Passthrough or CommandID must be set.
// When Passthrough is true, the prompt comes from the request body.
// When Passthrough is false, CommandID is required.
type WebhookClientConfig struct {
	Passthrough bool   `json:"passthrough" yaml:"passthrough"`
	CommandID   string `json:"commandId,omitempty" yaml:"commandId,omitempty"`
}

// Skill is the persistent identity of an Agent Skill. The store keeps the
// minimum needed to resolve cross-references (agents → skills via ID) and to
// locate the on-disk package (slug → directory). Everything else
// (name/description/instructions/resources) lives on disk as a real ADK
// SKILL.md package under data/skills/{slug}/, the layout consumed by ADK's
// skilltoolset (decision #29).
//
// Slug is the on-disk directory name and must match the SKILL.md
// frontmatter `name` field — that's an ADK invariant. The slug is fixed at
// upload time; renames go through the upload endpoint, which rewrites the
// frontmatter and renames the directory atomically.
type Skill struct {
	ID   string `json:"id" yaml:"id"`
	Slug string `json:"slug" yaml:"slug"`
}

// Command represents a reusable prompt that can be invoked against an agent
// via cron or webhook clients.
type Command struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Prompt      string `json:"prompt" yaml:"prompt"`
}

// FlowNodeType identifies the kind of vertex inside a flow graph.
const (
	// FlowNodeAgent wraps an AgentDefinition (or a sub-flow) and runs it.
	FlowNodeAgent = "agent"
	// FlowNodeRouter evaluates ordered CEL rules against the shared flow state
	// and emits a single route label, which its outgoing edges match.
	FlowNodeRouter = "router"
	// FlowNodeJoin is a fan-in barrier: it fires once after every declared
	// predecessor has completed. Routing into a join node is forbidden.
	FlowNodeJoin = "join"
	// FlowNodeParallel runs a wrapped agent once per item of a list-typed
	// input, concurrently, and aggregates the per-item outputs into a list.
	FlowNodeParallel = "parallel"
	// FlowNodeSubflow embeds another FlowDefinition as a single node (a nested
	// workflow). Its terminal output becomes this node's output.
	FlowNodeSubflow = "subflow"
	// FlowNodeExpression evaluates a CEL expression over `input` and `state`
	// and emits the result as its output. A deterministic transform node.
	FlowNodeExpression = "expression"
	// FlowNodeTemplate renders a text template with {{ input }} / {{ state.key }}
	// placeholders and emits the resulting string as its output.
	FlowNodeTemplate = "template"
)

// FlowStart is the reserved identifier for the graph entry sentinel. An edge
// whose From equals FlowStart is wired to the adk workflow Start node by the
// builder. Reserved so an operator cannot name a real node "START".
const FlowStart = "START"

// FlowRule is one ordered branch of a router node. When the CEL guard When
// evaluates to true against the flow state, the router emits Route as its
// label and stops evaluating later rules. When sees a `state` map (keys
// written through set_state, the "flow:" namespace) and an `iterations`
// integer (how many times this router has been activated in the current run).
type FlowRule struct {
	When  string `json:"when" yaml:"when"`
	Route string `json:"route" yaml:"route"`
}

// FlowNode is one vertex of the flow graph. ID is unique within the flow and
// becomes the adk workflow Node.Name(), so it must be a safe identifier: it
// also appears as the event Author used by the response filter and as a
// fragment of session-state keys.
type FlowNode struct {
	ID   string `json:"id" yaml:"id"`
	Type string `json:"type" yaml:"type"`

	// AgentID references an AgentDefinition by ID. Required when Type is
	// FlowNodeAgent (the agent to run) or FlowNodeParallel (the agent run once
	// per list item). Ignored for other types.
	AgentID string `json:"agentId,omitempty" yaml:"agentId,omitempty"`

	// ResponseAgent marks an agent node whose output is included in the final
	// response when the flow is invoked via webhook/cron. If no agent in the
	// flow is marked, all agent outputs are concatenated (default behaviour).
	// Only meaningful when Type is FlowNodeAgent.
	ResponseAgent bool `json:"responseAgent,omitempty" yaml:"responseAgent,omitempty"`

	// Rules and DefaultRoute drive a router node. Rules are evaluated in order;
	// DefaultRoute is emitted when no rule matches. Only meaningful when Type
	// is FlowNodeRouter.
	Rules        []FlowRule `json:"rules,omitempty" yaml:"rules,omitempty"`
	DefaultRoute string     `json:"defaultRoute,omitempty" yaml:"defaultRoute,omitempty"`

	// MaxConcurrency caps how many items a parallel node processes at once.
	// 0 means unlimited. Only meaningful when Type is FlowNodeParallel.
	MaxConcurrency int `json:"maxConcurrency,omitempty" yaml:"maxConcurrency,omitempty"`

	// FlowID references another FlowDefinition embedded as a nested workflow.
	// Required when Type is FlowNodeSubflow, ignored otherwise.
	FlowID string `json:"flowId,omitempty" yaml:"flowId,omitempty"`

	// Expression is a CEL expression evaluated over `input` (the previous
	// node's output) and `state` (the shared flow state). Its result becomes
	// this node's output. Required when Type is FlowNodeExpression.
	Expression string `json:"expression,omitempty" yaml:"expression,omitempty"`

	// Template is text with {{ input }}, {{ input.field }} and {{ state.key }}
	// placeholders; the rendered string becomes this node's output. Required
	// when Type is FlowNodeTemplate.
	Template string `json:"template,omitempty" yaml:"template,omitempty"`

	// OutputKey, when set, also writes this node's output into the shared flow
	// state under that key, readable downstream as state.<key>. It must be a
	// valid CEL identifier (letters, digits, underscore; no hyphen) so
	// downstream expressions can reference it. Meaningful for expression and
	// template nodes.
	OutputKey string `json:"outputKey,omitempty" yaml:"outputKey,omitempty"`

	// X and Y are the node's position on the visual editor canvas. They are a
	// layout hint for the admin UI only: the builder and validation ignore
	// them, so a flow authored without an editor (or with them omitted) still
	// runs. Persisted so the operator's arrangement survives a round-trip.
	X float64 `json:"x,omitempty" yaml:"x,omitempty"`
	Y float64 `json:"y,omitempty" yaml:"y,omitempty"`
}

// FlowEdge is a directed connection between two nodes. Route is only
// meaningful when From is a router node: it names the label the router must
// emit for this edge to be taken. An empty Route is an unconditional edge.
type FlowEdge struct {
	From  string `json:"from" yaml:"from"`
	To    string `json:"to" yaml:"to"`
	Route string `json:"route,omitempty" yaml:"route,omitempty"`
}

// ResponseAgentIDs returns the AgentDefinition IDs of every agent node marked
// with ResponseAgent.
func (f *FlowDefinition) ResponseAgentIDs() []string {
	var ids []string
	for i := range f.Nodes {
		n := &f.Nodes[i]
		if n.Type == FlowNodeAgent && n.AgentID != "" && n.ResponseAgent {
			ids = append(ids, n.AgentID)
		}
	}
	return ids
}

// ResponseAgentNames returns the adk node names whose output should appear in
// the final response. In the graph model a node's ID is its adk Node.Name()
// and therefore the event.Author the response filter matches against, so the
// names are simply the IDs of agent nodes flagged ResponseAgent. There is no
// synthetic naming convention to keep in lockstep with the builder anymore.
func (f *FlowDefinition) ResponseAgentNames() []string {
	var names []string
	for i := range f.Nodes {
		n := &f.Nodes[i]
		if n.Type == FlowNodeAgent && n.ResponseAgent {
			names = append(names, n.ID)
		}
	}
	return names
}

// FirstAgentID returns the AgentDefinition ID of the first agent node reached
// from the entry, breadth-first. Used to resolve voice config (TTS/STT) for a
// flow. Falls back to the first agent node in declaration order when the entry
// reaches none, and to "" when the flow has no agent node at all.
func (f *FlowDefinition) FirstAgentID() string {
	index := make(map[string]*FlowNode, len(f.Nodes))
	for i := range f.Nodes {
		index[f.Nodes[i].ID] = &f.Nodes[i]
	}
	successors := make(map[string][]string, len(f.Nodes))
	for _, e := range f.Edges {
		successors[e.From] = append(successors[e.From], e.To)
	}
	visited := map[string]bool{}
	queue := []string{f.Entry}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if visited[id] {
			continue
		}
		visited[id] = true
		if n, ok := index[id]; ok && (n.Type == FlowNodeAgent || n.Type == FlowNodeParallel) && n.AgentID != "" {
			return n.AgentID
		}
		queue = append(queue, successors[id]...)
	}
	for i := range f.Nodes {
		if (f.Nodes[i].Type == FlowNodeAgent || f.Nodes[i].Type == FlowNodeParallel) && f.Nodes[i].AgentID != "" {
			return f.Nodes[i].AgentID
		}
	}
	return ""
}

// AgentIDs returns all unique entity IDs this flow references: the AgentID of
// agent and parallel nodes, plus the FlowID of subflow nodes. Used by the
// topological sort to discover sub-flow dependencies (a referenced ID that
// resolves to another flow), so it must include every cross-entity reference.
func (f *FlowDefinition) AgentIDs() []string {
	seen := map[string]bool{}
	var ids []string
	add := func(id string) {
		if id != "" && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	for i := range f.Nodes {
		n := &f.Nodes[i]
		switch n.Type {
		case FlowNodeAgent, FlowNodeParallel:
			add(n.AgentID)
		case FlowNodeSubflow:
			add(n.FlowID)
		}
	}
	return ids
}

// FlowDefinition is a multi-agent workflow stored as a directed graph that
// maps one-to-one onto the adk-go v2 workflow engine ([]workflow.Edge wired
// into workflowagent.New). Entry names the node connected to the Start
// sentinel; it is recorded explicitly rather than inferred so a multi-root
// graph keeps the operator's intent.
type FlowDefinition struct {
	ID          string     `json:"id" yaml:"id"`
	Name        string     `json:"name" yaml:"name"`
	Description string     `json:"description,omitempty" yaml:"description,omitempty"`
	Entry       string     `json:"entry" yaml:"entry"`
	Nodes       []FlowNode `json:"nodes" yaml:"nodes"`
	Edges       []FlowEdge `json:"edges" yaml:"edges"`
	A2A         *A2AConfig `json:"a2a,omitempty" yaml:"a2a,omitempty"`
}

// Settings holds global configuration that applies to the launcher/runtime
// rather than to individual entities.
type Settings struct {
	SessionProvider  string `json:"sessionProvider,omitempty" yaml:"sessionProvider,omitempty"`
	LongTermProvider string `json:"longTermProvider,omitempty" yaml:"longTermProvider,omitempty"`
	// TemporaryDir is the absolute path used by tools and subsystems that need
	// to write transient files visible to other on-disk consumers (filesystem
	// MCPs, shell tools, etc.). When empty, callers must fall back to the OS
	// temporary directory via Store.ResolveTemporaryDir, which is the only
	// place that performs that fallback.
	TemporaryDir string `json:"temporaryDir,omitempty" yaml:"temporaryDir,omitempty"`
}

// Secret represents an encrypted key-value pair used for environment variable injection.
// The Key field is the environment variable name (e.g. OPENAI_API_KEY).
// The Value is stored encrypted at rest when an admin password is configured.
type Secret struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Key         string `json:"key" yaml:"key"`
	Value       string `json:"value" yaml:"value"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// StoreData is the top-level structure persisted to disk.
type StoreData struct {
	Settings        Settings            `json:"settings"`
	Backends        []BackendDefinition `json:"backends"`
	MemoryProviders []MemoryProvider    `json:"memoryProviders"`
	MCPServers      []MCPServer         `json:"mcpServers"`
	Skills          []Skill             `json:"skills"`
	Agents          []AgentDefinition   `json:"agents"`
	Clients         []ClientDefinition  `json:"clients"`
	Flows           []FlowDefinition    `json:"flows"`
	Commands        []Command           `json:"commands"`
	Secrets         []Secret            `json:"secrets"`
}
