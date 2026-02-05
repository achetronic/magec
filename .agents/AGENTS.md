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
cd deploy/docker-compose
nano config.yaml  # Add your LLM API key
docker-compose up -d

# OR from source
make infra        # Start PostgreSQL + Redis
make dev          # Build and run
```

Open http://localhost:8080 and start chatting.

## Architecture

```
magec/
├── server/                     # Go backend (core)
│   ├── main.go                 # HTTP server, routing, middleware
│   ├── agent/agent.go          # ADK agent with memory, MCP tools
│   ├── config/config.go        # YAML config parsing, backend resolution
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
├── models/                     # Wake word ONNX models + wakewords.yaml
├── pretrained/                 # Shared ONNX models (mel-spec, VAD, embeddings)
├── scripts/
│   └── download-model.go       # Wake word model downloader
├── deploy/
│   └── docker-compose/         # Production deployment
├── config.example.yaml
├── Dockerfile
├── Makefile
└── README.md
```

### Component Overview

| Component | Purpose |
|-----------|---------|
| `server/main.go` | HTTP server with ADK REST handler, Whisper/TTS proxies, WebSocket voice-events |
| `server/agent/agent.go` | ADK agent with memory tools, MCP integration |
| `server/voice/` | Server-side voice detection (wake word + VAD) via ONNX |
| `server/clients/telegram/` | Telegram bot with voice message support |
| `server/config/config.go` | YAML config with backend resolution and env var expansion |
| `voice-ui/src/app.js` | Main entry - MagecApp class orchestrates audio pipeline |
| `voice-ui/src/audio/` | AudioCapture, AudioRecorder, AudioConverter, FeedbackSound, OpenAITTS, VoiceEventsClient |
| `voice-ui/src/api/AgentClient.js` | Agent API client (sessions, messages) |
| `voice-ui/src/ui/UIController.js` | DOM manipulation, panels, notifications, sidebar |
| `voice-ui/src/ui/WaveformRenderer.js` | "Centella/Magec" audio visualizer with particles |

## HTTP Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET/POST | `/api/v1/agent/*` | ADK REST API (sessions, run, events) |
| POST | `/api/v1/agent/run` | Run agent (blocking) |
| POST | `/api/v1/agent/run_sse` | Run agent (SSE streaming) |
| POST | `/api/v1/transcription/*` | Proxy to Whisper backend |
| POST | `/api/v1/tts/*` | Proxy to TTS backend |
| WebSocket | `/api/v1/voice-events` | Voice events stream (wake word + VAD) |
| GET | `/api/v1/health` | Health check |
| GET | `/` | Static files from `voice-ui/` |

## Configuration (YAML)

Single YAML config file. Supports `${VAR}` for environment variable expansion.

```yaml
server:
  host: 0.0.0.0
  port: 8080

log:
  level: info   # debug, info, warn, error
  format: console  # console, json

# Reusable AI backends
backends:
  - name: ollama
    type: openai
    url: http://localhost:11434/v1

  - name: openai
    type: openai
    apiKey: ${OPENAI_API_KEY}

  - name: anthropic
    type: anthropic
    apiKey: ${ANTHROPIC_API_KEY}

  - name: parakeet
    type: openai
    url: http://127.0.0.1:5000/v1

# Server-side wake word detection
wakeWord:
  enabled: true

transcription:
  backend: parakeet
  model: whisper-1

llm:
  backend: ollama
  model: qwen3:8b

# Optional: customize agent behavior
agent:
  systemPrompt: "Custom system prompt..."
  systemPromptSuffix: "Additional instructions..."

# Optional: text-to-speech
tts:
  backend: openai
  model: tts-1
  voice: alloy        # or es-ES-AlvaroNeural for edge-tts
  speed: 1.0

memory:
  session:
    redis:
      address: localhost:6379
      password: ""
      db: 0
      ttl: 24h

  longTerm:
    embedding:
      backend: ollama
      model: nomic-embed-text
    postgres:
      connectionString: postgres://postgres:postgres@localhost:5432/magec?sslmode=disable

# MCP tool servers (HTTP or stdio)
mcpServers:
  - name: home-assistant
    type: http
    endpoint: http://localhost:8070/mcp
    headers:
      Authorization: Bearer ${HASS_TOKEN}
    systemPrompt: "Home automation tools"

  - name: local-tool
    type: stdio
    command: /path/to/tool
    args: ["--flag"]
    env:
      KEY: value
    workDir: /path/to/dir

# Clients
clients:
  telegram:
    enabled: true
    token: ${TELEGRAM_BOT_TOKEN}
    allowedUsers: [123456789]
    allowedChats: [-100123456789]
    voiceResponses: true
```

### Backend Types

| Type | Description | Required Fields |
|------|-------------|-----------------|
| `openai` | OpenAI-compatible API (OpenAI, Ollama, LM Studio, etc.) | `url` and/or `apiKey` |
| `anthropic` | Anthropic Claude API | `apiKey` |
| `gemini` | Google Gemini API | `apiKey` |

## Code Patterns

### Go Conventions

- **YAML config**: Single source of truth, env var expansion with `${VAR}`
- **Backend resolution**: Config resolves backend references at load time
- **ADK REST handler**: `adkrest.NewHandler()` provides full ADK API
- **Memory tools**: Agent has `search_memory` and `save_to_memory` tools
- **Agent instruction**: Reads memories at conversation start
- **WebSocket voice-events**: Server handles all ONNX inference (wake word + VAD), clients stream audio

### JavaScript Conventions (voice-ui)

- **ES Modules**: Uses `import` from CDN (no build step)
- **Class-based**: `MagecApp` orchestrates all components
- **i18n**: `t('key', {params})` function, `data-i18n` attributes in HTML
- **Storage keys**: `magec_settings`, `magec_language`, `magec_sessions`
- **Color palette**: piedra (grays), atlantico (cyan), lava (red), sol (yellow/orange), arena (text)
- **Centralized errors**: All API errors flow through ErrorHandler → notifications

### Audio Processing Pipeline (voice-ui)

1. **Microphone capture** → AudioCapture with AudioWorklet at 16kHz
2. **Voice events** → Audio streamed via WebSocket to server → wake word + VAD detection
3. **Recording** → AudioRecorder captures audio as webm (started by wake word, stopped by VAD)
4. **Conversion** → AudioConverter resamples to 16kHz WAV
5. **Transcription** → RemoteTranscriber POSTs to Whisper API
6. **Agent interaction** → AgentClient POSTs to ADK `/run` endpoint
7. **TTS** → OpenAITTS plays response via `/api/v1/tts/speech`

### Voice Events (Server-side)

```
Client (WebSocket)          Server (/api/v1/voice-events)
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
      "models": [{"id": "oye-magec", "phrase": "Oye Magec"}],
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
- **Voice messages**: Downloaded → transcribed → processed → optional voice response
- **Authorization**: `allowedUsers` and `allowedChats` allowlists
- **Sessions**: Scoped by `telegram_<chatID>`

## Development

### Build Commands

```bash
make build              # Build to bin/magec-server
make dev                # Build and run with config.yaml
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

### Provider Examples

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

## Docker Compose Deployment

Production deployment in `deploy/docker-compose/`:

**Services:**
- **magec** - Main server (ghcr.io/achetronic/magec:latest, port 8080)
- **redis** - Session storage
- **postgres** - Long-term memory (pgvector)
- **parakeet** - Speech-to-text (ghcr.io/achetronic/parakeet:latest)
- **tts** - Text-to-speech (travisvn/openai-edge-tts:latest)

```bash
cd deploy/docker-compose
docker-compose up -d
```

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

1. **No frontend build step**: Dependencies loaded from CDN. No npm/yarn required.

2. **Voice detection is server-side**: All ONNX inference (wake word + VAD) happens on the server via WebSocket.

3. **Wake word models location**: ONNX models in `models/` at project root.

4. **VAD stops recording**: When VAD detects speech end, recording automatically stops. No volume-threshold fallback needed.

5. **Memory is optional**: Without Redis/PostgreSQL, sessions are in-memory and long-term memory is disabled.

6. **MCP transports**: Supports both HTTP (StreamableClientTransport) and stdio (CommandTransport).

7. **PWA over HTTP**: Requires Chrome flag for non-localhost addresses.

8. **Telegram voice**: Requires TTS backend configured for voice responses.

9. **Security**: Always set `allowedUsers` in Telegram config to restrict access.

## Testing

Manual testing workflow:

1. Start infrastructure: `make infra`
2. Start server: `make dev`
3. Open http://localhost:8080
4. Allow microphone access
5. Say wake word or tap the orb to start recording
6. Speak - recording stops automatically when you stop talking (VAD)
7. Verify agent responds

## Related Resources

- [Google ADK](https://google.github.io/adk-docs/)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [OpenWakeWord](https://github.com/dscripka/openWakeWord)
- [pgvector](https://github.com/pgvector/pgvector)
- [Parakeet](https://github.com/achetronic/parakeet)
- [openai-edge-tts](https://github.com/travisvn/openai-edge-tts)
- [hass-mcp](https://github.com/achetronic/hass-mcp)
