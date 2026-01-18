# Magec

<p align="center">
  <img src="gui/assets/banner.png" alt="Magec - Voice assistant from Canary Islands" width="800">
</p>

## Why "Magec"?

**Magec** (/maˈxek/) was the god of the Sun worshipped by the Guanches, the aboriginal Berber inhabitants of Tenerife in the Canary Islands. Among their pantheon of deities, Magec represented light and warmth—essential forces that sustained life on the island.

The name honors this Canarian heritage while reflecting the assistant's purpose: to illuminate and assist.

## Screenshots

<p align="center">
  <img src="docs/img/home.png" alt="Home - Magec visualizer" width="280">
  <img src="docs/img/chat.png" alt="Chat conversation" width="280">
</p>
<p align="center">
  <img src="docs/img/settings.png" alt="Settings panel" width="280">
  <img src="docs/img/notifications.png" alt="Notifications" width="280">
</p>

## Features

### Voice & Speech
- **Wake word detection** - Custom OpenWakeWord models for hands-free activation ("Oye Magec", "Magec")
- **Speech transcription** - Local transcription via Whisper or remote through OpenAI-compatible APIs
- **Text-to-speech** - Natural voice responses via OpenAI-compatible TTS APIs

### AI & Intelligence
- **Multi-provider LLM** - Supports OpenAI, Anthropic (Claude), Google Gemini, and local models via Ollama
- **Configurable backends** - Define multiple AI backends and switch between them easily
- **ADK-based agent** - Built on Agent Development Kit for structured AI interactions

### Memory & Context
- **Session memory** - Redis-backed conversation history with configurable TTL
- **Long-term memory** - PostgreSQL with vector embeddings for persistent knowledge across sessions
- **Embedding models** - Configurable embedding backends for semantic search

### Extensibility
- **MCP toolsets** - Connect external tools via Model Context Protocol (filesystem, GitHub, Home Assistant, etc.)
- **YAML configuration** - Single config file for all settings: backends, memory, MCP servers

### Interface
- **Magec visualizer** - Mystical audio visualization that reacts to voice
- **PWA support** - Install as standalone app on mobile devices
- **Responsive design** - Works on desktop and mobile browsers
- **i18n** - Supports Spanish (default) and English

## Requirements

- Go 1.21+
- Docker (for infrastructure)
- Modern browser with WebAudio support

## Quick Start

```bash
# 1. Start infrastructure (PostgreSQL with pgvector + Redis)
make infra

# 2. Copy and edit config
cp config.example.yaml config.yaml
# Edit config.yaml with your API keys and settings

# 3. Build and run
make dev
```

Open http://localhost:8080

## Configuration

Magec uses a single YAML configuration file. The server accepts one flag:

```bash
./bin/magec-server -config path/to/config.yaml
```

### Config File Structure

```yaml
server:
  host: 0.0.0.0  # Use 127.0.0.1 to restrict to localhost
  port: 8080

log:
  level: info    # debug, info, warn, error
  format: console # console, json

# Reusable AI backends
backends:
  - name: ollama
    type: openai
    url: http://localhost:11434/v1

  - name: openai-cloud
    type: openai
    apiKey: ${OPENAI_API_KEY}

  - name: anthropic-cloud
    type: anthropic
    apiKey: ${ANTHROPIC_API_KEY}

  - name: local-whisper
    type: openai
    url: http://127.0.0.1:5000/v1

# Transcription settings
transcription:
  backend: local-whisper
  model: whisper-1

# LLM settings
llm:
  backend: ollama
  model: qwen3:8b

# TTS settings (optional)
tts:
  backend: openai-cloud
  model: tts-1
  voice: alloy
  speed: 1.0

# Memory configuration (optional)
# Without this section, sessions are stored in-memory and long-term memory is disabled
memory:
  # Session storage (optional - defaults to in-memory)
  session:
    redis:
      address: localhost:6379
      password: ""
      db: 0
      ttl: 24h

  # Long-term memory (optional - disabled if not configured)
  longTerm:
    embedding:
      backend: ollama
      model: nomic-embed-text
    postgres:
      connectionString: postgres://postgres:postgres@localhost:5432/magec?sslmode=disable

# MCP tool servers
mcpServers:
  - name: home-assistant
    endpoint: http://localhost:8070/mcp
```

