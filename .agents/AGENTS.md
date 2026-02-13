# AGENTS.md - Magec

Personal AI assistant from the Canary Islands 🇮🇨 that you control.

## Project Overview

**Magec** is a personal AI assistant that runs on your server. Named after the Guanche god of the Sun (/maˈxek/), it provides:

- **Multi-provider LLM**: OpenAI, Anthropic (Claude), Google Gemini, and local models via Ollama
- **Session memory**: Conversation history stored in Redis, survives restarts
- **Long-term memory**: Remembers things about you across sessions using PostgreSQL with pgvector
- **MCP tools**: Connect external tools via Model Context Protocol (Home Assistant, filesystem, GitHub, databases...)
- **Multiple clients**: Access via web (voice-ui), Telegram, or future interfaces

### Clients

| Client | Description |
|--------|-------------|
| **voice-ui** | Web interface with voice/text chat, wake word detection, audio visualizer, PWA support |
| **Telegram** | Text and voice messages from your phone |
| *(Future)* | Discord, Slack, WhatsApp, CLI... |

### Voice Capabilities

- **Wake word detection**: Server-side OpenWakeWord models ("Oye Magec", "Magec")
- **Speech-to-text**: OpenAI-compatible Whisper APIs (e.g., Parakeet)
- **Text-to-speech**: OpenAI-compatible TTS APIs (e.g., openai-edge-tts)

## Quick Start

```bash
# Docker Compose (recommended)
cd deploy/docker/fully-local
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
│   ├── main.go                 # HTTP server, routing, middleware
│   ├── agent/
│   │   ├── agent.go            # ADK agent with memory, MCP tools
│   │   └── flow.go             # Flow→ADK workflow agent builder (sequential/parallel/loop)
│   ├── admin/                  # Admin REST API (multi-agent management)
│   │   ├── handler.go          # Router + helpers
│   │   ├── agents.go           # Agent CRUD handlers
│   │   ├── backends.go         # Backend CRUD handlers
│   │   ├── commands.go         # Command CRUD handlers
│   │   ├── triggers.go         # Trigger CRUD handlers (cron + webhook)
│   │   ├── memory.go           # Memory provider CRUD + health check + /types
│   │   ├── flows.go            # Flow CRUD handlers + recursive validation
│   │   └── overview.go         # Overview/health handler
│   ├── store/                  # In-memory data store with JSON persistence
│   │   ├── store.go            # Store struct, CRUD operations, persistence, UUID migration, CronJob→Trigger migration
│   │   └── types.go            # AgentDefinition, BackendDefinition, MemoryProvider, Command, Trigger, FlowDefinition, etc.
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
│   └── clients/
│       └── telegram/           # Telegram bot client
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
│   │   │   ├── api/            # Fetch wrapper + CRUD per resource (agents, backends, flows, etc.)
│   │   │   └── stores/data.js  # Pinia central store (all resources + helpers)
│   │   ├── components/         # Shared: AppDialog, Card, Badge, FormInput, Icon, Toast, Tooltip, SkeletonCard, SearchPalette, Sidebar, TopBar, EmptyState, etc.
│   │   └── views/              # Entity views (one folder each):
│   │       ├── backends/       # BackendsList + BackendDialog
│   │       ├── memory/         # MemoryList + MemoryCard + MemoryDialog
│   │       ├── mcps/           # McpsList + McpDialog
│   │       ├── agents/         # AgentsList + AgentDetail + AgentDialog
│   │       ├── clients/        # ClientsList + ClientDialog
│   │       ├── crons/          # CronsList + CronDialog
│   │       └── flows/          # FlowsList + FlowDialog + FlowCanvas + FlowBlock
│   ├── index.html
│   ├── vite.config.js          # Vue plugin + Tailwind plugin + dev proxy to :8081
│   └── package.json            # vue, pinia, vuedraggable, tailwindcss v4
├── models/                     # Wake word ONNX models + wakewords.yaml
├── pretrained/                 # Shared ONNX models (mel-spec, VAD, embeddings)
├── scripts/
│   └── download-model.go       # Wake word model downloader
├── deploy/
│   └── docker/                 # Docker Compose deployments
│   └── docker-compose/         # Production deployment
├── config.example.yaml
├── Dockerfile
├── Makefile
└── README.md
```

