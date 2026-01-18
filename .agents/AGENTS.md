# AGENTS.md - Magec

Voice assistant from the Canary Islands 🇮🇨 with wake word detection, speech recognition, and AI agent capabilities.

## Project Overview

**Magec** is a voice assistant that runs in the browser with a Go backend. Named after the Guanche god of the Sun (/maˈxek/), it supports:

- **Wake word detection**: OpenWakeWord models (ONNX) trigger recording when user says a custom wake phrase
- **Speech transcription**: Local (Xenova/whisper-small via Transformers.js) or Remote (OpenAI-compatible API)
- **AI Agent**: ADK-based agent with multi-provider LLM support (OpenAI, Anthropic, Gemini, Ollama)
- **Long-term memory**: PostgreSQL with pgvector embeddings for persistent memory across sessions
- **Session memory**: Redis-backed conversation history with configurable TTL
- **MCP toolsets**: External tools via Model Context Protocol (Streamable HTTP)
- **PWA support**: Installable as standalone app on mobile devices
- **i18n**: Supports Spanish (default) and English

## Quick Start

```bash
# Download wake word models (interactive - also downloads pretrained)
make download-model

# Start infrastructure
make infra

# Start dev server
make dev
```

## Architecture

```
magec/
├── gui/                    # Frontend (HTML/CSS/JS)
│   ├── src/
│   │   ├── app.js          # Main application (MagecApp class)
│   │   ├── config.js       # Configuration
│   │   ├── audio/          # Audio processing modules
│   │   ├── i18n/           # Internationalization (es.js, en.js)
│   │   ├── ui/             # UI components (UIController, WaveformRenderer)
│   │   ├── session/        # Session management
│   │   ├── settings/       # Settings persistence
│   │   └── utils/          # Utilities (WakeLock)
│   ├── models/             # Wake word ONNX models + wakewords.json config
│   ├── assets/             # Logo, banner, PWA icons
│   ├── manifest.json       # PWA manifest
│   └── index.html
├── server/                 # Go backend
│   ├── main.go             # HTTP server, routing
│   ├── agent/agent.go      # ADK agent service
│   ├── config/config.go    # YAML config parsing
│   └── logging/logging.go  # Structured logging
├── scripts/
│   └── download-model.go   # Wake word model downloader (Go)
├── config.example.yaml     # Full config template
├── Makefile
└── README.md
```

### Component Overview

| Component | Purpose |
|-----------|---------|
| `gui/src/app.js` | Main entry - MagecApp class orchestrates audio pipeline, wake word, transcription, agent chat |
| `gui/src/config.js` | Frontend configuration (API endpoints, storage keys) |
| `gui/src/i18n/` | Translations (es.js, en.js) with `t()` function and `data-i18n` attributes |
| `gui/src/ui/UIController.js` | DOM manipulation, dynamic translations, settings panel |
| `gui/src/ui/WaveformRenderer.js` | "Magec" audio visualizer |
| `server/main.go` | HTTP server with ADK REST handler, Whisper proxy |
| `server/agent/agent.go` | ADK agent with memory tools, MCP integration |
| `server/config/config.go` | YAML config with backend resolution and env var expansion |

## Configuration (YAML)

Magec uses a single YAML config file for all settings:

```yaml
server:
  host: 0.0.0.0  # Use 127.0.0.1 to restrict to localhost
  port: 8080

log:
  level: info  # debug, info, warn, error
  format: console  # console, json

# Reusable AI backends
backends:
  - name: ollama
    type: openai
    url: http://localhost:11434/v1

  - name: anthropic
    type: anthropic
    apiKey: ${ANTHROPIC_API_KEY}

  - name: local-whisper
    type: whisper
    url: http://127.0.0.1:5092/v1

transcription:
  backend: local-whisper
  model: parakeet

llm:
  backend: ollama
  model: qwen3:8b

memory:
  session:
    redis:
      address: localhost:6379
      ttl: 24h

  longTerm:
    embedding:
      backend: ollama
      model: nomic-embed-text
    postgres:
      connectionString: postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable

mcpServers:
  - name: hass-mcp
    endpoint: http://localhost:8070/mcp
```

### Backend Types

| Type | Description | Required Fields |
|------|-------------|-----------------|
| `openai` | OpenAI-compatible API (LLM, Whisper, TTS, embeddings) | `url` or default, `apiKey` optional |
| `anthropic` | Anthropic Claude API | `apiKey` |
| `gemini` | Google Gemini API | `apiKey` |

The backend type indicates the API protocol, not the function. The function is determined by where the backend is used (`transcription`, `llm`, `tts`, `memory.longTerm.embedding`).

## Code Patterns

### Go Conventions

- **YAML config**: Single source of truth, env var expansion with `${VAR}`
- **Backend resolution**: Config resolves backend references at load time
- **ADK REST handler**: `adkrest.NewHandler()` provides full ADK API
- **Memory tools**: Agent has `search_memory` and `save_to_memory` tools
- **Agent instruction**: Includes directive to read memories at conversation start

### JavaScript Conventions

- **ES Modules**: Uses `import` from CDN (Transformers.js, ONNX Runtime, Hey-Buddy)
- **Class-based**: `MagecApp` orchestrates all components
- **i18n**: `t('key', {params})` function, `data-i18n` attributes in HTML
- **Storage keys**: `magec_settings`, `magec_language`, `magec_sessions`
- **Color palette**: piedra (grays), atlantico (cyan), lava (red), sol (yellow/orange), arena (text)

