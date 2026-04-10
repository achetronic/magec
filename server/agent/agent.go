// Copyright 2025 Alby Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/server/adkrest"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/mcptoolset"
	"google.golang.org/genai"

	artifactfs "github.com/achetronic/adk-utils-go/artifact/filesystem"
	genaianthro "github.com/achetronic/adk-utils-go/genai/anthropic"
	genaiopenai "github.com/achetronic/adk-utils-go/genai/openai"
	memorypostgres "github.com/achetronic/adk-utils-go/memory/postgres"
	"github.com/achetronic/adk-utils-go/plugin/contextguard"
	sessionredis "github.com/achetronic/adk-utils-go/session/redis"
	toolsmemory "github.com/achetronic/adk-utils-go/tools/memory"

	"github.com/achetronic/magec/server/config"
	"github.com/achetronic/magec/server/store"
)

const baseInstruction = `You are Magec, a helpful AI assistant that helps users with various tasks.
Keep responses concise and natural for interaction.
Respond in the same language as the user's input.`

const memoryInstruction = `
You have access to long-term memory tools:
- Use 'search_memory' to recall information from past conversations. IMPORTANT: When this tool returns memories, you MUST use that information in your response. The 'memories' array contains the actual data - read the 'text' field of each entry.
- Use 'save_to_memory' to remember important facts, user preferences, or anything the user asks you to remember

CRITICAL: At the START of every conversation, you MUST call search_memory with a broad query to retrieve any stored user preferences, instructions, or important information. This ensures you always have context about the user before responding.

When a user asks you to remember something or asks about past information:
1. First use search_memory to check if you have relevant information
2. If search_memory returns results (count > 0), USE the text from those memories in your answer
3. Only say you don't have information if search_memory returns count: 0

When a user shares preferences or important information, proactively save it to memory for future reference.`

const artifactInstruction = `
You have access to artifact tools for creating and managing files:
- Use 'save_artifact' to save code, documents, data files, or any content that should be delivered as a downloadable file. Provide a filename (e.g. "report.md", "main.py", "data.csv"), the content, and optionally a mime_type. For binary content, set is_base64=true and provide base64-encoded data.
- Use 'load_artifact' to retrieve a previously saved artifact by name.
- Use 'list_artifacts' to see all artifacts in the current session.

IMPORTANT: When generating code files, long documents, configuration files, scripts, or any substantial structured content, ALWAYS use save_artifact instead of pasting it in the chat. The artifact will be delivered to the user as a downloadable file automatically.`

// Service wraps the ADK REST handler that serves all configured agents.
// Incoming requests are routed to the correct agent by the appName field.
type Service struct {
	handler    http.Handler
	sessionSvc session.Service
	memorySvc  memory.Service
	adkAgents  map[string]agent.Agent
}