### Component Overview

| Component | Purpose |
|-----------|---------|
| `server/main.go` | HTTP server (port 8080) + admin server (port 8081), routing, middleware |
| `server/agent/agent.go` | Multi-agent ADK setup. `New()` accepts agents + flows, creates LLM agents + workflow agents, routes via `NewMultiLoader` |
| `server/agent/flow.go` | Translates `FlowDefinition` tree → ADK workflow agents (sequential/parallel/loop) |
| `server/admin/` | Admin REST API for managing agents, backends, MCPs, and memory providers at runtime |
| `server/memory/` | Extensible provider registry — interface + init() auto-registration pattern |
| `server/store/` | In-memory data store with JSON persistence (`data/store.json`). Immutable UUID v4 IDs. Auto-migration of legacy data |
| `server/voice/` | Server-side voice detection (wake word + VAD) via ONNX |
| `server/clients/telegram/` | Telegram bot with voice message support |
| `server/config/config.go` | YAML config for server infrastructure only (ports, log) |
| `admin-ui/` | Admin dashboard SPA (Vue 3 + Vite + Tailwind v4 + Pinia + vuedraggable). Canvas-based flow editor with drag-and-drop. Served on admin port |
| `voice-ui/src/app.js` | Main entry - MagecApp class orchestrates audio pipeline |
| `voice-ui/src/audio/` | AudioCapture, AudioRecorder, AudioConverter, FeedbackSound, OpenAITTS, VoiceEventsClient |
| `voice-ui/src/api/AgentClient.js` | Agent API client (sessions, messages) |
| `voice-ui/src/ui/UIController.js` | DOM manipulation, panels, notifications, sidebar |
| `voice-ui/src/ui/WaveformRenderer.js` | "Centella/Magec" audio visualizer with particles |

## HTTP Endpoints

### Main Server (port 8080)

| Method | Path | Description |
|--------|------|-------------|
| GET/POST | `/api/v1/agent/*` | ADK REST API (sessions, run, events). Uses `appName` to route to correct agent |
| POST | `/api/v1/agent/run` | Run agent (blocking) |
| POST | `/api/v1/agent/run_sse` | Run agent (SSE streaming) |
| POST | `/api/v1/voice/{agentId}/speech` | TTS proxy (resolves backend dynamically per agent from store) |
| POST | `/api/v1/voice/{agentId}/transcription` | STT proxy (resolves backend dynamically per agent from store) |
| WebSocket | `/api/v1/voice/events` | Voice events stream (wake word + VAD) |
| GET | `/api/v1/device/info` | Device info (paired status, allowed agents) |
| GET | `/api/v1/health` | Health check |
| GET | `/` | Static files from `voice-ui/` |

