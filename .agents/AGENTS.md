# AGENTS.md - Magec

Self-hosted multi-agent AI platform with voice, visual workflows, and tool integration.

## Project Overview

**Magec** is a multi-agent AI platform that runs on your server. Named after the Guanche god of the Sun (/maˈxek/), it provides:

- **Multi-agent system**: Per-agent LLM, memory, voice, and tools. Hot-reload from the Admin UI.
- **Agentic Flows**: Visual drag-and-drop editor. Sequential, parallel, loop, nested.
- **Any LLM backend**: OpenAI, Anthropic, Gemini, Ollama, or any OpenAI-compatible API.
- **MCP tools**: Home Assistant, GitHub, databases, and hundreds more via Model Context Protocol. HTTP headers and TLS skip supported.
- **Memory**: Session (Redis) + long-term semantic (PostgreSQL/pgvector).
- **Voice**: Wake word, VAD, STT, TTS. All server-side via ONNX Runtime. Privacy-first.
- **Clients**: Voice UI (PWA), Admin UI, Telegram, Discord, Slack, webhooks, cron, REST API.
- **A2A protocol**: Expose agents/flows as A2A-compatible endpoints for inter-agent communication.
- **Context guard**: Automatic context window management with LLM-powered summarization.

### Clients

| Client       | Type       | Description                                                                                        |
| ------------ | ---------- | -------------------------------------------------------------------------------------------------- |
| **Voice UI** | `direct`   | Vue 3 PWA with voice/text chat, wake word detection, audio visualizer                              |
| **Telegram** | `telegram` | Text and voice messages. Emoji reactions, per-chat agent switching, response modes                 |
| **Discord**  | `discord`  | Gateway WebSocket (no public URL). DMs, @mentions, voice messages, reactions, ! commands           |
| **Slack**    | `slack`    | Socket Mode (WebSocket, no public URL). DMs, @mentions, audio clips. See `.agents/SLACK_CLIENT.md` |
| **Webhook**  | `webhook`  | HTTP endpoint for external integrations (fixed command or passthrough prompt)                      |
| **Cron**     | `cron`     | Scheduled task that fires a command against agents on a schedule                                   |

## Architecture

