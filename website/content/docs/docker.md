---
title: "Docker Deployment"
---

Magec is designed to run as a set of Docker containers. The [install script](/magec/docs/getting-started/) handles everything automatically, but this page explains what's happening under the hood — the container architecture, the different deployment modes, how seeds work, and how to customize your setup.

## Container architecture

A full Magec deployment includes several containers working together:

| Container | Purpose | Always present |
|-----------|---------|---------------|
| **magec** | The Magec server — API, Admin UI, Voice UI, agent runtime | Yes |
| **redis** | Session memory storage | Yes |
| **postgres** | Long-term memory storage (pgvector) | Yes |
| **ollama** | Local LLM and embeddings (Qwen 3, nomic-embed-text) | Local mode only |
| **ollama-setup** | Downloads Ollama models on first start, then exits | Local mode only |
| **parakeet** | Local speech-to-text (NVIDIA Parakeet) | Local, Anthropic, Gemini |
| **tts** | Local text-to-speech (OpenAI Edge TTS) | Local, Anthropic, Gemini |

In cloud modes (OpenAI, Anthropic, Gemini), some containers are replaced by cloud API calls. Redis and PostgreSQL always run locally because they store your data.

## Docker Compose files

The installer downloads two Docker Compose files:

### docker-compose.yaml (base)

The base file defines all possible services with their default configuration. It includes Magec, Redis, PostgreSQL, Ollama, Parakeet, and TTS containers.

### docker-compose.override.yaml (provider-specific)

The override file is selected based on your deployment mode. It adjusts which services are active and sets environment variables for the chosen provider:

| Mode | Override file | What it configures |
|------|--------------|-------------------|
| `--local` | Local override | All services active. Ollama pulls `qwen3:8b` and `nomic-embed-text`. |
| `--openai` | OpenAI override | Disables Ollama, Parakeet, and TTS. Sets `OPENAI_API_KEY` from env. |
| `--anthropic` | Anthropic override | Disables only the Ollama LLM. Keeps Parakeet and TTS local. Sets `ANTHROPIC_API_KEY`. |
| `--gemini` | Gemini override | Disables only the Ollama LLM. Keeps Parakeet and TTS local. Sets `GEMINI_API_KEY`. |

## Seeds (MAGEC_SEED)

When Magec starts and finds an empty data store (`data/store.json` doesn't exist or is empty), it loads a **seed** — a pre-built configuration that gives you a working setup immediately.

The seed is controlled by the `MAGEC_SEED` environment variable, which the Docker Compose override sets automatically:

| Seed value | Set by | Pre-configures |
|-----------|--------|---------------|
| `voice-ui` | `--local` | Ollama backend (LLM + embeddings), Redis session memory, PostgreSQL long-term memory, default agent with voice, Voice UI client |
| `voice-ui-openai` | `--openai` | OpenAI backend (LLM + STT + TTS + embeddings), Redis session memory, PostgreSQL long-term memory, default agent, Voice UI client |
| `voice-ui-anthropic` | `--anthropic` | Anthropic backend (LLM) + Ollama (embeddings) + Parakeet (STT) + Edge TTS, memory, agent, client |
| `voice-ui-gemini` | `--gemini` | Gemini backend (LLM) + Ollama (embeddings) + Parakeet (STT) + Edge TTS, memory, agent, client |

### How seeds work

1. Magec starts and checks for `data/store.json`
2. If the store is empty and `MAGEC_SEED` is set, the seed is loaded
3. The seed populates the store with backends, memory providers, agents, and clients
4. The server starts with a fully working configuration
5. On subsequent starts, the seed is ignored — your configuration is yours

Seeds are a one-time setup. After the first run, everything you change through the Admin UI persists in `store.json`. The seed never overwrites your configuration.

### Credentials in seeds

Seeds use `${VAR}` syntax for sensitive values like API keys:

```json
{
  "apiKey": "${OPENAI_API_KEY}"
}
```

This means your API keys come from environment variables at runtime — they're never stored in plain text in the data directory.

## GPU support

For local deployments, you can enable NVIDIA GPU acceleration for Ollama:

```bash
curl -fsSL .../install.sh | bash -s -- --gpu
```

This adds the NVIDIA container runtime configuration to the Ollama service in Docker Compose. You need:

- An NVIDIA GPU
- [nvidia-container-toolkit](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html) installed
- Docker configured to use the NVIDIA runtime

GPU acceleration significantly speeds up local LLM inference and model loading.

## Data persistence

All persistent data is stored in the `data/` directory, which is mounted as a Docker volume:

```
data/
├── store.json           # All configuration (agents, backends, clients, etc.)
└── conversations.json   # Conversation history
```

Redis and PostgreSQL also use Docker volumes for persistence. Your data survives container restarts, image updates, and `docker compose down/up` cycles.

## Customizing your deployment

### Adding more services

You can extend the Docker Compose configuration to add MCP servers or other services:

```yaml
services:
  hass-mcp:
    image: ghcr.io/achetronic/hass-mcp:latest
    environment:
      - HASS_URL=http://homeassistant:8123
      - HASS_TOKEN=${HASS_TOKEN}
    ports:
      - "8888:8080"
```

Then add the MCP server in the Admin UI pointing at `http://hass-mcp:8080/sse`.

### Changing ports

Override the default ports in your Docker Compose:

```yaml
services:
  magec:
    ports:
      - "3000:8080"  # Voice UI + User API on port 3000
      - "3001:8081"  # Admin UI + Admin API on port 3001
```

### Environment variables

All Magec configuration supports environment variable substitution. You can pass variables through Docker Compose:

```yaml
services:
  magec:
    environment:
      - MAGEC_SEED=voice-ui
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - LOG_LEVEL=debug
```

## Common operations

```bash
cd magec                                    # your deployment directory

# Logs
docker compose logs -f                      # all services
docker compose logs -f magec                # Magec server only
docker compose logs -f ollama               # Ollama only

# Lifecycle
docker compose down                         # stop everything
docker compose up -d                        # start everything
docker compose restart magec                # restart Magec only

# Updates
docker compose pull                         # pull latest images
docker compose up -d                        # restart with new versions

# Backup
cp data/store.json store-backup.json        # backup configuration
cp data/conversations.json conv-backup.json # backup conversations

# Reset (start fresh)
docker compose down -v                      # stop and remove volumes
rm -rf data/                                # remove configuration
docker compose up -d                        # fresh start with seed
```

{{< callout type="info" >}}
When updating, `docker compose pull && docker compose up -d` is all you need. Your configuration in `data/` persists across image updates. The seed only runs on first start with an empty store.
{{< /callout >}}