// New builds an ADK agent for every AgentDefinition in the store, wires up
// their LLM, session, memory, and MCP toolsets, and returns a Service that
// routes requests to the right agent based on the appName in the request body.
// Any FlowDefinitions are translated into ADK workflow agents and registered
// alongside the regular agents.
func New(ctx context.Context, agents []store.AgentDefinition, backends []store.BackendDefinition, memoryProviders []store.MemoryProvider, mcpServers []store.MCPServer, skills []store.Skill, flows []store.FlowDefinition, settings store.Settings, registry contextguard.ModelRegistry) (*Service, error) {
	if len(agents) == 0 {
		return nil, fmt.Errorf("no agents defined")
	}

	backendMap := make(map[string]store.BackendDefinition, len(backends))
	for _, b := range backends {
		backendMap[b.ID] = b
	}

	memoryProviderMap := make(map[string]store.MemoryProvider, len(memoryProviders))
	for _, m := range memoryProviders {
		memoryProviderMap[m.ID] = m
	}

	mcpServerMap := make(map[string]store.MCPServer, len(mcpServers))
	for _, m := range mcpServers {
		mcpServerMap[m.ID] = m
	}

	skillMap := make(map[string]store.Skill, len(skills))
	for _, sk := range skills {
		skillMap[sk.ID] = sk
	}

	sessionSvc, err := createSessionService(settings, memoryProviderMap)
	if err != nil {
		return nil, fmt.Errorf("session service: %w", err)
	}

	memorySvc, err := createMemoryService(ctx, settings, memoryProviderMap, backendMap)
	if err != nil {
		return nil, fmt.Errorf("memory service: %w", err)
	}

	var rootAgent agent.Agent
	var otherAgents []agent.Agent
	adkAgentMap := make(map[string]agent.Agent, len(agents))
	// llmMap maps agent ID → LLM instance. The ContextGuard plugin uses it
	// so each agent summarizes with its own model, matching user expectations.
	// Rebuilt from scratch on every hot-reload (store change).
	llmMap := make(map[string]model.LLM, len(agents))

	artifactSvc, err := artifactfs.NewFilesystemService(artifactfs.FilesystemServiceConfig{
		BasePath: filepath.Join("data", "artifacts"),
	})
	if err != nil {
		return nil, fmt.Errorf("artifact service: %w", err)
	}

	baseTset, err := newBaseToolset()
	if err != nil {
		return nil, fmt.Errorf("failed to create base toolset: %w", err)
	}

	for i, agentDef := range agents {

		llmBackend, ok := backendMap[agentDef.LLM.Backend]
		if !ok {
			return nil, fmt.Errorf("agent %q: LLM backend %q not found", agentDef.ID, agentDef.LLM.Backend)
		}
		llmModel, err := createLLM(ctx, llmBackend, agentDef.LLM)
		if err != nil {
			return nil, fmt.Errorf("agent %q: failed to create LLM: %w", agentDef.ID, err)
		}
		// Register this agent's LLM so ContextGuard can use it for summarization.
		llmMap[agentDef.ID] = llmModel

		toolsets, err := buildToolsets(agentDef, mcpServerMap, memorySvc)
		if err != nil {
			return nil, fmt.Errorf("agent %q: failed to build toolsets: %w", agentDef.ID, err)
		}
		toolsets = append(toolsets, baseTset)

		instruction := buildInstruction(agentDef, mcpServerMap, skillMap, filepath.Join("data", "skills"), memorySvc)

		agentCfg := llmagent.Config{
			Name:                agentDef.ID,
			Model:               llmModel,
			Description:         agentDef.Name,
			InstructionProvider: makeInstructionProvider(instruction),
			Toolsets:            toolsets,
			OutputKey:           agentDef.OutputKey,
		}

		adkAgent, err := llmagent.New(agentCfg)
		if err != nil {
			return nil, fmt.Errorf("agent %q: failed to create: %w", agentDef.ID, err)
		}

		if i == 0 {
			rootAgent = adkAgent
		} else {
			otherAgents = append(otherAgents, adkAgent)
		}
		adkAgentMap[agentDef.ID] = adkAgent

		slog.Info("Agent initialized", "id", agentDef.ID, "name", agentDef.Name)
	}

	for _, flow := range flows {
		flowAgent, err := BuildFlowAgent(flow, adkAgentMap)
		if err != nil {
			slog.Warn("Failed to build flow", "flow", flow.Name, "error", err)
			continue
		}
		otherAgents = append(otherAgents, flowAgent)
		adkAgentMap[flow.ID] = flowAgent
		slog.Info("Flow initialized", "id", flow.ID, "name", flow.Name)
	}

	loader, err := agent.NewMultiLoader(rootAgent, otherAgents...)
	if err != nil {
		return nil, fmt.Errorf("failed to create multi-loader: %w", err)
	}

	launcherCfg := &launcher.Config{
		SessionService:  sessionSvc,
		AgentLoader:     loader,
		ArtifactService: artifactSvc,
	}
	if memorySvc != nil {
		launcherCfg.MemoryService = memorySvc
	}
	// Wire the ContextGuard plugin if a context window registry was provided.
	// The plugin receives the full llmMap so every agent summarizes with its
	// own model — a user on a powerful model gets a high-quality summary,
	// a user on a cheap model gets a summary matching those expectations.
	if registry != nil {
		guard := contextguard.New(registry)
		for _, agentDef := range agents {
			cg := agentDef.ContextGuard
			if cg == nil || !cg.Enabled {
				continue
			}
			var opts []contextguard.AgentOption
			switch cg.Strategy {
			case contextguard.StrategySlidingWindow:
				opts = append(opts, contextguard.WithSlidingWindow(cg.MaxTurns))
			default:
				if cg.MaxTokens > 0 {
					opts = append(opts, contextguard.WithMaxTokens(cg.MaxTokens))
				}
			}
			guard.Add(agentDef.ID, llmMap[agentDef.ID], opts...)
		}
		launcherCfg.PluginConfig = guard.PluginConfig()
	}

	return &Service{
		handler:    adkrest.NewHandler(launcherCfg, 15*time.Minute),
		sessionSvc: sessionSvc,
		memorySvc:  memorySvc,
		adkAgents:  adkAgentMap,
	}, nil
}