```
magec/
├── server/                     # Go backend
│   ├── main.go                 # HTTP server (:8080 user + :8081 admin), routing, middleware
│   ├── agent/
│   │   ├── agent.go            # Multi-agent ADK setup, MCP transport, memory tools, ContextGuard wiring
│   │   ├── flow.go             # Flow→ADK workflow agent builder (sequential/parallel/loop)
│   │   └── base_toolset.go     # Base toolset (currently empty, placeholder for future tools)
│   ├── api/
│   │   ├── admin/              # Admin REST API (CRUD for all resources)
│   │   │   ├── handler.go      # Router + helpers
│   │   │   ├── agents.go       # Agent CRUD + MCP linking
│   │   │   ├── backends.go     # Backend CRUD
│   │   │   ├── clients.go      # Client CRUD + /types (JSON Schema) + token regen
│   │   │   ├── commands.go     # Command CRUD
│   │   │   ├── skills.go       # Skill CRUD + reference file upload/download/delete
│   │   │   ├── memory.go       # Memory provider CRUD + ping + /types
│   │   │   ├── secrets.go      # Secrets CRUD (encrypted at rest)
│   │   │   ├── settings.go     # Global settings (session/longterm provider)
│   │   │   ├── flows.go        # Flow CRUD + recursive validation
│   │   │   ├── conversations.go # Conversation audit (list/get/delete/clear/stats/summary/pair/reset-session)
│   │   │   ├── backup.go       # Backup/restore (tar.gz of data/ directory)
│   │   │   ├── voice.go        # Voice provider types endpoint
│   │   │   └── docs/           # Generated swagger
│   │   └── user/               # User-facing REST API
│   │       ├── handlers.go     # Health, ClientInfo, Voice, Webhook swagger types
│   │       ├── doc.go          # Swagger metadata
│   │       ├── a2a_swagger.go  # A2A swagger documentation stubs
│   │       ├── adk_swagger.go  # ADK REST API swagger documentation stubs
│   │       └── docs/           # Generated swagger (userapi)
│   ├── a2a/                   # A2A protocol handler
│   │   └── handler.go          # Per-agent/flow JSON-RPC endpoints, agent cards, SSE streaming

│   ├── middleware/
│   │   ├── middleware.go       # AccessLog (httpsnoop), CORS, ClientAuth, AdminAuth (rate-limited)
│   │   ├── recorder.go         # ConversationRecorder + ConversationRecorderSSE (dual-perspective)
│   │   ├── flowfilter.go       # Flow response filtering by responseAgent
│   │   ├── normalize.go        # SnakeCaseNormalize: recursive snake_case→camelCase for /run and /run_sse
│   │   ├── sessionensure.go    # Idempotent session creation (prevents overwriting ContextGuard state)
│   │   └── sessionstate.go     # Seeds outputKey values into session state for {{agent.output:key}} resolution
│   ├── clients/                # Client type registry + runtime
│   │   ├── provider.go         # Provider interface: Type(), DisplayName(), ConfigSchema()
│   │   ├── registry.go         # Register(), ValidateConfig() with oneOf support
│   │   ├── executor.go         # RunClient() — executes commands against all allowedAgents
│   │   ├── direct/spec.go      # Direct provider (empty schema)
│   │   ├── telegram/           # Telegram bot (spec.go + bot.go)
│   │   ├── discord/            # Discord Gateway bot (spec.go + bot.go)
│   │   ├── slack/              # Slack Socket Mode bot (spec.go + bot.go)
│   │   ├── webhook/            # Webhook handler (spec.go + handler.go)
│   │   └── cron/               # Cron scheduler (spec.go + cron.go + scheduler.go)
│   ├── memory/                 # Extensible memory provider registry
│   │   ├── provider.go         # Provider interface, Category, HealthResult
│   │   ├── registry.go         # Register(), Get(), All(), ValidTypeForCategory()
│   │   ├── redis/redis.go      # Redis provider (session)
│   │   └── postgres/postgres.go # Postgres provider (longterm, pgvector)
│   ├── store/                  # In-memory store + JSON persistence
│   │   ├── store.go            # Load/Save, CRUD, migration chain, OnChange(), env var expansion
│   │   ├── types.go            # All entity types (MCPServer includes Headers + Insecure)
│   │   ├── crypto.go           # AES-256-GCM encryption/decryption for secrets (PBKDF2)
│   │   └── conversations.go    # ConversationStore (data/conversations.json)
│   ├── schema/validate.go      # JSON Schema validation (google/jsonschema-go)
│   ├── config/config.go        # YAML config parsing (server + voice + log)
│   ├── logging/logging.go      # Structured logging (slog)
│   ├── voice/                  # Server-side voice detection (ONNX) + voice provider registry
│   │   ├── provider.go         # TTSProvider/STTProvider interfaces, TTSRequest/STTRequest types
│   │   ├── registry.go         # Register(), Get(), All() — same pattern as clients/memory
│   │   ├── detector.go         # OpenWakeWord inference
│   │   ├── vad.go              # Silero VAD inference
│   │   ├── handler.go          # WebSocket handler for audio streaming
│   │   ├── resampler.go        # Audio resampling to 16kHz
│   │   ├── openai/openai.go    # OpenAI-compatible TTS+STT provider (/v1/audio/speech, /v1/audio/transcriptions)
│   │   └── gemini/gemini.go    # Gemini TTS+STT provider (generateContent + speechConfig / inlineData)
│   ├── frontend/               # Embedded UI dist files (//go:embed)
│   │   ├── embed.go
│   │   ├── admin-ui/           # Built admin UI (copied by Makefile)
│   │   └── voice-ui/           # Built voice UI (copied by Makefile)
│   └── models/                 # Embedded ONNX models
│       ├── embed.go
│       ├── wakeword/           # Wake word models
│       └── auxiliary/          # mel-spec, VAD, embeddings
├── frontend/
│   ├── admin-ui/               # Admin UI (Vue 3 + Vite + Tailwind v4 + Pinia)
│   │   ├── src/
│   │   │   ├── main.js         # Vue app entry with Pinia
│   │   │   ├── App.vue         # Layout, sidebar, global ConfirmDialog/Toast/SearchPalette
│   │   │   ├── style.css       # Tailwind v4 @theme (piedra/atlantico/lava/sol/arena)
│   │   │   ├── lib/api/        # Fetch wrapper + CRUD per resource
│   │   │   ├── lib/frontmatter.js # YAML frontmatter parser for skill cards (uses js-yaml)
│   │   │   ├── lib/stores/data.js # Pinia central store
│   │   │   ├── components/     # Shared: AppDialog, Card, Badge, FormInput, Icon, Toast, SegmentedControl, etc.
│   │   │   └── views/          # Entity views (backends/, memory/, mcps/, agents/, skills/,
│   │   │                       #   clients/, commands/, flows/, conversations/)
│   │   ├── vite.config.js      # Vue + Tailwind plugin + dev proxy to :8081
│   │   └── package.json        # vue, pinia, vuedraggable, marked, js-yaml, tailwindcss v4
│   └── voice-ui/               # Voice UI (Vue 3 + Vite + Tailwind v4 + Pinia)
│       ├── src/
│       │   ├── main.js         # Vue app entry
│       │   ├── App.vue         # Main app shell
│       │   ├── style.css       # Tailwind v4 @theme
│       │   ├── lib/            # config, audio/, api/, i18n/, session/, settings/, stores/
│       │   └── components/     # 14 components: AgentSwitcher, CentellaOrb, ChatMessage, etc.
│       └── package.json        # vue, pinia, tailwindcss v4
├── models/                     # Source ONNX models (copied to server/ at build time)
│   ├── wakeword/               # Wake word models + wakewords.yaml
│   └── auxiliary/              # Downloaded by scripts/download-model.go
├── scripts/
│   ├── download-model.go       # Model downloader
│   └── install.sh              # One-line installer (downloads docker-compose, --gpu flag)
├── docker/
│   ├── build/Dockerfile        # Multi-stage: frontend → models → ffmpeg → onnx → go build → distroless
│   └── compose/
│       ├── docker-compose.yaml # Single self-contained file (all local services)
│       └── config.yaml         # Default config (also embedded in Docker image)
├── website/                    # Documentation site (Hugo)
│   ├── hugo.toml               # Site config + sidebar navigation
│   ├── content/docs/           # Markdown docs (getting-started, install-*, configuration, etc.)
│   └── themes/magec/           # Custom Hugo theme (layouts, css, js, shortcodes)
├── config.example.yaml
├── Makefile
└── README.md
```