### Admin Server (port 8081)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Admin UI (static files from `admin-ui/`) |
| GET | `/api/v1/admin/swagger/` | Swagger UI (interactive API docs) |
| GET | `/api/v1/admin/overview` | Overview: agent/backend/MCP counts + agent summaries |
| GET | `/api/v1/admin/backends` | List all backends |
| POST | `/api/v1/admin/backends` | Create a backend |
| GET | `/api/v1/admin/backends/{id}` | Get a backend by ID |
| PUT | `/api/v1/admin/backends/{id}` | Update a backend |
| DELETE | `/api/v1/admin/backends/{id}` | Delete a backend |
| GET | `/api/v1/admin/memory` | List all memory providers |
| POST | `/api/v1/admin/memory` | Create a memory provider |
| GET | `/api/v1/admin/memory/types` | List registered provider types + supported categories |
| GET | `/api/v1/admin/memory/{id}` | Get a memory provider by ID |
| PUT | `/api/v1/admin/memory/{id}` | Update a memory provider |
| DELETE | `/api/v1/admin/memory/{id}` | Delete a memory provider |
| GET | `/api/v1/admin/memory/{id}/health` | Real-time health check (Ping) for a provider |
| GET | `/api/v1/admin/mcps` | List all MCP servers (global) |
| POST | `/api/v1/admin/mcps` | Create an MCP server |
| GET | `/api/v1/admin/mcps/{id}` | Get an MCP server by ID |
| PUT | `/api/v1/admin/mcps/{id}` | Update an MCP server |
| DELETE | `/api/v1/admin/mcps/{id}` | Delete an MCP server |
| GET | `/api/v1/admin/agents` | List all agents |
| POST | `/api/v1/admin/agents` | Create an agent |
| GET | `/api/v1/admin/agents/{id}` | Get an agent by ID |
| PUT | `/api/v1/admin/agents/{id}` | Update an agent |
| DELETE | `/api/v1/admin/agents/{id}` | Delete an agent |
| GET | `/api/v1/admin/agents/{id}/mcps` | List resolved MCPs for an agent |
| PUT | `/api/v1/admin/agents/{id}/mcps/{mcpId}` | Link an MCP to an agent |
| DELETE | `/api/v1/admin/agents/{id}/mcps/{mcpId}` | Unlink an MCP from an agent |
| GET | `/api/v1/admin/clients` | List all clients |
| POST | `/api/v1/admin/clients` | Create a client |
| GET | `/api/v1/admin/clients/types` | List registered client types |
| GET | `/api/v1/admin/clients/{id}` | Get a client by ID |
| PUT | `/api/v1/admin/clients/{id}` | Update a client |
| DELETE | `/api/v1/admin/clients/{id}` | Delete a client |
| POST | `/api/v1/admin/clients/{id}/regenerate-token` | Regenerate client auth token |
| GET | `/api/v1/admin/crons` | List all cron jobs (legacy) |
| POST | `/api/v1/admin/crons` | Create a cron job (legacy) |
| GET | `/api/v1/admin/crons/{id}` | Get a cron job by ID (legacy) |
| PUT | `/api/v1/admin/crons/{id}` | Update a cron job (legacy) |
| DELETE | `/api/v1/admin/crons/{id}` | Delete a cron job (legacy) |
| GET | `/api/v1/admin/commands` | List all commands |
| POST | `/api/v1/admin/commands` | Create a command |
| GET | `/api/v1/admin/commands/{id}` | Get a command by ID |
| PUT | `/api/v1/admin/commands/{id}` | Update a command |
| DELETE | `/api/v1/admin/commands/{id}` | Delete a command |
| GET | `/api/v1/admin/triggers` | List all triggers |
| POST | `/api/v1/admin/triggers` | Create a trigger |
| GET | `/api/v1/admin/triggers/{id}` | Get a trigger by ID |
| PUT | `/api/v1/admin/triggers/{id}` | Update a trigger |
| DELETE | `/api/v1/admin/triggers/{id}` | Delete a trigger |
| GET | `/api/v1/admin/flows` | List all flows |
| POST | `/api/v1/admin/flows` | Create a flow |
| GET | `/api/v1/admin/flows/{id}` | Get a flow by ID |
| PUT | `/api/v1/admin/flows/{id}` | Update a flow |
| DELETE | `/api/v1/admin/flows/{id}` | Delete a flow |

See [MULTI_AGENT_ADMIN_API.md](MULTI_AGENT_ADMIN_API.md) for full API reference with request/response schemas.

## Configuration

Magec uses a **split configuration** model:

- **`config.yaml`** — Server infrastructure only (ports, logging). Read at startup.
- **Admin API + Store** — All resources (agents, backends, MCPs, memory providers, devices, crons). Managed via the Admin UI at `:8081` or the REST API. Persisted to `data/store.json`.

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
| **Commands** | Reusable prompts that can be invoked against agents by triggers |
| **Triggers** | Automation: cron schedules or webhooks that fire commands. Authenticate via linked Client token |
| **Clients** | Access points (voice-UI, Telegram, Discord) with token-based auth |
| **Cron Jobs** | Legacy — migrated automatically to Commands + Triggers on load |

On first run with no `data/store.json`, the store starts empty. Configure everything via the Admin UI.

### Backend Types

| Type | Description | Required Fields |
|------|-------------|-----------------|
| `openai` | OpenAI-compatible API (OpenAI, Ollama, LM Studio, etc.) | `url` and/or `apiKey` |
| `anthropic` | Anthropic Claude API | `apiKey` |
| `gemini` | Google Gemini API | `apiKey` |

## Code Patterns

### Go Conventions

