# AGENTS.md - Magec

Personal AI assistant from the Canary Islands 🇮🇨 that you control.

## Project Overview

**Magec** is a personal AI assistant that runs on your server. Named after the Guanche god of the Sun (/maˈxek/), it provides:

- **Multi-provider LLM**: OpenAI, Anthropic (Claude), Google Gemini, and local models via Ollama
- **Session memory**: Conversation history stored in Redis, survives restarts
- **Long-term memory**: Remembers things about you across sessions using PostgreSQL with pgvector
- **MCP tools**: Connect external tools via Model Context Protocol (Home Assistant, filesystem, GitHub, databases...)
- **Multiple clients**: Access via web (voice-ui), Telegram, webhooks, cron jobs, or future interfaces

### Clients

| Client | Type | Description |
|--------|------|-------------|
| **voice-ui** | `direct` | Web interface with voice/text chat, wake word detection, audio visualizer, PWA support |
| **Telegram** | `telegram` | Text and voice messages from your phone |
| **Webhook** | `webhook` | HTTP endpoint for external integrations (fixed command or passthrough prompt) |
| **Cron** | `cron` | Scheduled task that fires a command against agents on a schedule |
| *(Future)* | | Discord, Slack, WhatsApp, CLI... |

### Voice Capabilities

- **Wake word detection**: Server-side OpenWakeWord models ("Oye Magec", "Magec")
- **Speech-to-text**: OpenAI-compatible Whisper APIs (e.g., Parakeet)
- **Text-to-speech**: OpenAI-compatible TTS APIs (e.g., openai-edge-tts)

## Quick Start

```bash
# Docker Compose (recommended)
cd docker/compose/fully-local
docker-compose up -d

# OR from source
make infra        # Start PostgreSQL + Redis
make dev          # Build and run
```

Open http://localhost:8081 to configure agents and backends via the Admin UI.
Then open http://localhost:8080 to start chatting.

## Architecture