// Handler returns the HTTP handler that serves the ADK REST API.
func (s *Service) Handler() http.Handler {
	return s.handler
}

// SessionService returns the session.Service used by the launcher.
func (s *Service) SessionService() session.Service {
	return s.sessionSvc
}

// MemoryService returns the memory.Service used by the launcher (may be nil).
func (s *Service) MemoryService() memory.Service {
	return s.memorySvc
}

// ADKAgents returns the map of agent ID → ADK agent instance.
// Used by the A2A handler to create per-agent executors.
func (s *Service) ADKAgents() map[string]agent.Agent {
	return s.adkAgents
}

// createSessionService returns the session backend based on global settings.
// Falls back to in-memory if no provider is configured.
func createSessionService(settings store.Settings, memoryProviders map[string]store.MemoryProvider) (session.Service, error) {
	if settings.SessionProvider == "" {
		return session.InMemoryService(), nil
	}

	provider, ok := memoryProviders[settings.SessionProvider]
	if !ok {
		return session.InMemoryService(), nil
	}

	connStr, _ := provider.Config["connectionString"].(string)
	if connStr == "" {
		return session.InMemoryService(), nil
	}

	ttlStr, _ := provider.Config["ttl"].(string)
	if ttlStr == "" {
		ttlStr = "24h"
	}
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		ttl = 24 * time.Hour
	}

	addr, password, db := parseRedisURL(connStr)
	svc, err := sessionredis.NewRedisSessionService(sessionredis.RedisSessionServiceConfig{
		Addr:     addr,
		Password: password,
		DB:       db,
		TTL:      ttl,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis session service: %w", err)
	}
	return svc, nil
}

// parseRedisURL splits a redis:// connection string into the host:port address,
// password, and database number that the Redis client needs.
//
// TODO: This lives here because adk-utils-go/session/redis expects individual
// fields (Addr, Password, DB) instead of a connection string. Once that library
// accepts a connection string directly, this function can be removed.
func parseRedisURL(rawURL string) (addr, password string, db int) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, "", 0
	}
	addr = u.Host
	if addr == "" {
		addr = "localhost:6379"
	}
	if !strings.Contains(addr, ":") {
		addr += ":6379"
	}
	if u.User != nil {
		password, _ = u.User.Password()
	}
	if len(u.Path) > 1 {
		if n, err := strconv.Atoi(u.Path[1:]); err == nil {
			db = n
		}
	}
	return
}