## HTTP Endpoints

### Main Server (port 8080) — User API

| Method    | Path                                                | Description                                                         |
| --------- | --------------------------------------------------- | ------------------------------------------------------------------- |
| GET/POST  | `/api/v1/agent/*`                                   | ADK REST API (sessions, run, events)                                |
| POST      | `/api/v1/agent/run_sse`                             | Run agent (SSE streaming) — all clients use this                    |
| POST      | `/api/v1/webhooks/{clientId}`                       | Webhook endpoint — Bearer token auth                                |
| POST      | `/api/v1/voice/{agentId}/speech`                    | TTS proxy (per-agent backend)                                       |
| POST      | `/api/v1/voice/{agentId}/transcription`             | STT proxy (per-agent backend)                                       |
| WebSocket | `/api/v1/voice/events`                              | Voice events stream (wake word + VAD)                               |
| GET       | `/api/v1/client/info`                               | Client info (paired status, allowed agents with type/nested agents) |
| POST      | `/api/v1/a2a/{agentID}`                             | A2A JSON-RPC endpoint (per-agent/flow)                              |
| GET       | `/api/v1/a2a/.well-known/agent-card.json`           | A2A global agent card discovery (all enabled agents)                |
| GET       | `/api/v1/a2a/{agentID}/.well-known/agent-card.json` | A2A per-agent card (no auth required)                               |
| GET       | `/api/v1/health`                                    | Health check                                                        |
| GET       | `/api/v1/swagger/`                                  | Swagger UI                                                          |
| GET       | `/`                                                 | Voice UI static files                                               |

### Admin Server (port 8081) — Admin API