```
magec/
├── server/                     # Go backend (core)
│   ├── main.go                 # HTTP server, routing
│   ├── agent/
│   │   ├── agent.go            # ADK agent with memory, MCP tools
│   │   └── flow.go             # Flow→ADK workflow agent builder (sequential/parallel/loop)
│   ├── api/                    # REST API packages
│   │   ├── admin/              # Admin REST API (multi-agent management)
│   │   │   ├── handler.go      # Router + helpers
│   │   │   ├── agents.go       # Agent CRUD handlers
│   │   │   ├── backends.go     # Backend CRUD handlers
│   │   │   ├── clients.go      # Client CRUD handlers + /types (JSON Schema)
│   │   │   ├── commands.go     # Command CRUD handlers
│   │   │   ├── memory.go       # Memory provider CRUD + health check + /types
│   │   │   ├── flows.go        # Flow CRUD handlers + recursive validation
│   │   │   ├── conversations.go # Conversation audit handlers (list/get/delete/clear/stats/summary)
│   │   │   └── docs/           # Generated swagger (swagger.json/yaml)
│   │   └── user/               # User-facing REST API
│   │       ├── handlers.go     # Health, DeviceInfo, Voice stubs, Webhook swagger types
│   │       ├── doc.go          # Swagger metadata (title, version, host, security)
│   │       └── docs/           # Generated swagger (userapi_swagger.json/yaml)
│   ├── middleware/              # HTTP middleware (AccessLog, CORS, ClientAuth)
│   │   ├── middleware.go        # Uses httpsnoop — see DECISIONS.md
│   │   └── recorder.go          # ConversationRecorder middleware (captures /run + /run_sse)
│   ├── clients/                 # Client types: registry + specs + runtime
│   │   ├── provider.go          # Provider interface: Type(), DisplayName(), ConfigSchema()
│   │   ├── registry.go          # Global registry: Register(), ValidateConfig() with oneOf support
│   │   ├── executor.go          # RunClient() — executes commands against all allowedAgents
│   │   ├── direct/
│   │   │   └── spec.go          # Direct provider (empty config)
│   │   ├── webhook/
│   │   │   ├── spec.go          # Webhook provider schema (passthrough/commandId oneOf)
│   │   │   └── handler.go       # Webhook HTTP handler — Bearer token auth, passthrough/fixed modes
│   │   ├── cron/
│   │   │   ├── spec.go          # Cron provider schema (schedule, commandId)
│   │   │   ├── cron.go          # Cron expression parser
│   │   │   └── scheduler.go     # Cron scheduler — filters cron-type clients, fires on schedule
│   │   └── telegram/
│   │       ├── spec.go          # Telegram provider schema (botToken, allowedUsers, responseMode)
│   │       └── bot.go           # Telegram bot — voice, per-chat agents, response modes
│   ├── schema/                 # Shared JSON Schema validation (google/jsonschema-go)
│   │   └── validate.go         # Validate(schema, data) — marshal→unmarshal→resolve→validate
│   ├── store/                  # In-memory data store with JSON persistence
│   │   ├── conversations.go    # ConversationStore — separate JSON persistence (data/conversations.json)
│   ├── memory/                 # Extensible memory provider registry
│   │   ├── provider.go         # Provider interface, Category type, HealthResult
│   │   ├── registry.go         # Global registry: Register(), Get(), All(), ValidTypeForCategory()
│   │   ├── redis/redis.go      # Redis provider (session), Ping via ParseURL
│   │   └── postgres/postgres.go # Postgres provider (longterm), Ping via sql.Open
│   ├── config/config.go        # YAML config parsing (server + log only)
│   ├── logging/logging.go      # Structured logging (slog)
│   ├── voice/                  # Server-side voice detection (wake word + VAD)
│   │   ├── detector.go         # ONNX-based OpenWakeWord inference
│   │   ├── vad.go              # ONNX-based Silero VAD inference
│   │   ├── handler.go          # WebSocket handler for audio streaming
│   │   └── resampler.go        # Audio resampling to 16kHz
├── voice-ui/                   # Web client (voice interface)
│   ├── src/
│   │   ├── app.js              # Main application (MagecApp class)
│   │   ├── config.js           # API endpoints, session config
│   │   ├── audio/              # Audio capture, recording, TTS, VoiceEventsClient
│   │   ├── api/                # Agent API client
│   │   ├── transcription/      # Remote speech-to-text
│   │   ├── i18n/               # Internationalization (es.js, en.js)
│   │   ├── ui/                 # UIController, WaveformRenderer, templates
│   │   ├── session/            # Session management (local + server)
│   │   ├── settings/           # Settings persistence
│   │   ├── errors/             # Centralized error handling
│   │   └── utils/              # Utilities (WakeLock)
│   ├── assets/                 # PWA icons
│   ├── manifest.json           # PWA manifest
│   └── index.html
├── admin-ui/                   # Admin web interface (Vue 3 + Vite + Tailwind v4 + Pinia)
│   ├── src/
│   │   ├── main.js             # Vue app entry with Pinia
│   │   ├── App.vue             # Layout, sidebar navigation, global ConfirmDialog/Toast/SearchPalette
│   │   ├── style.css           # Tailwind v4 @theme (piedra/atlantico/lava/sol/arena)
│   │   ├── lib/
│   │   │   ├── api/            # Fetch wrapper + CRUD per resource (agents, backends, clients, commands, flows, conversations, etc.)
│   │   │   └── stores/data.js  # Pinia central store (all resources + helpers)
│   │   ├── components/         # Shared: AppDialog, Card, Badge, FormInput, Icon, Toast, Tooltip, SkeletonCard, SearchPalette, Sidebar, TopBar, EmptyState, etc.
│   │   └── views/              # Entity views (one folder each):
│   │       ├── backends/       # BackendsList + BackendDialog
│   │       ├── memory/         # MemoryList + MemoryCard + MemoryDialog
│   │       ├── mcps/           # McpsList + McpDialog
│   │       ├── agents/         # AgentsList + AgentDetail + AgentDialog
│   │       ├── clients/        # ClientsList + ClientDialog (JSON Schema renderer)
│   │       ├── commands/       # CommandsList + CommandDialog
│   │       ├── crons/          # CronsList + CronDialog (legacy, auto-migrated)
│   │       ├── flows/          # FlowsList + FlowDialog + FlowCanvas + FlowBlock
│   │       └── conversations/  # ConversationsView + ConversationsList + ConversationDetail (audit log)
│   ├── index.html
│   ├── vite.config.js          # Vue plugin + Tailwind plugin + dev proxy to :8081
│   └── package.json            # vue, pinia, vuedraggable, tailwindcss v4
├── models/                     # Wake word ONNX models + wakewords.yaml
├── pretrained/                 # Shared ONNX models (mel-spec, VAD, embeddings)
├── scripts/
│   └── download-model.go       # Wake word model downloader
├── docker/
│   ├── build/                  # Dockerfile + entrypoint
│   └── compose/                # Docker Compose deployments
├── website/                   # Static landing page + docs (GitHub Pages)
│   ├── index.html             # Landing page with hero, features, architecture
│   ├── docs.html              # Full documentation with sidebar navigation
│   ├── css/                   # Design tokens + styles (Canarian palette)
│   ├── js/                    # Centella orb + nav + animations
│   └── assets/                # Logo, screenshots, architecture SVG
├── config.example.yaml
├── Makefile
└── README.md
```