### Backend Types

| Type | Description | Required Fields |
|------|-------------|-----------------|
| `openai` | OpenAI-compatible API (OpenAI, Ollama, LM Studio, etc.) | `url` and/or `apiKey` |
| `anthropic` | Anthropic Claude API | `apiKey` |
| `gemini` | Google Gemini API | `apiKey` |

### Environment Variables

Config values support `${VAR}` syntax for environment variable expansion:

```yaml
backends:
  - name: openai
    type: openai
    apiKey: ${OPENAI_API_KEY}  # Reads from environment
```

## Examples

### LLM with Ollama (fully local)

```yaml
backends:
  - name: ollama
    type: openai
    url: http://localhost:11434/v1

llm:
  backend: ollama
  model: qwen3:8b
```

```bash
make ollama  # Start Ollama with required models
make dev
```

### LLM with OpenAI

```yaml
backends:
  - name: openai
    type: openai
    apiKey: ${OPENAI_API_KEY}

llm:
  backend: openai
  model: gpt-4o-mini
```

### LLM with Anthropic

```yaml
backends:
  - name: anthropic
    type: anthropic
    apiKey: ${ANTHROPIC_API_KEY}

llm:
  backend: anthropic
  model: claude-sonnet-4-20250514
```

### Session Memory with Redis

Persist conversation history across server restarts:

```bash
docker run -d --name redis -p 6379:6379 redis:alpine
```

```yaml
memory:
  session:
    redis:
      address: localhost:6379
      ttl: 24h
```

### Long-Term Memory with PostgreSQL

Enable persistent memory that survives across sessions. Requires both PostgreSQL (with pgvector) and an embedding model:

```bash
docker run -d --name postgres \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=magec \
  pgvector/pgvector:pg17
```

```yaml
backends:
  - name: ollama
    type: openai
    url: http://localhost:11434/v1

memory:
  longTerm:
    embedding:
      backend: ollama
      model: nomic-embed-text
    postgres:
      connectionString: postgres://postgres:postgres@localhost:5432/magec?sslmode=disable
```

Or with OpenAI embeddings:

```yaml
backends:
  - name: openai
    type: openai
    apiKey: ${OPENAI_API_KEY}

memory:
  longTerm:
    embedding:
      backend: openai
      model: text-embedding-3-small
    postgres:
      connectionString: postgres://postgres:postgres@localhost:5432/magec?sslmode=disable
```

### STT with Parakeet (local Whisper)