| Method | Path                       | Description                                                                                                                                                                             |
| ------ | -------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| GET    | `/`                        | Admin UI static files                                                                                                                                                                   |
| GET    | `/api/v1/admin/auth/check` | Verify admin credentials (200 if valid)                                                                                                                                                 |
|        | **Backends**               | CRUD: `/backends`, `/backends/{id}`                                                                                                                                                     |
|        | **Memory**                 | CRUD: `/memory`, `/memory/{id}`, `/memory/types`, `/memory/{id}/health`                                                                                                                 |
|        | **MCP Servers**            | CRUD: `/mcps`, `/mcps/{id}`                                                                                                                                                             |
|        | **Skills**                 | CRUD: `/skills`, `/skills/{id}` + references: `/skills/{id}/references`, `/skills/{id}/references/{filename}` + package: `/skills/{id}/package`                                         |
|        | **Agents**                 | CRUD: `/agents`, `/agents/{id}`, `/agents/{id}/mcps`, `/agents/{id}/mcps/{mcpId}`                                                                                                       |
|        | **Clients**                | CRUD: `/clients`, `/clients/{id}`, `/clients/types`, `/clients/{id}/regenerate-token`                                                                                                   |
|        | **Commands**               | CRUD: `/commands`, `/commands/{id}`                                                                                                                                                     |
|        | **Flows**                  | CRUD: `/flows`, `/flows/{id}`                                                                                                                                                           |
|        | **Secrets**                | CRUD: `/secrets`, `/secrets/{id}` (GET never returns value)                                                                                                                             |
|        | **Settings**               | GET/PUT: `/settings` (global memory provider selection)                                                                                                                                 |
|        | **Conversations**          | `/conversations`, `/conversations/{id}`, `/conversations/clear`, `/conversations/stats`, `/conversations/{id}/summary`, `/conversations/{id}/pair`, `/conversations/{id}/reset-session` |
|        | **Backup**                 | GET `/settings/backup`, POST `/settings/restore` (tar.gz of data/)                                                                                                                      |
|        | **Voice**                  | GET `/voice/types` (registered voice providers with JSON Schemas)                                                                                                                       |

## Configuration

**Split model**:

- **`config.yaml`** — Server infrastructure only (ports, logging, voice/ONNX). Read at startup.
- **Admin API + Store** — All resources managed via Admin UI at `:8081`. Persisted to `data/store.json`.

```yaml
server:
  host: 0.0.0.0
  port: 8080
  adminPort: 8081
  # adminPassword: ""  # Admin API auth (Bearer token)
  # encryptionKey: ""  # Encrypt secrets at rest (AES-256-GCM, independent from adminPassword)
  # publicURL: ""       # Public URL for A2A agent cards (defaults to http://localhost:{port})

voice:
  ui:
    enabled: true
  onnxLibraryPath: "" # default: /usr/lib/libonnxruntime.so

log:
  level: info # debug, info, warn, error
  format: console # console, json
```

## Code Patterns

### Go Conventions