### Component Overview

| Component | Purpose |
|-----------|---------|
| `server/main.go` | HTTP server (port 8080) + admin server (port 8081), routing, middleware |
| `server/agent/agent.go` | Multi-agent ADK setup. `New()` accepts agents + flows, creates LLM agents + workflow agents, routes via `NewMultiLoader` |
| `server/agent/flow.go` | Translates `FlowDefinition` tree → ADK workflow agents (sequential/parallel/loop) |
| `server/api/admin/` | Admin REST API for managing all resources at runtime |
| `server/clients/` | Unified client package: provider registry (specs + schemas) + runtime execution (executor, webhook handler, cron scheduler, telegram bot) |
| `server/memory/` | Extensible provider registry — interface + init() auto-registration pattern |
| `server/store/` | In-memory data store with JSON persistence (`data/store.json`). Immutable UUID v4 IDs |
| `server/store/conversations.go` | Conversation audit store — separate file (`data/conversations.json`). Types: Conversation, ConversationMessage, ToolCallInfo |
| `server/middleware/recorder.go` | HTTP middleware that captures all `/run` and `/run_sse` requests for conversation auditing |
| `server/api/user/` | User-facing API handlers + Swagger docs (health, device info, voice, webhooks) |
| `server/voice/` | Server-side voice detection (wake word + VAD) via ONNX |
| `server/clients/telegram/` | Telegram bot with voice message support |
| `server/config/config.go` | YAML config for server infrastructure only (ports, log) |
| `admin-ui/` | Admin dashboard SPA (Vue 3 + Vite + Tailwind v4 + Pinia + vuedraggable). JSON Schema-driven client forms. Canvas-based flow editor. Served on admin port |
| `voice-ui/src/app.js` | Main entry - MagecApp class orchestrates audio pipeline |

## HTTP Endpoints

### Main Server (port 8080) — User API