- **YAML config**: Server infrastructure only (ports, log), env var expansion with `${VAR}`
- **Store-based resources**: All agents, backends, MCPs, memory, devices, crons managed via admin API
- **No YAML seed**: Store starts empty on first run; configure via Admin UI at `:8081`
- **Multi-agent ADK**: `agent.New()` accepts agents + flows, creates LLM agents + workflow agents, `NewMultiLoader` routes by `appName`
- **Immutable UUID v4 IDs**: All entities use `google/uuid` v4. Cross-references store IDs, not names. Auto-migration from legacy name-based refs on first load
- **Flows**: `FlowDefinition` with recursive `FlowStep` tree. Maps 1:1 to ADK workflow agents (sequential/parallel/loop)
- **Hot-reload**: Store `OnChange()` channel fires on `persist()`. `agentRouterHandler` rebuilds agent on store changes with 500ms debounce
- **ADK REST handler**: `adkrest.NewHandler()` provides full ADK API
- **Memory tools**: Agent has `search_memory` and `save_to_memory` tools
- **Agent instruction**: Reads memories at conversation start
- **Voice endpoints**: `/api/v1/voice/{agentId}/speech` and `/transcription` resolve backends dynamically per agent. API keys forwarded to upstream
- **WebSocket voice-events**: Server handles all ONNX inference (wake word + VAD), clients stream audio at `/api/v1/voice/events`
- **Device auth middleware**: Token-based auth on port 8080. Whitelist: health, voice/events, device/info, OPTIONS, static files
- **No rename cascade needed**: IDs are immutable UUIDs. Renaming = PUT with new `name` in body. Cross-references use stable IDs

### JavaScript Conventions (admin-ui)

- **Vue 3 Composition API**: `<script setup>` in all components, no Options API
- **Pinia**: Single store (`data.js`) with `init()` + `refresh()` pattern
- **No Vue Router**: Tab navigation via `activeTab` ref + `location.hash`
- **Dialog pattern**: `defineExpose({ open })`, parents call `ref.value?.open(data)`. Native `<dialog>` with `showModal()`
- **Delete confirmation**: Global via `provide('requestDelete')` / `inject('requestDelete')`
- **Toast notifications**: Global via `provide('toast')` / `inject('toast')` — `toast.success()`, `toast.error()`, `toast.info()`
- **Keyboard shortcuts**: Global handler in `App.vue` — `n` (new), `r` (refresh), `Cmd+K` (search). Skips inputs/textareas/dialogs. List views register via `inject('registerNew')`
- **Search palette**: `SearchPalette.vue` — `Cmd+K` triggers, searches all 8 entity types by name/description, keyboard navigation
- **Loading skeletons**: `SkeletonCard.vue` shown when `store.loading && !collection.length`
- **Section transitions**: Vue `<Transition>` with `mode="out-in"` keyed by `activeTab`
- **Responsive sidebar**: Mobile drawer (`mobileOpen` prop on Sidebar), hamburger in TopBar, backdrop overlay
- **Tooltips**: `Tooltip.vue` on cross-reference badges (agent/command badges show prompt snippet on hover)
- **Entity views**: `*List.vue` + `*Dialog.vue` per entity under `src/views/<entity>/`
- **Flow editor**: `FlowCanvas.vue` (pan/zoom/toolbar) + `FlowBlock.vue` (recursive, vuedraggable)
- **Tailwind v4**: `@tailwindcss/vite` plugin, `@theme` directive for custom colors
- **Build**: `npx vite build` → `admin-ui/dist/`, Go serves from there

### JavaScript Conventions (voice-ui)

- **ES Modules**: Uses `import` from CDN (no build step)
- **Class-based**: `MagecApp` orchestrates all components
- **i18n**: `t('key', {params})` function, `data-i18n` attributes in HTML
- **Storage keys**: `magec_settings`, `magec_language`, `magec_sessions`
- **Color palette**: piedra (grays), atlantico (cyan), lava (red), sol (yellow/orange), arena (text)
- **Centralized errors**: All API errors flow through ErrorHandler → notifications
- **Multi-agent**: `setAgent(agentId)` propagated to `AgentClient`, `SessionService`, `OpenAITTS`, `RemoteTranscriber`
- **Agent switching**: Dropdown triggers new session creation + endpoint reconfiguration

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