[Parakeet](https://github.com/achetronic/parakeet) is a lightweight OpenAI-compatible Whisper server.

```bash
docker run -d --name parakeet \
  -p 5000:5000 \
  ghcr.io/achetronic/parakeet:latest
```

```yaml
backends:
  - name: parakeet
    type: openai
    url: http://localhost:5000/v1

transcription:
  backend: parakeet
  model: whisper-1
```

### TTS with openai-edge-tts (local)

[openai-edge-tts](https://github.com/travisvn/openai-edge-tts) provides free TTS using Microsoft Edge's speech service with an OpenAI-compatible API.

```bash
docker run -d --name edge-tts \
  -p 5050:5050 \
  travisvn/openai-edge-tts:latest
```

```yaml
backends:
  - name: edge-tts
    type: openai
    url: http://localhost:5050/v1
    apiKey: "dummy"  # Required but not used

tts:
  backend: edge-tts
  model: tts-1
  voice: es-ES-AlvaroNeural  # Spanish voice
  speed: 1.0
```

Available Spanish voices: `es-ES-AlvaroNeural`, `es-ES-ElviraNeural`, `es-MX-DaliaNeural`, `es-MX-JorgeNeural`

### MCP with Home Assistant

[hass-mcp](https://github.com/achetronic/hass-mcp) exposes Home Assistant as an MCP server, allowing Magec to control your smart home.

1. Create a config file `hass-mcp.yaml`:

```yaml
server:
  transport:
    type: http
    http:
      host: ":8080"

home_assistant:
  url: http://homeassistant.local:8123
  token: YOUR_LONG_LIVED_ACCESS_TOKEN
```

2. Run the container:

```bash
docker run -d --name hass-mcp \
  -p 8070:8080 \
  -v $(pwd)/hass-mcp.yaml:/app/config.yaml \
  ghcr.io/achetronic/hass-mcp:latest \
  --config /app/config.yaml
```

3. Add to Magec config:

```yaml
mcpServers:
  - name: home-assistant
    endpoint: http://localhost:8070/mcp
```

### MCP Tools

```yaml
mcpServers:
  - name: filesystem
    endpoint: http://localhost:8080/mcp

  - name: github
    endpoint: https://mcp-github.example.com/mcp
    headers:
      Authorization: Bearer ${GITHUB_TOKEN}
```

## Docker

### Using the image

```bash
docker run -d --name magec \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  ghcr.io/achetronic/magec:latest
```

### Building locally

```bash
docker build -t magec .
```

## Infrastructure

```bash
# Start PostgreSQL (pgvector) and Redis
make infra

# Start Ollama with models (optional, for local LLM)
make ollama

# Stop infrastructure
make infra-stop

# Remove all containers and volumes
make infra-clean
```

### Individual Services

```bash
make postgres   # PostgreSQL with pgvector extension
make redis      # Redis for session storage
make ollama     # Ollama with qwen3:8b + nomic-embed-text
```

## Mobile Installation (PWA)

Magec can be installed as a standalone app on Android and iOS devices.

### Android (Chrome)

1. Open Magec in Chrome
2. Tap the menu (⋮) → "Add to Home screen" or "Install app"
3. The app will open without browser UI

### iOS (Safari)

1. Open Magec in Safari
2. Tap Share (□↑) → "Add to Home Screen"
3. The app will open in standalone mode

### HTTP without HTTPS

PWA features require HTTPS, except for `localhost`. If accessing Magec from a local network IP (e.g., `http://192.168.x.x:8080`), you need to add a security exception in Chrome:

1. Go to `chrome://flags/#unsafely-treat-insecure-origin-as-secure`
2. Add your server URL (e.g., `http://192.168.2.197:8080`)
3. Set to "Enabled"
4. Restart Chrome

After this, PWA installation and features will work over HTTP.

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/agent/*` | * | ADK REST API (sessions, run, events) |
| `/api/v1/agent/run` | POST | Run agent (blocking) |
| `/api/v1/agent/run_sse` | POST | Run agent (SSE streaming) |
| `/api/v1/transcription/*` | POST | Proxy to Whisper backend |
| `/api/v1/tts/*` | POST | Proxy to TTS backend |
| `/api/v1/health` | GET | Health check |

### Agent Request Example

```bash
curl -X POST http://localhost:8080/api/v1/agent/run \
  -H "Content-Type: application/json" \
  -d '{
    "app_name": "magec_agent",
    "user_id": "user123",
    "session_id": "session456",
    "new_message": {
      "role": "user",
      "parts": [{ "text": "Hello!" }]
    }
  }'
```

## Development

### Project Structure

```
magec/
├── gui/                  # Frontend (HTML/CSS/JS)
│   ├── src/              # Application source
│   │   ├── app.js        # Main application (MagecApp class)
│   │   ├── config.js     # Frontend configuration
│   │   ├── audio/        # Audio processing modules
│   │   ├── i18n/         # Translations (es.js, en.js)
│   │   ├── ui/           # UI components
│   │   └── ...
│   ├── assets/           # Logo, banner, PWA icons
│   ├── models/           # Wake word ONNX models + wakewords.json
│   └── manifest.json     # PWA manifest
├── server/               # Go backend
│   ├── main.go           # HTTP server
│   ├── agent/            # ADK agent service
│   ├── config/           # YAML config parsing
│   └── logging/          # Structured logging
├── config.example.yaml   # Config template
├── Dockerfile
└── Makefile
```

### Make Commands

```bash
make help           # Show all available commands
make build          # Build to bin/magec-server
make dev            # Build and run with config.yaml
make clean          # Remove build artifacts
```

### Dependencies

**Go backend:**
- `google.golang.org/adk` - Agent Development Kit
- `github.com/achetronic/adk-utils-go` - ADK utilities (providers, session, memory, tools)
- `github.com/modelcontextprotocol/go-sdk` - MCP client

**Frontend (CDN, no build step):**
- `@huggingface/transformers` - Whisper inference (browser)
- `onnxruntime-web` - OpenWakeWord model inference
- Tailwind CSS - Styling

## License

MIT