| Method | Path | Description |
|--------|------|-------------|
| GET/POST | `/api/v1/agent/*` | ADK REST API (sessions, run, events). Uses `appName` to route to correct agent |
| POST | `/api/v1/agent/run` | Run agent (blocking) |
| POST | `/api/v1/agent/run_sse` | Run agent (SSE streaming) |
| POST | `/api/v1/webhooks/{clientId}` | **Webhook endpoint** — Bearer token auth, passthrough or fixed command |
| POST | `/api/v1/voice/{agentId}/speech` | TTS proxy (resolves backend dynamically per agent from store) |
| POST | `/api/v1/voice/{agentId}/transcription` | STT proxy (resolves backend dynamically per agent from store) |
| WebSocket | `/api/v1/voice/events` | Voice events stream (wake word + VAD) |
| GET | `/api/v1/client/info` | Client info (paired status, allowed agents) |
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/swagger/` | Swagger UI (userapi docs) |
| GET | `/` | Static files from `voice-ui/` |

### Admin Server (port 8081) — Admin API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Admin UI (static files from `admin-ui/`) |
| GET | `/api/v1/admin/swagger/` | Swagger UI (admin API docs) |
| GET | `/api/v1/admin/overview` | Overview: counts + agent summaries |
| | **Backends** | |
| GET | `/api/v1/admin/backends` | List all backends |
| POST | `/api/v1/admin/backends` | Create a backend |
| GET | `/api/v1/admin/backends/{id}` | Get a backend by ID |
| PUT | `/api/v1/admin/backends/{id}` | Update a backend |
| DELETE | `/api/v1/admin/backends/{id}` | Delete a backend |
| | **Memory Providers** | |
| GET | `/api/v1/admin/memory` | List all memory providers |
| POST | `/api/v1/admin/memory` | Create a memory provider |
| GET | `/api/v1/admin/memory/types` | List registered provider types + categories |
| GET | `/api/v1/admin/memory/{id}` | Get a memory provider by ID |
| PUT | `/api/v1/admin/memory/{id}` | Update a memory provider |
| DELETE | `/api/v1/admin/memory/{id}` | Delete a memory provider |
| GET | `/api/v1/admin/memory/{id}/health` | Real-time health check (Ping) |
| | **MCP Servers** | |
| GET | `/api/v1/admin/mcps` | List all MCP servers |
| POST | `/api/v1/admin/mcps` | Create an MCP server |
| GET | `/api/v1/admin/mcps/{id}` | Get an MCP server by ID |
| PUT | `/api/v1/admin/mcps/{id}` | Update an MCP server |
| DELETE | `/api/v1/admin/mcps/{id}` | Delete an MCP server |
| | **Agents** | |
| GET | `/api/v1/admin/agents` | List all agents |
| POST | `/api/v1/admin/agents` | Create an agent |
| GET | `/api/v1/admin/agents/{id}` | Get an agent by ID |
| PUT | `/api/v1/admin/agents/{id}` | Update an agent |
| DELETE | `/api/v1/admin/agents/{id}` | Delete an agent |
| GET | `/api/v1/admin/agents/{id}/mcps` | List resolved MCPs for an agent |
| PUT | `/api/v1/admin/agents/{id}/mcps/{mcpId}` | Link an MCP to an agent |
| DELETE | `/api/v1/admin/agents/{id}/mcps/{mcpId}` | Unlink an MCP from an agent |
| | **Clients** | |
| GET | `/api/v1/admin/clients` | List all clients |
| POST | `/api/v1/admin/clients` | Create a client |
| GET | `/api/v1/admin/clients/types` | List registered types with JSON Schema |
| GET | `/api/v1/admin/clients/{id}` | Get a client by ID |
| PUT | `/api/v1/admin/clients/{id}` | Update a client |
| DELETE | `/api/v1/admin/clients/{id}` | Delete a client |
| POST | `/api/v1/admin/clients/{id}/regenerate-token` | Regenerate client auth token |
| | **Commands** | |
| GET | `/api/v1/admin/commands` | List all commands |
| POST | `/api/v1/admin/commands` | Create a command |
| GET | `/api/v1/admin/commands/{id}` | Get a command by ID |
| PUT | `/api/v1/admin/commands/{id}` | Update a command |
| DELETE | `/api/v1/admin/commands/{id}` | Delete a command |
| | **Flows** | |
| GET | `/api/v1/admin/flows` | List all flows |
| POST | `/api/v1/admin/flows` | Create a flow |
| GET | `/api/v1/admin/flows/{id}` | Get a flow by ID |
| PUT | `/api/v1/admin/flows/{id}` | Update a flow |
| DELETE | `/api/v1/admin/flows/{id}` | Delete a flow |

See [MULTI_AGENT_ADMIN_API.md](MULTI_AGENT_ADMIN_API.md) for full API reference with request/response schemas.

## Configuration

Magec uses a **split configuration** model:

- **`config.yaml`** — Server infrastructure only (ports, logging). Read at startup.
- **Admin API + Store** — All resources managed via the Admin UI at `:8081` or the REST API. Persisted to `data/store.json`.

### config.yaml (Infrastructure)

```yaml
server:
  host: 0.0.0.0
  port: 8080
  adminPort: 8081  # Admin UI + API (default: 8081)
  onnxLibraryPath: /usr/lib/libonnxruntime.so  # Optional, default shown

log:
  level: info   # debug, info, warn, error
  format: console  # console, json
```

### Store Resources (Admin API)

All of the following are managed at runtime via `http://localhost:8081`:

| Resource | Description |
|----------|-------------|
| **Backends** | Reusable AI backends (OpenAI, Ollama, Anthropic, Gemini) |
| **Memory Providers** | Redis (session), PostgreSQL (long-term) |
| **MCP Servers** | External tool servers (HTTP or stdio) |
| **Agents** | Independent units with own LLM, memory, tools, prompts |
| **Commands** | Reusable prompts that can be invoked by cron/webhook clients |
| **Clients** | Access points with token-based auth. Types: `direct` (voice-UI), `telegram`, `cron` (scheduled), `webhook` (HTTP) |
| **Flows** | Multi-agent workflows (sequential/parallel/loop) |

