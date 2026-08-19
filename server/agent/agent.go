// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/artifact"
	"google.golang.org/adk/v2/memory"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/model/gemini"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/server/adkrest"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/mcptoolset"
	"google.golang.org/adk/v2/tool/skilltoolset"
	"google.golang.org/genai"

	"github.com/1set/starlet"
	artifactfs "github.com/achetronic/adk-utils-go/artifact/filesystem"
	genaianthro "github.com/achetronic/adk-utils-go/genai/anthropic"
	genaiopenai "github.com/achetronic/adk-utils-go/genai/openai/completions"
	memorypostgres "github.com/achetronic/adk-utils-go/memory/postgres"
	"github.com/achetronic/adk-utils-go/plugin/contextguard"
	sessionredis "github.com/achetronic/adk-utils-go/session/redis"
	toolsmemory "github.com/achetronic/adk-utils-go/tools/memory"

	"github.com/achetronic/magec/server/agent/runrecorder"
	"github.com/achetronic/magec/server/agent/secrets"
	toolsartifacts "github.com/achetronic/magec/server/agent/tools/artifacts"
	toolsflowstate "github.com/achetronic/magec/server/agent/tools/flowstate"
	toolsskills "github.com/achetronic/magec/server/agent/tools/skills"
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
- Use 'load_artifact' to read a previously saved artifact, or to access a file the user attached in an earlier message (these are listed in the MAGEC_ATTACHED_ARTIFACTS block). After calling it, the artifact contents arrive on the next turn as a native multimodal attachment that you can read directly — you do NOT need to decode any base64 yourself.
- Use 'list_artifacts' to see all artifacts in the current session.
- Use 'export_artifact' when you need to write an artifact's bytes to a file on the local filesystem so other tools can read it. The tool returns the absolute path of the resulting file; pass that path to whatever filesystem-aware tool needs it.
- Use 'get_artifact_url' when the consumer of the artifact runs in a different process or container and cannot reach the local filesystem (for example, a remote tool that fetches files over HTTP). The tool returns a short-lived signed URL that serves the artifact's raw bytes without authentication; pass that URL to the consumer.

IMPORTANT: When generating code files, long documents, configuration files, scripts, or any substantial structured content, ALWAYS use save_artifact instead of pasting it in the chat. The artifact will be delivered to the user as a downloadable file automatically.`

const flowStateInstruction = `
You are running inside a multi-agent workflow. You have shared state tools available to coordinate with the other agents in this flow:
- Use 'set_state' to record a value (string, number, boolean, list, or object) under a key. Other agents in the same workflow can read it later in the same conversation.
- Use 'get_state' to read a value previously stored by another agent (or by an earlier turn of yours). It returns {found: true, value: ...} when present and {found: false} when absent.