- **Store-based resources**: All entities managed via admin API, persisted to `data/store.json`
- **On first run**: Store starts empty. Configure via Admin UI at `:8081`
- **Multi-agent ADK**: `agent.New()` accepts agents + flows, `NewMultiLoader` routes by `appName`
- **Immutable UUID v4 IDs**: All entities use `google/uuid`. Cross-references store IDs, not names
- **Client type registry**: JSON Schema based. Each provider declares `ConfigSchema()`. Validation via `ValidateConfig()` with recursive `oneOf`/`required`/`properties`
- **Memory provider registry**: Same pattern as clients — `init()` + blank imports in `main.go`
- **Hot-reload**: Store `OnChange()` channel → `agentRouterHandler` rebuilds with 500ms debounce
- **MCP transports**: HTTP (`StreamableClientTransport` with optional headers + TLS skip) and stdio (`CommandTransport`). Stdio spawns subprocesses — works best with binary installs, not Docker
- **Migration chain** (on load): `migrateTTSConfig` moves legacy `tts.speed` → `tts.config.openai.speed` and flat Gemini config → `tts.config.gemini.*`. Operates on raw JSON before unmarshal. All idempotent
- **Webhook auth**: Separate from `clientAuthMiddleware`. Webhook handler validates Bearer token against client's `cl.Token`
- **Flow execution**: `FlowDefinition` recursive tree maps 1:1 to ADK workflow agents. `responseAgent` flag on `FlowStep` filters output
- **Voice endpoints**: `/api/v1/voice/{agentId}/speech` and `/transcription` resolve backends dynamically per agent
- **MCP headers/TLS**: `MCPServer` struct has `Headers map[string]string` and `Insecure bool`. `httpClientForMCP()` creates transport with optional `InsecureSkipVerify`
- **Skill injection**: Skills are injected into the agent system prompt at build time. Instructions appended as `--- Skill: {name} ---`, reference file contents appended as `[Reference: {filename}]`. Files read from `data/skills/{skillId}/`
- **Encryption key**: `server.encryptionKey` in config.yaml. Independent from `adminPassword`. Used to encrypt secrets at rest (AES-256-GCM, PBKDF2-derived)
- **ContextGuard plugin**: Externalized to `adk-utils-go/plugin/contextguard` (v0.7.0). Builder API: `contextguard.New(registry)` + `guard.Add(agentID, llm, opts...)` + `guard.PluginConfig()`. Two strategies: `threshold` (token-based, auto-detect via CrushRegistry or manual `WithMaxTokens`) and `sliding_window` (turn-count via `WithSlidingWindow`). Each agent summarizes with its own LLM. Summary persisted in session state with `{agentName}` suffix keys. `CrushRegistry` fetches model metadata from Crush's provider.json with 6h background refresh
- **A2A protocol**: Agents/flows with `A2A.Enabled` get JSON-RPC endpoints via `a2a-go` + ADK `adka2a`. Agent cards auto-generated with capabilities and skills. SSE streaming for responses
- **Dual-perspective conversation recording**: Middleware chains recorder twice: "admin" perspective (all events, before FlowResponseFilter) and "user" perspective (filtered, after). Each conversation has a `ParentID` linking the pair
- **Store dual-copy pattern**: Store maintains `rawData` (unexpanded, with `${VAR}` refs) and `data` (env-expanded). API responses use raw data, runtime uses expanded. Secret values injected as env vars before expansion
- **Session middleware**: `SessionEnsure` prevents overwriting existing sessions (protects ContextGuard summaries). `SessionStateSeed` injects empty outputKey values so `{{agent.output:key}}` references in flow agent prompts resolve correctly
- **SnakeCaseNormalize middleware**: Intercepts POST `/run` and `/run_sse`, recursively converts all snake_case JSON keys to camelCase before ADK processes the request. ADK uses `DisallowUnknownFields()` so any unconverted snake_case key causes a 400. Conversion is generic (not a fixed key list) — handles all nesting levels including `genai.Part` fields like `inlineData`, `mimeType`, `functionCall`. When both forms coexist in the same object, camelCase wins. Single-word keys are never modified. Placed outermost in the middleware chain (before ConversationRecorder)
- **InstructionProvider pattern**: Agents use `InstructionProvider` instead of static `Instruction` strings to bypass ADK's built-in `{variable}` substitution (which conflicts with curly braces in prompts, JSON examples, scripts, etc). Magec resolves its own `{{agent.output:key}}` pattern from session state inside the provider. Plain `{text}` in prompts is never touched
- **Flow wrapAgent pattern**: Same agent can appear in multiple flow steps — `wrapAgent()` creates uniquely-named delegate agents to satisfy ADK's single-parent constraint
- **Voice provider registry**: TTS/STT proxies dispatch to per-backend-type providers via `voice.Get(backend.Type)`. Same pattern as clients/memory — `init()` + blank imports. OpenAI provider handles `/v1/audio/speech` and `/v1/audio/transcriptions`. Gemini provider translates to `generateContent` with `speechConfig`/`inlineData`. `TTSRef.Config` and `BackendRef.Config` use typed structs (`TTSConfig`, `STTConfig`) with per-provider namespaces matching the `ClientConfig` pattern (e.g. `config.openai.speed`, `config.gemini.languageCode`). `GET /voice/types` returns JSON Schemas per provider. Store migration moves legacy `tts.speed` → `tts.config.openai.speed` and flat config fields → `tts.config.gemini.*`

### Design Philosophy