On first run with no `data/store.json`, the store starts empty. Configure everything via the Admin UI.

### Client Types

| Type | Config Schema | Use Case |
|------|--------------|----------|
| `direct` | `{}` (empty) | Voice-UI tablets, apps — token-only auth |
| `telegram` | `botToken`, `allowedUsers`, `allowedChats`, `responseMode` | Telegram bot |
| `cron` | `schedule` (5-field cron or shorthand like `@daily`), `commandId` (ref to Commands) | Scheduled automation |
| `webhook` | `passthrough` (bool) + `commandId` (oneOf exclusive) | HTTP endpoint for integrations |

### Backend Types

| Type | Description | Required Fields |
|------|-------------|-----------------|
| `openai` | OpenAI-compatible API (OpenAI, Ollama, LM Studio, etc.) | `url` and/or `apiKey` |
| `anthropic` | Anthropic Claude API | `apiKey` |
| `gemini` | Google Gemini API | `apiKey` |

## Code Patterns

### Go Conventions

- **YAML config**: Server infrastructure only (ports, log), env var expansion with `${VAR}`
- **Store-based resources**: All entities managed via admin API
- **No YAML seed**: Store starts empty on first run; configure via Admin UI at `:8081`
- **Multi-agent ADK**: `agent.New()` accepts agents + flows, creates LLM agents + workflow agents, `NewMultiLoader` routes by `appName`
- **Immutable UUID v4 IDs**: All entities use `google/uuid` v4. Cross-references store IDs, not names
- **Client type registry**: JSON Schema based. Each provider declares `ConfigSchema()`. Validation via `ValidateConfig()` with recursive `oneOf`/`required`/`properties` walking
- **Automation**: `trigger/` package handles cron scheduling + webhook serving + command execution. `RunClient()` iterates all `allowedAgents`
- **Flows**: `FlowDefinition` with recursive `FlowStep` tree. Maps 1:1 to ADK workflow agents (sequential/parallel/loop)
- **Hot-reload**: Store `OnChange()` channel fires on `persist()`. `agentRouterHandler` rebuilds agent on store changes with 500ms debounce
- **ADK REST handler**: `adkrest.NewHandler()` provides full ADK API
- **Memory tools**: Agent has `search_memory` and `save_to_memory` tools
- **Agent instruction**: Reads memories at conversation start
- **Voice endpoints**: `/api/v1/voice/{agentId}/speech` and `/transcription` resolve backends dynamically per agent. API keys forwarded to upstream
- **WebSocket voice-events**: Server handles all ONNX inference (wake word + VAD), clients stream audio at `/api/v1/voice/events`
- **Client auth middleware**: Token-based auth on port 8080. Whitelist: health, voice/events, client/info, webhooks, OPTIONS, static files
- **Webhook auth**: Separate from middleware — webhook.go validates Bearer token against the client's own `cl.Token`
- **Migration chain**: On load: `devices→clients` → `cronJobs→triggers` → `triggers→clients` → `device→direct` → `migrateIDs`. Each step idempotent

### JavaScript Conventions (admin-ui)

- **Vue 3 Composition API**: `<script setup>` in all components, no Options API
- **Pinia**: Single store (`data.js`) with `init()` + `refresh()` pattern
- **No Vue Router**: Tab navigation via `activeTab` ref + `location.hash`
- **Dialog pattern**: `defineExpose({ open })`, parents call `ref.value?.open(data)`. Native `<dialog>` with `showModal()`
- **Delete confirmation**: Global via `provide('requestDelete')` / `inject('requestDelete')`
- **Toast notifications**: Global via `provide('toast')` / `inject('toast')` — `toast.success()`, `toast.error()`, `toast.info()`
- **Keyboard shortcuts**: Global handler in `App.vue` — `n` (new), `r` (refresh), `Cmd+K` (search). Skips inputs/textareas/dialogs
- **Search palette**: `SearchPalette.vue` — `Cmd+K` triggers, searches all 8 entity types by name/description
- **JSON Schema form renderer**: `ClientDialog.vue` renders forms dynamically from `ConfigSchema()`. Supports `oneOf` branch matching, `x-entity` (entity select), `enum` (select), `boolean` (toggle), `x-format:password`
- **Entity views**: `*List.vue` + `*Dialog.vue` per entity under `src/views/<entity>/`
- **Flow editor**: `FlowCanvas.vue` (pan/zoom/toolbar) + `FlowBlock.vue` (recursive, vuedraggable)
- **Tailwind v4**: `@tailwindcss/vite` plugin, `@theme` directive for custom colors
- **Build**: `npx vite build` → `admin-ui/dist/`, Go serves from there
- **8 active tabs**: backends, memory, mcps, agents, flows, commands, clients, conversations