### Audio Processing Pipeline

1. **Microphone capture** → AudioContext at 16kHz sample rate
2. **Wake word detection** → OpenWakeWord ONNX models trigger recording
4. **Recording** → MediaRecorder captures audio as webm
5. **Transcription** → Local (Transformers.js) or remote (Whisper API proxy)
6. **Agent interaction** → ADK REST `/run_sse` endpoint with streaming
7. **TTS** → Browser speech synthesis for responses

## Agent Behavior

The agent (defined in `server/agent/agent.go`) has these key behaviors:

1. **Memory reading at start**: Agent is instructed to call `search_memory` at the beginning of every conversation to load user preferences and stored information

2. **Proactive memory saving**: When users share preferences or important info, agent saves to memory

3. **Language matching**: Responds in the same language as user input

4. **Concise responses**: Optimized for voice interaction

## UI Components

### Magec (Audio Visualizer)

The `WaveformRenderer.js` contains the "Magec" visualizer:

- **Mystical circle** that breathes and reacts to audio
- **Sleeping (yellow)**: Slow floating glitter, subtle internal spirals
- **Awake (red)**: Fast glitter, more particles, more energy
- **Smooth transitions** between states (color, size, particle speed)
- **Max size capped** at 384px to prevent oversized rendering on large screens
- **Canvas fills full vertical** space for effects to render without clipping
- **Interaction**: Tapping Magec activates/deactivates recording

### Settings Panel

- Wake word toggle (enable/disable voice activation)
- Wake word model selector (dynamically loaded from `wakewords.json`)
- STT mode selector (server/browser transcription)
- TTS toggle (enable/disable voice responses)
- Language selector (Spanish/English)
- Session management (view/delete past sessions)

## Development

### Build Commands

```bash
make build              # Build to bin/magec-server
make dev                # Build and run with config.yaml
make clean              # Remove generated files
make download-model     # Download wake word + pretrained models (interactive)
```

### Infrastructure (Docker)

```bash
make postgres           # Start PostgreSQL with pgvector
make redis              # Start Redis
make ollama             # Start Ollama with qwen3:8b + nomic-embed-text
make infra              # Start postgres + redis
make infra-stop         # Stop infrastructure
make infra-clean        # Remove all containers and volumes
```

### Using with Different Providers

**Ollama (local):**
```yaml
backends:
  - name: ollama
    type: openai
    url: http://localhost:11434/v1

llm:
  backend: ollama
  model: qwen3:8b
```

**OpenAI:**
```yaml
backends:
  - name: openai
    type: openai
    apiKey: ${OPENAI_API_KEY}

llm:
  backend: openai
  model: gpt-4o-mini
```

**Anthropic:**
```yaml
backends:
  - name: anthropic
    type: anthropic
    apiKey: ${ANTHROPIC_API_KEY}

llm:
  backend: anthropic
  model: claude-sonnet-4-20250514
```

**Gemini:**
```yaml
backends:
  - name: gemini
    type: gemini
    apiKey: ${GEMINI_API_KEY}

llm:
  backend: gemini
  model: gemini-2.0-flash
```

## PWA Installation

Magec can be installed as a standalone app:

1. **Android**: Chrome menu → "Add to Home screen"
2. **iOS**: Safari share → "Add to Home Screen"

**Note**: PWA features require HTTPS. For local network access over HTTP, add exception in `chrome://flags/#unsafely-treat-insecure-origin-as-secure`

## Dependencies

**Go backend:**
- `google.golang.org/adk` - Agent Development Kit
- `github.com/achetronic/adk-utils-go` - ADK utilities (providers, session, memory, tools)
- `github.com/modelcontextprotocol/go-sdk` - MCP client
- `gopkg.in/yaml.v3` - YAML config parsing

**Frontend (CDN):**
- `@huggingface/transformers` - Whisper inference
- `onnxruntime-web` - ONNX model inference
- `onnxruntime-web` - Wake word model inference (OpenWakeWord)
- Tailwind CSS - Styling

## Gotchas

1. **No build step**: Frontend dependencies loaded from CDN. No npm/yarn required.

2. **Wake word models**: OpenWakeWord ONNX models in `gui/models/`. Config in `wakewords.json`.

3. **Infrastructure required**: Redis and PostgreSQL (pgvector) must be running.

4. **Whisper backend optional**: Remote transcription requires a Whisper-compatible API.

5. **MCP via Streamable HTTP**: Uses 2025-03-26 spec with `mcp.StreamableClientTransport`.

6. **PWA over HTTP**: Requires Chrome flag for non-localhost addresses.

7. **Memory at conversation start**: Agent reads memories at the beginning of each conversation - user can store preferences and instructions there.

## Testing

Manual testing workflow:

1. Start infrastructure: `make infra`
2. Start server: `make dev`
3. Open http://localhost:8080
4. Allow microphone access
5. Wait for models to load
6. Say wake word (e.g., "Oye Magec") or tap Magec to record
7. Speak and release to transcribe
8. Verify agent responds

## Related Resources

- [Google ADK](https://google.github.io/adk-docs/)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [OpenWakeWord](https://github.com/dscripka/openWakeWord)
- [Xenova Transformers.js](https://huggingface.co/docs/transformers.js)
- [pgvector](https://github.com/pgvector/pgvector)