**Capabilities message (sent on connect):**
```json
{
  "type": "capabilities",
  "data": {
    "wakewords": {
      "models": [{"id": "oye-magec", "name": "Oye Magec", "phrase": "Oye Magec"}],
      "active": "oye-magec"
    },
    "vad": {
      "enabled": true,
      "silenceTimeout": 2000
    }
  }
}
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
- **Authorization**: `allowedUsers` and `allowedChats` allowlists
- **Sessions**: Scoped by `telegram_<chatID>`
- **User context**: Every message sent to the LLM includes a `<!--MAGEC_META:{...}:MAGEC_META-->` prefix with Telegram metadata as JSON (source, user ID, username, display name, chat ID, chat title, chat type) so the agent always knows who is talking and from where. This format is shared across all clients (Telegram, voice-ui with future OpenID JWT claims) and is stripped from user-facing views by the `stripMetadata()` utility
- **Response mode**: Configurable via `responseMode` in the agent's Telegram settings (admin API). Can also be changed at runtime with the `/responsemode` Telegram command (persists until pod restart)

#### Telegram Commands

| Command | Description |
|---------|-------------|
| `/responsemode` | Show current response mode |
| `/responsemode text` | Force text-only responses |
| `/responsemode voice` | Force voice-only responses (requires TTS) |
| `/responsemode mirror` | Reply in the same format as the input |
| `/responsemode both` | Reply with both text and voice |
| `/responsemode reset` | Revert to the config file default |

#### Response Modes

| Mode | Behavior |
|------|----------|
| `text` | Always reply with text only (default) |
| `voice` | Always reply with voice only (requires TTS) |
| `mirror` | Reply in the same format as the input (text→text, voice→voice) |
| `both` | Reply with both text and voice (requires TTS) |

## Development

### Build Commands

```bash
make build              # Build to bin/magec-server
make dev                # Build and run with config.yaml
make swagger            # Regenerate Swagger docs from annotations
make clean              # Remove build artifacts
make download-model     # Download wake word + pretrained models (interactive)
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

### Backend Types

Backends are created via the Admin API (`POST /api/v1/admin/backends`).

| Type | Description | Required Fields |
|------|-------------|-----------------|
| `openai` | OpenAI-compatible API (OpenAI, Ollama, LM Studio, etc.) | `url` and/or `apiKey` |
| `anthropic` | Anthropic Claude API | `apiKey` |
| `gemini` | Google Gemini API | `apiKey` |

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

```bash
cd deploy/docker/fully-local
docker compose up -d
```

### OpenAI (`deploy/docker/remote-openai/`)

Only infra runs locally. LLM, STT, TTS, and embeddings use OpenAI APIs.

**Services:**
- **magec** - Main server (port 8080) + Admin UI (port 8081)
- **redis** - Session storage
- **postgres** - Long-term memory (pgvector)

```bash
cd deploy/docker/remote-openai
export OPENAI_API_KEY=sk-...
docker compose up -d
```

### Legacy (`deploy/docker/legacy/`)

Original hybrid deployment. Still functional but superseded by the above.

## Dependencies

**Go backend:**
- `google.golang.org/adk` - Agent Development Kit
- `github.com/achetronic/adk-utils-go` - ADK utilities (providers, session, memory, tools)
- `github.com/modelcontextprotocol/go-sdk` - MCP client
- `github.com/yalue/onnxruntime_go` - ONNX runtime for wake word
- `gopkg.in/yaml.v3` - YAML config parsing

**Frontend (CDN):**
- Tailwind CSS - Styling

## Wake Word Models

Models in `models/`, configured in `wakewords.yaml`:

| ID | Phrase | Threshold |
|----|--------|-----------|
| `oye-magec` | "Oye Magec" | 0.5 |
| `magec` | "Magec" | 0.3 |

Pretrained models in `pretrained/`:
- `mel-spectrogram.onnx` - Audio to mel-spectrogram
- `speech-embedding.onnx` - Mel to speech embedding
- `silero-vad.onnx` - Voice activity detection

## Gotchas

1. **Voice-UI has no build step**: Dependencies loaded from CDN. No npm required. **Admin-UI uses Vite** — `make build-admin` or `cd admin-ui && npx vite build`.