**DRY, KISS, elegant, and decoupled does not mean over-engineered.** When two options exist — one simple and one cleverly abstracted — prefer the simple one. Specifically:

- Don't avoid a straightforward API call (with built-in cache + HTTP fallback) just to use a lower-level cache-only lookup that may silently miss. `s.Channel()` is the right call in discordgo; `s.State.Channel()` is an unnecessary micro-optimisation that trades reliability for nothing.
- Don't redesign function signatures (variadic, extra parameters, new types) when the problem is already solved at the call site with one extra line.
- Complexity is only justified when it removes real duplication or real coupling — not hypothetical ones.

### Frontend Conventions (admin-ui)

- **Vue 3 Composition API**: `<script setup>` everywhere, no Options API
- **Pinia**: Single store (`data.js`) with `init()` + `refresh()`
- **No Vue Router**: Tab navigation via `activeTab` ref + `location.hash`
- **Dialog pattern**: `defineExpose({ open })`, parents call `ref.value?.open(data)`. Native `<dialog>` + `showModal()`
- **JSON Schema form renderer**: `ClientDialog.vue` renders forms dynamically from `ConfigSchema()`
- **Flow editor**: `FlowCanvas.vue` (pan/zoom/toolbar) + `FlowBlock.vue` (recursive, vuedraggable)
- **Tailwind v4**: `@tailwindcss/vite` plugin, `@theme` directive for custom colors
- **11 active tabs**: backends, memory, mcps, agents, flows, commands, skills, clients, secrets, conversations, settings
- **Keyboard shortcuts**: `n` (new entity), `r` (refresh), `Cmd+K` (search palette)
- **Settings view**: Global memory provider selection + backup/restore (tar.gz)

### Frontend Conventions (voice-ui)

- **Vue 3 + Vite + Tailwind v4 + Pinia**: 14 components, single Pinia store
- **Audio pipeline**: Plain JS classes (AudioCapture, AudioRecorder, OpenAITTS, VoiceEventsClient)
- **Spokesperson**: User picks among `responseAgent`s for TTS/STT. Persisted per flow in localStorage
- **i18n**: Spanish (default) and English
- **PWA**: Installable, service worker

## Build Commands

```bash
make build              # Build frontends + models + Go binary → bin/magec-server
make dev                # Build all + start server (CONFIG=config.yaml)
make build-admin        # Build admin UI only
make build-voice        # Build voice UI only
make dev-admin          # Admin UI Vite dev server (port 5173)
make dev-voice          # Voice UI Vite dev server (port 5174)
make swagger            # Regenerate Swagger docs (admin + user)
make download-model     # Download wake word + auxiliary models
make clean              # Remove build artifacts

make docker-build       # Single-arch Docker build
make docker-buildx      # Multi-arch (amd64 + arm64)
make docker-push        # Multi-arch + push to GHCR

make infra              # Start PostgreSQL + Redis
make ollama             # Start Ollama with qwen3:8b + nomic-embed-text
make infra-stop         # Stop PostgreSQL + Redis
make infra-clean        # Remove all containers + volumes
```

## Docker Compose

Single `docker-compose.yaml` in `docker/compose/`. Self-contained with all local services:

- **magec** — Main server (:8080) + Admin UI (:8081)
- **redis** — Session storage
- **postgres** — Long-term memory (pgvector)
- **ollama** + **ollama-setup** — LLM (qwen3:8b) + embeddings (nomic-embed-text)
- **parakeet** — Speech-to-text (URL: `http://parakeet:5092`, no `/v1`)
- **tts** — Text-to-speech via openai-edge-tts (URL: `http://tts:5050`, no `/v1`, `REQUIRE_API_KEY=False`)

GPU section commented out by default. Users who want cloud providers create different backends in Admin UI.

## Dependencies

**Go backend:**