### JavaScript Conventions (voice-ui)

- **ES Modules**: Uses `import` from CDN (no build step)
- **Class-based**: `MagecApp` orchestrates all components
- **i18n**: `t('key', {params})` function, `data-i18n` attributes in HTML
- **Storage keys**: `magec_settings`, `magec_language`, `magec_sessions`
- **Color palette**: piedra (grays), atlantico (cyan), lava (red), sol (yellow/orange), arena (text)
- **Centralized errors**: All API errors flow through ErrorHandler → notifications
- **Multi-agent**: `setAgent(agentId)` propagated to `AgentClient`, `SessionService`, `OpenAITTS`, `RemoteTranscriber`

### Audio Processing Pipeline (voice-ui)

1. **Microphone capture** → AudioCapture with AudioWorklet at 16kHz
2. **Voice events** → Audio streamed via WebSocket to server → wake word + VAD detection
3. **Recording** → AudioRecorder captures audio as webm (started by wake word, stopped by VAD)
4. **Conversion** → AudioConverter resamples to 16kHz WAV
5. **Transcription** → RemoteTranscriber POSTs to Whisper API
6. **Agent interaction** → AgentClient POSTs to ADK `/run` endpoint
7. **TTS** → OpenAITTS plays response via `/api/v1/voice/{agentId}/speech`

### Voice Events (Server-side)

```
Client (WebSocket)          Server (/api/v1/voice/events)
       │                              │
       │<── capabilities ─────────────│ (on connect)
       │    {wakewords, vad}          │
       │                              │
       │─── config {sampleRate} ──────>│
       │─── audio (float32 LE) ───────>│ Resample to 16kHz
       │                              │
       │                              │ [Wake Word Detection]
       │                              │ Mel-spectrogram (ONNX)
       │                              │ Speech embedding (ONNX)
       │                              │ Wake word model (ONNX)
       │<── wakeword {model} ─────────│
       │                              │
       │                              │ [VAD Detection]
       │                              │ Silero VAD (ONNX)
       │<── speech_start ─────────────│
       │<── speech_end ───────────────│
```

## Agent Behavior

The agent (defined in `server/agent/agent.go`) has these key behaviors:

1. **Memory reading at start**: Calls `search_memory` at the beginning of every conversation
2. **Proactive memory saving**: Saves user preferences and important info to memory
3. **Language matching**: Responds in the same language as user input
4. **Concise responses**: Optimized for voice interaction
5. **MCP tools**: External tools from configured MCP servers are available

## Clients

### voice-ui (Web Interface)

PWA-enabled web interface with:

- **Centella/Magec visualizer**: Animated orb with particles, reacts to audio and recording state
- **Settings panel**: Wake word toggle, model selector, TTS toggle, language selector
- **Session management**: Local storage + server-side session listing
- **Notification system**: Centralized error handling with badges
- **i18n**: Spanish (default) and English

### Telegram Client

- **Text messages**: Processed through the agent
- **Voice messages**: Downloaded → converted OGG→WAV (ffmpeg) → transcribed → processed → optional voice response
- **Authorization**: `allowedUsers` and `allowedChats` allowlists (configured in client config)
- **Sessions**: Scoped by `telegram_<chatID>`
- **User context**: Every message sent to the LLM includes a `<!--MAGEC_META:{...}:MAGEC_META-->` prefix with Telegram metadata as JSON
- **Response mode**: Configurable via `responseMode` in client config. Can be changed at runtime with `/responsemode` command

### Webhook Client