Use shared state for orchestration signals (e.g. an approval flag, a quality score, a list of pending items), not for bulky content — keep large outputs in artifacts.`

// secretsInstructionHeader precedes the list of secret keys an agent may use.
const secretsInstructionHeader = `
You have access to stored secrets through placeholders. Write {{secret:KEY}} wherever a tool argument needs the secret's value (an API key, a token, a password) and the real value is substituted right before the tool runs. You never see the real values and must never try to print them.
Available secret keys:`

// Service wraps the ADK REST handler that serves all configured agents.
// Incoming requests are routed to the correct agent by the appName field.
type Service struct {
	handler     http.Handler
	sessionSvc  session.Service
	memorySvc   memory.Service
	artifactSvc artifact.Service
	adkAgents   map[string]agent.Agent
}

// New builds an ADK agent for every AgentDefinition in the store, wires up
// their LLM, session, memory, and MCP toolsets, and returns a Service that
// routes requests to the right agent based on the appName in the request body.
// Any FlowDefinitions are translated into ADK workflow agents and registered
// alongside the regular agents.
//
// tempDirProvider returns the directory used by tools that need a transient
// filesystem location (export_artifact, etc.). The caller is the single
// source of truth for that path — agent.New does not perform any fallback.
//
// skillsDir is the absolute path to the directory that holds every
// skill package (one sub-directory per slug). Agents with non-empty
// AgentDefinition.Skills get a per-agent skilltoolset rooted at that
// directory; pass "" to disable skill loading entirely.
//
// artifactURLBuilder mints short-lived signed URLs for artifacts (consumed
// by the get_artifact_url tool). May be nil; when nil the tool is not
// registered, so deployments that have not configured the signing secret do
// not advertise a capability that always fails.
func New(ctx context.Context, agents []store.AgentDefinition, backends []store.BackendDefinition, memoryProviders []store.MemoryProvider, mcpServers []store.MCPServer, skills []store.Skill, flows []store.FlowDefinition, storeSecrets []store.Secret, settings store.Settings, registry contextguard.ModelRegistry, tempDirProvider func() string, skillsDir string, artifactURLBuilder toolsartifacts.ArtifactURLBuilder, recorder *runrecorder.Recorder) (*Service, error) {
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

	secretsSnap := secrets.NewSnapshot(storeSecrets)

	mcpServerMap := make(map[string]store.MCPServer, len(mcpServers))
	for _, m := range mcpServers {
		mcpServerMap[m.ID] = m
	}

	// Skills are referenced by ID in the agent definition but loaded
	// from disk by slug, so we only need an ID -> slug index. Any
	// other skill metadata lives in SKILL.md and is read by the
	// skilltoolset when the LLM calls list_skills/load_skill.
	skillSlugIndex := make(map[string]string, len(skills))
	for _, sk := range skills {
		if sk.Slug != "" {
			skillSlugIndex[sk.ID] = sk.Slug
		}
	}

	sessionSvc, err := createSessionService(settings, memoryProviderMap)
	if err != nil {
		return nil, fmt.Errorf("session service: %w", err)
	}

	memorySvc, err := createMemoryService(ctx, settings, memoryProviderMap, backendMap)
	if err != nil {
		return nil, fmt.Errorf("memory service: %w", err)
	}

	artifactSvc, err := artifactfs.NewFilesystemService(artifactfs.FilesystemServiceConfig{
		BasePath: filepath.Join("data", "artifacts"),
	})
	if err != nil {
		return nil, fmt.Errorf("artifact service: %w", err)
	}

	baseTset, err := newBaseToolset(tempDirProvider, artifactURLBuilder)
	if err != nil {
		return nil, fmt.Errorf("failed to create base toolset: %w", err)
	}

	adkAgentMap := make(map[string]agent.Agent, len(agents)+len(flows))
	agentDefMap := make(map[string]store.AgentDefinition, len(agents))
	llmMap := make(map[string]model.LLM, len(agents))
	var rootAgent agent.Agent
	var otherAgents []agent.Agent

	// Build ADK agents
	for i, agentDef := range agents {
		adkAgent, llmModel, err := BuildAgentInstance(BuildAgentInstanceParams{
			Ctx:             ctx,
			AgentDef:        agentDef,
			BackendMap:      backendMap,
			MCPServerMap:    mcpServerMap,
			SkillSlugs:      skillSlugIndex,
			SkillsDir:       skillsDir,
			MemorySvc:       memorySvc,
			BaseToolset:     baseTset,
			SecretsSnapshot: secretsSnap,
		})
		if err != nil {
			return nil, fmt.Errorf("agent %q: %w", agentDef.ID, err)
		}
		llmMap[agentDef.ID] = llmModel
		adkAgentMap[agentDef.ID] = adkAgent
		agentDefMap[agentDef.ID] = agentDef

		if i == 0 {
			rootAgent = adkAgent
		} else {
			otherAgents = append(otherAgents, adkAgent)
		}
		slog.Info("Agent initialized", "id", agentDef.ID, "name", agentDef.Name)
	}

	flowStateToolset, err := toolsflowstate.NewToolset()
	if err != nil {
		return nil, fmt.Errorf("failed to create flow_state toolset: %w", err)
	}

	flowDefMap := make(map[string]store.FlowDefinition, len(flows))
	for _, f := range flows {
		flowDefMap[f.ID] = f
	}

	// Build the enabled starlet module loader list once: all built-in modules
	// minus whatever the admin disabled. Shared safely across code-node runs
	// (each run builds its own fresh Machine). If the loader list cannot be
	// built (e.g. an unknown name in DisabledLibraries), log and fall back to
	// an empty list so the server still starts.
	allNames := starlet.GetAllBuiltinModuleNames()
	disabledSet := make(map[string]bool, len(settings.Flows.DisabledLibraries))
	for _, n := range settings.Flows.DisabledLibraries {
		disabledSet[n] = true
	}
	enabledNames := make([]string, 0, len(allNames))
	for _, n := range allNames {
		if !disabledSet[n] {
			enabledNames = append(enabledNames, n)
		}
	}
	starletLoaders, loaderErr := starlet.MakeBuiltinModuleLoaderList(enabledNames...)
	if loaderErr != nil {
		slog.Warn("Failed to build Starlark module loader list; code nodes will have no modules", "error", loaderErr)
		starletLoaders = nil
	}

	flowDeps := FlowBuildDeps{
		Ctx:              ctx,
		AgentDefs:        agentDefMap,
		FlowDefs:         flowDefMap,
		BackendMap:       backendMap,
		MCPServerMap:     mcpServerMap,
		SkillSlugs:       skillSlugIndex,
		SkillsDir:        skillsDir,
		MemorySvc:        memorySvc,
		BaseToolset:      baseTset,
		FlowStateToolset: flowStateToolset,
		StarletLoaders:   starletLoaders,
		FlowsSettings:    settings.Flows,
		SecretsSnapshot:  secretsSnap,
	}

	// Build flows
	for _, flow := range flows {
		flowAgent, err := BuildFlowAgent(flow, flowDeps)
		if err != nil {
			slog.Warn("Failed to build flow", "flow", flow.Name, "error", err)
			continue
		}
		otherAgents = append(otherAgents, flowAgent)
		adkAgentMap[flow.ID] = flowAgent
		slog.Info("Flow initialized", "id", flow.ID, "name", flow.Name)
	}

	// Snapshot every built flow's node ID -> type map for the run recorder,
	// so a run's audit record states what each node was at execution time
	// even if the flow is edited later. Subflow nodes are folded in because
	// their events surface inside the parent flow's run.
	if recorder != nil {
		nodeTypes := make(map[string]map[string]string, len(flows))
		for _, flow := range flows {
			if _, ok := adkAgentMap[flow.ID]; !ok {
				continue
			}
			nodeTypes[flow.ID] = collectNodeTypes(flow, flowDefMap, map[string]bool{})
		}
		recorder.SetNodeTypes(nodeTypes)
	}

	loader, err := agent.NewMultiLoader(rootAgent, otherAgents...)
	if err != nil {
		return nil, fmt.Errorf("failed to create multi-loader: %w", err)
	}

	restCfg := adkrest.ServerConfig{
		SessionService:  sessionSvc,
		AgentLoader:     loader,
		ArtifactService: artifactSvc,
		MemoryService:   memorySvc,
		SSEWriteTimeout: 15 * time.Minute,
	}

	if registry != nil {
		restCfg.PluginConfig = buildContextGuardConfig(agents, llmMap, registry)
	}
	// The model-boundary secret guard redacts known secret values from every
	// LLM request, whatever path they took into it.
	if guardPlugin, err := secretsSnap.Plugin(); err != nil {
		slog.Warn("Failed to build secretguard plugin; model requests will not be redacted", "error", err)
	} else {
		restCfg.PluginConfig.Plugins = append(restCfg.PluginConfig.Plugins, guardPlugin)
	}
	// The run recorder plugin coexists with contextguard: PluginConfig.Plugins
	// is a slice and the runner invokes every plugin's callbacks.
	if recorder != nil {
		recorderPlugin, err := recorder.Plugin()
		if err != nil {
			slog.Warn("Failed to build runrecorder plugin; runs will not be audited", "error", err)
		} else {
			restCfg.PluginConfig.Plugins = append(restCfg.PluginConfig.Plugins, recorderPlugin)
		}
	}

	restServer, err := adkrest.NewServer(restCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ADK REST server: %w", err)
	}

	return &Service{
		handler:     restServer,
		sessionSvc:  sessionSvc,
		memorySvc:   memorySvc,
		artifactSvc: artifactSvc,
		adkAgents:   adkAgentMap,
	}, nil
}

// collectNodeTypes flattens a flow's node ID -> node type map, recursing into
// subflows since their nodes emit events inside the parent flow's run. The
// parent's own IDs win on collision; visited guards against flow cycles.
func collectNodeTypes(flow store.FlowDefinition, flowDefs map[string]store.FlowDefinition, visited map[string]bool) map[string]string {
	types := map[string]string{}
	if visited[flow.ID] {
		return types
	}
	visited[flow.ID] = true

	for _, node := range flow.Nodes {
		if node.Type != store.FlowNodeSubflow {
			continue
		}
		sub, ok := flowDefs[node.FlowID]
		if !ok {
			continue
		}
		for id, t := range collectNodeTypes(sub, flowDefs, visited) {
			types[id] = t
		}
	}
	for _, node := range flow.Nodes {
		types[node.ID] = node.Type
	}
	return types
}

// sortFlowsTopologically performs a topological sort on the flow definitions.
// This detects circular dependencies and ensures that sub-flows are constructed
// and registered before the parent flows that depend on them.
func sortFlowsTopologically(flows []store.FlowDefinition) ([]store.FlowDefinition, error) {
	if len(flows) == 0 {
		return flows, nil
	}

	flowMap := make(map[string]store.FlowDefinition, len(flows))
	for _, f := range flows {
		flowMap[f.ID] = f
	}

	var sorted []store.FlowDefinition
	visited := make(map[string]bool)
	temp := make(map[string]bool)

	var visit func(id string) error
	visit = func(id string) error {
		if temp[id] {
			return fmt.Errorf("circular dependency detected at flow %s", id)
		}
		if visited[id] {
			return nil
		}
		temp[id] = true

		f, ok := flowMap[id]
		if !ok {
			temp[id] = false
			visited[id] = true
			return nil
		}

		deps := f.AgentIDs()
		for _, dep := range deps {
			if _, isFlow := flowMap[dep]; isFlow {
				if err := visit(dep); err != nil {
					return err
				}
			}
		}

		temp[id] = false
		visited[id] = true
		sorted = append(sorted, f)
		return nil
	}

	for _, f := range flows {
		if !visited[f.ID] {
			if err := visit(f.ID); err != nil {
				return nil, err
			}
		}
	}

	return sorted, nil
}

// BuildAgentInstanceParams bundles the inputs to BuildAgentInstance. Using a
// struct keeps the call site readable as the parameter list grew (per-flow
// extra toolsets, custom instance names) and lets the flow builder add
// scope-dependent extras without rippling signature changes through the
// caller.
type BuildAgentInstanceParams struct {
	Ctx          context.Context
	AgentDef     store.AgentDefinition
	BackendMap   map[string]store.BackendDefinition
	MCPServerMap map[string]store.MCPServer
	// SkillSlugs maps every store-registered skill ID to its on-disk
	// slug (= directory name under SkillsDir, also the SKILL.md
	// frontmatter `name`). The agent builder uses it to translate
	// AgentDef.Skills (a list of IDs) into the slug whitelist that
	// scopes the per-agent skilltoolset to ADK's expected file layout.
	SkillSlugs map[string]string
	// SkillsDir is the absolute path that contains every skill
	// package (typically Store.SkillsDir(), i.e. data/skills/). The
	// builder feeds it through os.DirFS into ADK's filesystem source
	// so the LLM only ever sees on-disk SKILL.md files — no JSON
	// shadow copy in the store, no config drift between disk and
	// admin API.
	SkillsDir   string
	MemorySvc   memory.Service
	BaseToolset tool.Toolset
	// InstanceName is the ADK agent name to register. Defaults to AgentDef.ID
	// for the standalone catalogue copy. Flow builders pass a flow-scoped
	// unique name so the same logical agent can appear multiple times across
	// the workflow tree without violating ADK's single-parent constraint.
	InstanceName string
	// ExtraToolsets are additional toolsets injected on top of BaseToolset.
	// The flow builder uses this to wire flow-only capabilities (shared state
	// tools, exit_loop) without altering the standalone tool catalogue. May
	// be nil; passed through as-is.
	ExtraToolsets []tool.Toolset
	// ExtraTools is a flat list of individual tools wrapped into a single
	// throw-away toolset and appended after ExtraToolsets. Used for scope-
	// specific singletons such as exit_loop, which only makes sense inside a
	// loop-with-exitLoop subtree. May be nil.
	ExtraTools []tool.Tool
	// IncludeFlowStateInstruction appends the flow_state usage paragraph to
	// the agent instruction. Set by the flow builder for every agent inside
	// a flow so the model knows it can call set_state and get_state.
	IncludeFlowStateInstruction bool
	// SecretsSnapshot carries the store secrets. When the agent has secrets
	// allowed, its MCP toolsets are wrapped for placeholder expansion and the
	// instruction lists the available keys. May be nil.
	SecretsSnapshot *secrets.Snapshot
}

// BuildAgentInstance constructs an individual ADK agent instance from its
// definition. It resolves the associated LLM backend, assembles its toolsets
// (MCPs, skills, memory, base, plus any flow-scoped extras), and builds its
// persona/instruction context.
//
// This function is callable from outside the package because flow.go invokes
// it once per agent appearance inside a flow tree, with extra toolsets that
// only apply to that specific appearance.
func BuildAgentInstance(p BuildAgentInstanceParams) (agent.Agent, model.LLM, error) {
	llmBackend, ok := p.BackendMap[p.AgentDef.LLM.Backend]
	if !ok {
		return nil, nil, fmt.Errorf("LLM backend %q not found", p.AgentDef.LLM.Backend)
	}
	llmModel, err := createLLM(p.Ctx, llmBackend, p.AgentDef.LLM)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create LLM: %w", err)
	}

	toolsets, err := buildToolsets(p.AgentDef, p.MCPServerMap, p.MemorySvc)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build toolsets: %w", err)
	}

	// Secrets expand only inside MCP tool calls (the toolsets built above),
	// scoped to this agent's allowlist. Internal toolsets (memory, artifacts,
	// skills, flow state) keep placeholders as-is by design.
	var allowedSecrets map[string]string
	if p.SecretsSnapshot != nil && (p.AgentDef.AllowAllSecrets || len(p.AgentDef.Secrets) > 0) {
		allowedSecrets = p.SecretsSnapshot.Map(p.AgentDef.Secrets, p.AgentDef.AllowAllSecrets)
		for i, ts := range toolsets {
			toolsets[i] = p.SecretsSnapshot.WrapToolset(ts, allowedSecrets)
		}
	}

	toolsets = append(toolsets, p.BaseToolset)
	if skillTs, err := buildSkillToolset(p.Ctx, p.AgentDef, p.SkillSlugs, p.SkillsDir); err != nil {
		return nil, nil, fmt.Errorf("failed to build skill toolset: %w", err)
	} else if skillTs != nil {
		toolsets = append(toolsets, skillTs)
	}
	toolsets = append(toolsets, p.ExtraToolsets...)
	if len(p.ExtraTools) > 0 {
		toolsets = append(toolsets, &flatToolset{tools: p.ExtraTools})
	}

	instruction := buildInstruction(p.AgentDef, p.MCPServerMap, p.MemorySvc)
	if p.IncludeFlowStateInstruction {
		instruction += flowStateInstruction
	}
	if len(allowedSecrets) > 0 {
		keys := make([]string, 0, len(allowedSecrets))
		for k := range allowedSecrets {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		instruction += secretsInstructionHeader
		for _, k := range keys {
			instruction += "\n- " + k
		}
	}

	name := p.InstanceName
	if name == "" {
		name = p.AgentDef.ID
	}

	agentCfg := llmagent.Config{
		Name:                name,
		Model:               llmModel,
		Description:         p.AgentDef.Name,
		InstructionProvider: makeInstructionProvider(instruction),
		Toolsets:            toolsets,
		OutputKey:           p.AgentDef.OutputKey,
	}

	adkAgent, err := llmagent.New(agentCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return adkAgent, llmModel, nil
}

// flatToolset is a minimal tool.Toolset adaptor that exposes a fixed list
// of pre-built tools. We use it inside BuildAgentInstance to fold one-off
// tools (e.g. exit_loop wired contextually by the flow builder) into the
// toolset list without forcing every caller to invent a stateful toolset.
type flatToolset struct {
	tools []tool.Tool
}

func (f *flatToolset) Name() string { return "flow_extra_tools" }

func (f *flatToolset) Tools(_ agent.ReadonlyContext) ([]tool.Tool, error) {
	return f.tools, nil
}

// buildContextGuardConfig generates the ContextGuard plugin configuration
// from all agent definitions that have it enabled, enforcing their specific
// max-token or sliding-window boundaries on the LLM prompt size.
func buildContextGuardConfig(agents []store.AgentDefinition, llmMap map[string]model.LLM, registry contextguard.ModelRegistry) runner.PluginConfig {
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
	return guard.PluginConfig()
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

// ArtifactService returns the artifact.Service that backs the save_artifact /
// load_artifact tools. Clients use it to persist user-uploaded files that are
// too large to embed inline in the /run_sse request.
func (s *Service) ArtifactService() artifact.Service {
	return s.artifactSvc
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

// dialectFor selects the provider dialect from the backend URL. OpenAI's own
// API and generic compatible servers (Ollama, vLLM, ...) get nil, which keeps
// the adapter OpenAI-pure.
func dialectFor(baseURL string) genaiopenai.Dialect {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}
	switch strings.ToLower(u.Hostname()) {
	case "openrouter.ai":
		return genaiopenai.OpenRouter
	case "api.deepseek.com":
		return genaiopenai.DeepSeek
	default:
		return nil
	}
}

// createLLM instantiates the language model client for a backend definition.
// Supports OpenAI-compatible, Anthropic, and Gemini backends.
func createLLM(ctx context.Context, backend store.BackendDefinition, llmRef store.BackendRef) (model.LLM, error) {
	headers := mergeHeaders(backend.Headers, llmRef.Headers)

	switch backend.Type {
	case config.BackendTypeOpenAI:
		switch backend.API {
		case "", config.BackendAPICompletions:
			return genaiopenai.New(genaiopenai.Config{
				APIKey:    backend.APIKey,
				BaseURL:   backend.URL,
				ModelName: llmRef.Model,
				Dialect:   dialectFor(backend.URL),
				HTTPOptions: genaiopenai.HTTPOptions{
					Headers: headers,
				},
			}), nil
		case config.BackendAPIResponses:
			// Placeholder: adk-utils-go will ship the Responses adapter next;
			// the field is already persisted so backends are ready for it.
			return nil, fmt.Errorf("responses API is not supported yet by adk-utils-go (backend %q)", backend.Name)
		default:
			return nil, fmt.Errorf("unsupported API %q for OpenAI-compatible backend %q (expected %q or %q)", backend.API, backend.Name, config.BackendAPICompletions, config.BackendAPIResponses)
		}

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

// buildSkillToolset returns ADK's skilltoolset scoped to the agent's
// linked skills, or nil when the agent has none. The toolset wraps
// data/skills/ in a per-agent fs.FS that filters out every directory
// not whitelisted by AgentDef.Skills, so list_skills/load_skill never
// surface a skill the operator didn't enable for this agent.
//
// Skills are referenced by ID in the agent definition (UUIDs are stable
// across renames, slugs are not), so we translate IDs to slugs through
// the SkillSlugs map before building the whitelist. Unknown IDs are
// skipped silently — the admin UI guarantees integrity, and a stale
// reference is not worth aborting agent construction.
func buildSkillToolset(ctx context.Context, agentDef store.AgentDefinition, slugIndex map[string]string, skillsDir string) (tool.Toolset, error) {
	if len(agentDef.Skills) == 0 || skillsDir == "" {
		return nil, nil
	}
	allowed := make([]string, 0, len(agentDef.Skills))
	for _, id := range agentDef.Skills {
		if slug, ok := slugIndex[id]; ok && slug != "" {
			allowed = append(allowed, slug)
		}
	}
	if len(allowed) == 0 {
		return nil, nil
	}

	root := os.DirFS(skillsDir)
	agentFS := toolsskills.NewAgentFS(root, allowed)
	// Wrap the per-agent FS in a TolerantSource so SKILL.md files
	// with extra non-canonical frontmatter keys (`version:`, `author:`
	// …) don't blow up ListFrontmatters and abort the whole LLM
	// request. Decision #29 requires the runtime to keep working
	// even when individual skills don't strictly satisfy ADK's
	// `KnownFields(true)` parser.
	src := toolsskills.NewTolerantSource(agentFS)
	ts, err := skilltoolset.New(ctx, skilltoolset.Config{
		Source: src,
	})
	if err != nil {
		return nil, fmt.Errorf("create skilltoolset: %w", err)
	}
	return ts, nil
}

// buildToolsets assembles all tool providers for an agent: memory tools
// (search/save) if the agent has long-term memory, plus any MCP server
// toolsets referenced by name. Skills are intentionally NOT included
// here — they live in their own scope-aware toolset built by
// buildSkillToolset and appended after the base toolset (decision #29).
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
			slog.Warn("Agent references an MCP server that does not exist; its tools will not be available",
				"agent", agentDef.ID, "mcp", mcpName)
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
		return &stdioCommandTransport{srv: srv}, nil

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

// stdioCommand launches a fresh subprocess for every MCP connection.
// The ADK's mcptoolset reconnects when a session dies, but mcp.CommandTransport
// reuses the same *exec.Cmd — and exec.Cmd is single-use, so any reconnect
// after the first fails with "exec: Stdout already set" (issue #70). Holding
// the server definition instead of a command instance keeps reconnects working
// for the lifetime of the toolset.
type stdioCommandTransport struct {
	srv *store.MCPServer
}

// stdioCommand builds the command for one subprocess run from the server
// definition.
func stdioCommand(srv *store.MCPServer) *exec.Cmd {
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
	return cmd
}

// Connect starts a new subprocess and returns its connection.
func (t *stdioCommandTransport) Connect(ctx context.Context) (mcp.Connection, error) {
	return (&mcp.CommandTransport{Command: stdioCommand(t.srv)}).Connect(ctx)
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

// buildInstruction assembles the system prompt for an agent. It starts
// with the agent's custom prompt (or a default), appends memory and
// artifact instructions, and finally appends every linked MCP server's
// system prompt. Skills are NOT inlined here — they are exposed as
// callable tools through the per-agent skilltoolset (decision #29) so
// the LLM only loads the specific skill it needs, when it needs it.
func buildInstruction(agentDef store.AgentDefinition, mcpServerMap map[string]store.MCPServer, memorySvc memory.Service) string {
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