// createMemoryService returns the long-term memory backend based on global settings.
// Returns nil if no provider is configured.
func createMemoryService(ctx context.Context, settings store.Settings, memoryProviders map[string]store.MemoryProvider, backends map[string]store.BackendDefinition) (memory.Service, error) {
	if settings.LongTermProvider == "" {
		return nil, nil
	}

	provider, ok := memoryProviders[settings.LongTermProvider]
	if !ok {
		return nil, nil
	}

	connStr, _ := provider.Config["connectionString"].(string)
	if connStr == "" {
		return nil, nil
	}

	if provider.Embedding == nil || provider.Embedding.Backend == "" {
		return nil, nil
	}

	embeddingBackend, ok := backends[provider.Embedding.Backend]
	if !ok {
		return nil, nil
	}

	svc, err := memorypostgres.NewPostgresMemoryService(ctx, memorypostgres.PostgresMemoryServiceConfig{
		ConnString: connStr,
		EmbeddingModel: memorypostgres.NewOpenAICompatibleEmbedding(memorypostgres.OpenAICompatibleEmbeddingConfig{
			BaseURL: embeddingBackend.URL,
			Model:   provider.Embedding.Model,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Postgres memory service: %w", err)
	}
	return svc, nil
}

// mergeHeaders combines backend-level and agent-level headers into an http.Header.
// Agent-level headers override backend-level headers when the same key is set.
func mergeHeaders(backendHeaders, agentHeaders map[string]string) http.Header {
	if len(backendHeaders) == 0 && len(agentHeaders) == 0 {
		return nil
	}
	h := make(http.Header)
	for k, v := range backendHeaders {
		h.Set(k, v)
	}
	for k, v := range agentHeaders {
		h.Set(k, v)
	}
	return h
}

// createLLM instantiates the language model client for a backend definition.
// Supports OpenAI-compatible, Anthropic, and Gemini backends.
func createLLM(ctx context.Context, backend store.BackendDefinition, llmRef store.BackendRef) (model.LLM, error) {
	headers := mergeHeaders(backend.Headers, llmRef.Headers)

	switch backend.Type {
	case config.BackendTypeOpenAI:
		return genaiopenai.New(genaiopenai.Config{
			APIKey:    backend.APIKey,
			BaseURL:   backend.URL,
			ModelName: llmRef.Model,
			HTTPOptions: genaiopenai.HTTPOptions{
				Headers: headers,
			},
		}), nil

	case config.BackendTypeAnthropic:
		return genaianthro.New(genaianthro.Config{
			APIKey:    backend.APIKey,
			ModelName: llmRef.Model,
			HTTPOptions: genaianthro.HTTPOptions{
				Headers: headers,
			},
		}), nil

	case config.BackendTypeGemini:
		return gemini.NewModel(ctx, llmRef.Model, &genai.ClientConfig{
			APIKey: backend.APIKey,
			HTTPOptions: genai.HTTPOptions{
				Headers: headers,
			},
		})

	default:
		return nil, fmt.Errorf("unsupported LLM backend type: %s", backend.Type)
	}
}

// buildToolsets assembles all tool providers for an agent: memory tools
// (search/save) if the agent has long-term memory, plus any MCP server
// toolsets referenced by name.
func buildToolsets(agentDef store.AgentDefinition, mcpServerMap map[string]store.MCPServer, memorySvc memory.Service) ([]tool.Toolset, error) {
	var toolsets []tool.Toolset

	if memorySvc != nil {
		ts, err := toolsmemory.NewToolset(toolsmemory.ToolsetConfig{
			MemoryService: memorySvc,
			AppName:       agentDef.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create memory toolset: %w", err)
		}
		toolsets = append(toolsets, ts)
	}

	for _, mcpName := range agentDef.MCPServers {
		srv, ok := mcpServerMap[mcpName]
		if !ok {
			continue
		}
		transport, err := createMCPTransport(&srv)
		if err != nil {
			return nil, fmt.Errorf("failed to create MCP transport %q: %w", srv.Name, err)
		}
		ts, err := mcptoolset.New(mcptoolset.Config{
			Transport: transport,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create MCP toolset %q: %w", srv.Name, err)
		}
		toolsets = append(toolsets, ts)
	}

	return toolsets, nil
}

// createMCPTransport returns the appropriate MCP transport (stdio subprocess
// or HTTP/SSE) for a server definition.
func createMCPTransport(srv *store.MCPServer) (mcp.Transport, error) {
	switch srv.Type {
	case "stdio":
		if srv.Command == "" {
			return nil, fmt.Errorf("stdio transport requires 'command' field")
		}
		cmd := exec.Command(srv.Command, srv.Args...)
		if srv.WorkDir != "" {
			cmd.Dir = srv.WorkDir
		}
		if len(srv.Env) > 0 {
			cmd.Env = os.Environ()
			for k, v := range srv.Env {
				cmd.Env = append(cmd.Env, k+"="+v)
			}
		}
		return &mcp.CommandTransport{Command: cmd}, nil

	case "http", "":
		if srv.Endpoint == "" {
			return nil, fmt.Errorf("http transport requires 'endpoint' field")
		}
		return &mcp.StreamableClientTransport{
			Endpoint:   srv.Endpoint,
			HTTPClient: httpClientForMCP(srv.Headers, srv.Insecure),
			MaxRetries: 5,
		}, nil

	default:
		return nil, fmt.Errorf("unknown MCP transport type: %s", srv.Type)
	}
}

// httpClientForMCP returns an HTTP client configured with custom headers
// and optional TLS verification skip for MCP servers.
func httpClientForMCP(headers map[string]string, insecure bool) *http.Client {
	base := http.DefaultTransport
	if insecure {
		base = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	if len(headers) == 0 && !insecure {
		return http.DefaultClient
	}
	if len(headers) == 0 {
		return &http.Client{Transport: base}
	}
	return &http.Client{
		Transport: &headerTransport{
			base:    base,
			headers: headers,
		},
	}
}

// headerTransport is an http.RoundTripper that injects fixed headers
// into every outgoing request.
type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

// RoundTrip adds the configured headers and delegates to the base transport.
func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return t.base.RoundTrip(req)
}

// buildInstruction assembles the system prompt for an agent. It starts with
// the agent's custom prompt (or a default), appends memory instructions if
// long-term memory is enabled, and appends any MCP server system prompts, then skills.
func buildInstruction(agentDef store.AgentDefinition, mcpServerMap map[string]store.MCPServer, skillMap map[string]store.Skill, skillsBaseDir string, memorySvc memory.Service) string {
	instruction := baseInstruction
	if agentDef.SystemPrompt != "" {
		instruction = agentDef.SystemPrompt
	}

	if memorySvc != nil {
		instruction += memoryInstruction
	}

	instruction += artifactInstruction

	for _, mcpName := range agentDef.MCPServers {
		if srv, ok := mcpServerMap[mcpName]; ok && srv.SystemPrompt != "" {
			instruction += "\n\n" + srv.SystemPrompt
		}
	}

	for _, skillID := range agentDef.Skills {
		sk, ok := skillMap[skillID]
		if !ok {
			continue
		}
		instruction += "\n\n--- Skill: " + sk.Name + " ---\n" + sk.Instructions
		for _, ref := range sk.References {
			content, err := os.ReadFile(filepath.Join(skillsBaseDir, skillID, ref.Filename))
			if err != nil {
				slog.Warn("Failed to read skill reference", "skill", sk.Name, "file", ref.Filename, "error", err)
				continue
			}
			instruction += "\n\n[Reference: " + ref.Filename + "]\n" + string(content)
		}
	}

	return instruction
}

var stateVarRegex = regexp.MustCompile(`\{\{agent\.output:([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)

func makeInstructionProvider(template string) llmagent.InstructionProvider {
	if !stateVarRegex.MatchString(template) {
		return func(_ agent.ReadonlyContext) (string, error) {
			return template, nil
		}
	}
	return func(ctx agent.ReadonlyContext) (string, error) {
		return stateVarRegex.ReplaceAllStringFunc(template, func(match string) string {
			sub := stateVarRegex.FindStringSubmatch(match)
			if len(sub) < 2 {
				return match
			}
			val, err := ctx.ReadonlyState().Get(sub[1])
			if err != nil || val == nil {
				return ""
			}
			return fmt.Sprintf("%v", val)
		}), nil
	}
}