- **Endpoint**: `POST /api/v1/webhooks/{clientId}` on port 8080
- **Auth**: `Authorization: Bearer <mgc_token>` (client's own token)
- **Passthrough mode** (`passthrough: true`): prompt comes from request body `{"prompt": "..."}`
- **Fixed command mode** (`passthrough: false`): prompt comes from referenced Command, body ignored
- **Execution**: Runs against ALL agents in client's `allowedAgents` list (agents and flows)
- **Bypass**: Not subject to `clientAuthMiddleware` — has its own auth in webhook.go

### Flow Execution & `responseAgent`

When a client targets a flow (via `allowedAgents`), the executor:
1. Detects the ID is a flow (`store.GetFlow(agentID)`)
2. Walks the flow tree for steps marked `responseAgent: true`
3. Passes those agent IDs as `responseFilter` to `extractResponseText`
4. Only ADK events where `event.author` matches a filtered agent are included

**`responseAgent` flag** lives on `FlowStep`, not `AgentDefinition`. The same agent can be a responseAgent in one flow but not another.

**Backwards compatible**: If no step has `responseAgent: true`, all events with text are returned (concatenated with `\n---\n`).

**Multiple responseAgents**: Multiple steps can be marked — their outputs are concatenated.

**Recommended pattern** (fan-out/synthesize):
```
Sequential:
  1. Parallel(Agent_A[outputKey=a_result], Agent_B[outputKey=b_result])
  2. Synthesizer[responseAgent=true, prompt reads {a_result} and {b_result}]
```

### Cron Client

- **Schedule**: 5-field cron expressions (`0 0 * * *`) or shorthands (`@daily`, `@hourly`, `@weekly`, `@monthly`, `@yearly`, `@annually`, `@midnight`)
- **Command**: References a Command entity by ID
- **Execution**: Fires `RunClient()` on schedule against all `allowedAgents`
- **Scheduler**: `trigger/scheduler.go` filters cron-type clients from store, manages next-fire times

## Development

### Build Commands

```bash
make build              # Build to bin/magec-server
make dev                # Build and run with config.yaml
make swagger            # Regenerate Swagger docs from annotations
make clean              # Remove build artifacts
make download-model     # Download wake word + pretrained models (interactive)

# Regenerate userapi swagger
cd server && go run github.com/swaggo/swag/cmd/swag init --dir ./userapi --generalInfo doc.go --output ./userapi/docs --parseDependency --parseInternal --instanceName userapi

# Regenerate admin swagger
cd server && go run github.com/swaggo/swag/cmd/swag init --dir ./admin --generalInfo doc.go --output ./admin/docs --parseDependency --parseInternal
```

### Infrastructure (Docker)

```bash
make postgres           # PostgreSQL with pgvector (port 5432)
make redis              # Redis (port 6379)
make ollama             # Ollama with qwen3:8b + nomic-embed-text
make infra              # postgres + redis
make infra-stop         # Stop postgres + redis
make infra-clean        # Remove all containers and volumes
```

## Docker Compose Deployments

Two ready-to-use deployments in `deploy/docker/`:

### Local (`deploy/docker/fully-local/`)

Fully self-hosted. No API keys needed. All AI runs locally.

**Services:**
- **magec** - Main server (port 8080) + Admin UI (port 8081)
- **redis** - Session storage
- **postgres** - Long-term memory (pgvector)
- **ollama** - LLM (qwen3:8b) + embeddings (nomic-embed-text)
- **ollama-setup** - Init container that pulls models on first start
- **parakeet** - Speech-to-text
- **tts** - Text-to-speech (openai-edge-tts)

### OpenAI (`deploy/docker/remote-openai/`)

Only infra runs locally. LLM, STT, TTS, and embeddings use OpenAI APIs.

## Dependencies

**Go backend:**
- `google.golang.org/adk` - Agent Development Kit
- `github.com/achetronic/adk-utils-go` - ADK utilities (providers, session, memory, tools)
- `github.com/modelcontextprotocol/go-sdk` - MCP client
- `github.com/yalue/onnxruntime_go` - ONNX runtime for wake word
- `gopkg.in/yaml.v3` - YAML config parsing

**Frontend (CDN):**
- Tailwind CSS - Styling

## Gotchas

1. **Voice-UI has no build step**: Dependencies loaded from CDN. No npm required. **Admin-UI uses Vite** — `make build-admin` or `cd admin-ui && npx vite build`.

2. **Voice detection is server-side**: All ONNX inference (wake word + VAD) happens on the server via WebSocket.

3. **Wake word models location**: ONNX models in `models/` at project root.

4. **VAD stops recording**: When VAD detects speech end, recording automatically stops.

5. **Memory is optional**: Without Redis/PostgreSQL, sessions are in-memory and long-term memory is disabled.

6. **MCP transports**: Supports both HTTP (StreamableClientTransport) and stdio (CommandTransport).

7. **PWA over HTTP**: Requires Chrome flag for non-localhost addresses.

8. **Telegram voice**: Requires TTS backend configured for voice responses. ffmpeg required in container.

9. **Security**: Always set `allowedUsers` in Telegram client config to restrict access.

10. **Memory providers use connectionString**: Both Redis and Postgres use a universal `connectionString` field. Provider-specific extra fields live in the `config` map.

11. **Memory provider registry**: Providers register via `init()` + blank imports in `main.go`. The `Provider` interface requires `Type()`, `DisplayName()`, `SupportedCategories()`, `ConfigSchema()`, and `Ping(ctx, config)`. Config validation via `ValidateConfig()` walks JSON Schema (same pattern as client registry).

12. **Client type registry**: Same pattern as memory — `init()` + blank imports. Provider interface: `Type()`, `DisplayName()`, `ConfigSchema()`. Config validation via `ValidateConfig()` walks JSON Schema recursively (supports `oneOf`, `required`, `properties`).

13. **JSON Schema extensions for client configs**: `x-entity: "commands"` (entity select), `x-format: "password"` (password input), `x-placeholder: "..."` (placeholder text). Frontend renders forms dynamically from these.

14. **Webhook auth is separate from middleware**: `clientAuthMiddleware` is bypassed for `/api/v1/webhooks/` paths. Webhook handler validates Bearer token against the client's own `cl.Token` directly.

15. **Cron supports both 5-field and shorthand expressions**: `@daily`, `@hourly`, `@weekly`, `@monthly`, `@yearly`, `@annually`, `@midnight` are all valid. They expand to their 5-field equivalents before parsing.

16. **Immutable IDs**: All entities use UUID v4 (`google/uuid`). Cross-references store IDs not names.

17. **Hot-reload**: Store changes fire `OnChange()` channel → `agentRouterHandler` rebuilds with 500ms debounce. No server restart needed.

18. **Multi-agent routing**: ADK uses `appName` in requests to route to the correct agent via `NewMultiLoader`.

19. **Flow editor is nested HTML, not a graph library**: The flow data model is a strict recursive tree. `FlowCanvas.vue` handles pan/zoom, `FlowBlock.vue` is recursive with `vuedraggable`.

20. **OutputKey on AgentDefinition, not FlowStep**: ADK's `OutputKey` is set on the agent, not per flow step.

21. **Webhook modes are exclusive**: Either `commandId` is set (fixed mode) OR `passthrough` is true (prompt from body). Enforced via `oneOf` in JSON Schema and server-side validation.

22. **Cron/webhook execute against ALL allowedAgents**: Not a single agentId. The same command runs against every agent in the client's `allowedAgents` list. Results joined with `\n---\n`.

23. **`responseAgent` is per-flow-step, not per-agent**: The flag lives on `FlowStep` in the flow definition. The executor resolves it at runtime via `flow.ResponseAgentIDs()`. If none marked, all events returned (backwards-compat). The flag is toggled in the flow editor UI via a broadcast icon on agent nodes.

## Testing

Manual testing workflow:

1. Start infrastructure: `make infra`
2. Start server: `make dev`
3. Open http://localhost:8081 to configure backends and agents via Admin UI
4. Open http://localhost:8080
5. Allow microphone access
6. Say wake word or tap the orb to start recording
7. Speak - recording stops automatically when you stop talking (VAD)
8. Verify agent responds

## Related Resources

- [Google ADK](https://google.github.io/adk-docs/)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [OpenWakeWord](https://github.com/dscripka/openWakeWord)
- [pgvector](https://github.com/pgvector/pgvector)
- [Parakeet](https://github.com/achetronic/parakeet)
- [openai-edge-tts](https://github.com/travisvn/openai-edge-tts)
- [hass-mcp](https://github.com/achetronic/hass-mcp)