- `google.golang.org/adk` — Agent Development Kit (v0.4.0)
- `google.golang.org/genai` — Google GenAI SDK (v1.40.0)
- `github.com/achetronic/adk-utils-go` — ADK utilities (v0.7.0): providers, session, memory tools, ContextGuard plugin
- `github.com/a2aproject/a2a-go` — A2A protocol library (v0.3.3)
- `github.com/modelcontextprotocol/go-sdk` — MCP client (v1.2.0)
- `github.com/gorilla/mux` — HTTP router (v1.8.1)
- `github.com/gorilla/websocket` — WebSocket for voice handler (v1.5.3)
- `github.com/mymmrac/telego` — Telegram bot API (v1.5.1)
- `github.com/bwmarrin/discordgo` — Discord bot API + Gateway WebSocket (v0.29.0)
- `github.com/slack-go/slack` — Slack API + Socket Mode (v0.17.3)
- `github.com/yalue/onnxruntime_go` — ONNX runtime for voice models (v1.25.0)
- `golang.org/x/crypto` — PBKDF2 for secret encryption (v0.46.0)
- `github.com/felixge/httpsnoop` — Middleware metrics
- `gopkg.in/yaml.v3` — YAML config parsing

**Frontends:**

- Vue 3, Vite 7.3, Tailwind CSS 4.1, Pinia 3
- vuedraggable (admin-ui flow editor)
- marked (admin-ui markdown rendering in conversations)

## Gotchas

1. **Both UIs use Vite**: `cd frontend/admin-ui && npx vite build` and `cd frontend/voice-ui && npx vite build`.
2. **Voice detection is server-side**: All ONNX inference (wake word + VAD) via WebSocket.
3. **Memory is optional**: Without Redis/PostgreSQL, sessions are in-memory and long-term memory is disabled.
4. **PWA over HTTP**: Requires Chrome flag for non-localhost addresses.
5. **Telegram/Discord voice**: Requires TTS backend configured. ffmpeg required in container.
6. **Parakeet/Edge TTS URLs**: Do NOT include `/v1` — Magec auto-appends it.
7. **Edge TTS auth**: Use `REQUIRE_API_KEY=False` env var, no API key needed in backend config.
8. **JSON Schema extensions**: `x-entity` (entity select), `x-format: password`, `x-placeholder`. Frontend renders dynamically.
9. **OutputKey on AgentDefinition, not FlowStep**: ADK's `OutputKey` is set on the agent.
10. **`responseAgent` is per-flow-step**: Lives on `FlowStep`. Executor resolves via `flow.ResponseAgentIDs()`. If none marked, all events returned.
11. **Cron supports shorthands**: `@daily`, `@hourly`, `@weekly`, etc. expand to 5-field expressions.
12. **Docker image includes default config.yaml**: Baked in at `/app/config.yaml`. Override with `-v`.
13. **Git branch is `master`**, not `main`. All raw GitHub URLs use `master`.
14. **Go 1.25+, Node 22+, Hugo v0.155+**.
15. **A2A agent card endpoints bypass client auth**: `.well-known/agent-card.json` paths are exempted from `ClientAuth` middleware so external agents can discover cards.
16. **ContextGuard `safeSplitIndex`**: When splitting conversation history for summarization, the split point is adjusted to avoid orphaning Anthropic `tool_result` blocks.
17. **Store env var expansion**: All store fields support `${VAR}` syntax. Secrets are injected as env vars (`os.Setenv`) before the store is expanded, so secrets can be referenced in backend URLs, bot tokens, etc.
18. **Voice API routes always registered**: STT/TTS proxy endpoints are available regardless of Voice UI toggle, since Telegram/Discord/Slack clients need them.
19. **ADK REST API accepts both camelCase and snake_case**: The `SnakeCaseNormalize` middleware converts snake_case keys recursively before ADK sees the request. This applies only to `/run` and `/run_sse` — the only ADK endpoints with multi-word JSON body fields. Session create/get/delete use single-word body fields (`state`, `events`) or path parameters only.

## Testing

```bash
make infra              # Start PostgreSQL + Redis
make dev                # Build and run
# Open http://localhost:8081 → configure backends, agents, clients
# Open http://localhost:8080 → voice/text chat
```

## Related Resources

- [Google ADK](https://google.github.io/adk-docs/)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [OpenWakeWord](https://github.com/dscripka/openWakeWord)
- [pgvector](https://github.com/pgvector/pgvector)
- [Parakeet](https://github.com/achetronic/parakeet)
- [openai-edge-tts](https://github.com/travisvn/openai-edge-tts)
- [hass-mcp](https://github.com/achetronic/hass-mcp)