2. **Voice detection is server-side**: All ONNX inference (wake word + VAD) happens on the server via WebSocket.

3. **Wake word models location**: ONNX models in `models/` at project root.

4. **VAD stops recording**: When VAD detects speech end, recording automatically stops. No volume-threshold fallback needed.

5. **Memory is optional**: Without Redis/PostgreSQL, sessions are in-memory and long-term memory is disabled.

6. **MCP transports**: Supports both HTTP (StreamableClientTransport) and stdio (CommandTransport).

7. **PWA over HTTP**: Requires Chrome flag for non-localhost addresses.

8. **Telegram voice**: Requires TTS backend configured for voice responses. ffmpeg required in container for OGG→WAV conversion.

9. **Security**: Always set `allowedUsers` in Telegram config to restrict access.

10. **Telegram user context**: The LLM receives a `[context: telegram_user_id: ..., telegram_username: @..., ...]` prefix on every message. This is injected by the Telegram client, not configurable.

11. **Memory providers use connectionString**: Both Redis and Postgres use a universal `connectionString` field (`redis://...`, `postgres://...`). Provider-specific extra fields (like `ttl` for Redis) live in the `config` map alongside the connection string.

12. **Memory provider registry**: Providers register via `init()` + blank imports in `main.go`. Adding a new type = new package under `server/memory/<name>/` + blank import. The `Provider` interface requires `Type()`, `DisplayName()`, `SupportedCategories()`, `ConfigFields()`, and `Ping(ctx, config)`.

13. **Schema-driven memory forms**: The admin UI renders memory provider forms dynamically from `ConfigFields()` specs returned by `/memory/types`. Zero hardcoded fields per provider type — new providers get forms for free.

14. **Category is per-instance, not per-type**: `MemoryProvider.Category` (string in store) vs `Provider.SupportedCategories()` (capability). A single type like Redis could serve both session and long-term roles.

15. **Voice endpoint proxy forwards API keys**: `serveSpeechProxy` and `serveTranscriptionProxy` forward the backend's `apiKey` as `Authorization: Bearer` to the upstream. Without this, backends that require API keys (like OpenAI Edge TTS) return 401.

16. **Immutable IDs**: All entities use UUID v4 (`google/uuid`). Cross-references store IDs not names. On first load, `migrateIDs()` auto-generates UUIDs for entities without them and rewrites name-based cross-references to IDs. Idempotent — uses `isUUID()` regex to skip already-migrated fields.

17. **Hot-reload**: Store changes (via admin API) fire `OnChange()` channel → `agentRouterHandler` rebuilds the ADK agent with 500ms debounce. No server restart needed for config changes.

18. **Multi-agent routing**: ADK uses `appName` in requests to route to the correct agent via `NewMultiLoader`. Voice-UI must send the correct `appName` matching the agent ID in the store.

19. **Admin UI dialog validation**: Cancel/close buttons use `formnovalidate` to bypass HTML5 validation on required fields. Without this, dialogs with required empty fields cannot be dismissed.

20. **Flow editor is nested HTML, not a graph library**: The flow data model is a strict recursive tree (not a DAG/graph), so Vue Flow was replaced with nested HTML divs + `vuedraggable`. `FlowCanvas.vue` handles pan/zoom via pointer events + CSS transforms. `FlowBlock.vue` is recursive. Each step gets a `__key` property for vuedraggable's `item-key` — stripped by `cleanStep()` in `FlowDialog.vue` before API calls.

21. **Flow editor drag `.stop` modifiers are critical**: The toolbar→canvas drop uses HTML5 drag events. All three events on the container drop zone (`@dragover.prevent.stop`, `@dragleave.stop`, `@drop.prevent.stop`) need `.stop` to prevent event bubbling to parent containers. Without this, dropping into a nested container adds the item to both the target and its parent.

22. **OutputKey on AgentDefinition, not FlowStep**: ADK's `OutputKey` (saves agent output to session state under a key) is set on the `AgentDefinition`, not per flow step. It's passed to `llmagent.Config` when the agent is created in `agent.go`. This keeps flows simple — they just look up pre-built agents by ID. The output key is configured in the Agents tab of the admin UI.

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
