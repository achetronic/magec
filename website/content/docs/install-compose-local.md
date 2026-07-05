---
title: "Docker Compose"
---

The fully local deployment. Everything runs on your machine: LLM, speech-to-text, text-to-speech, embeddings, and memory. No API keys, no cloud accounts, no data leaving your network.

**Requirements:**
- Docker and Docker Compose
- At least 8 GB of RAM (for Ollama + the LLM model)
- ~6 GB disk space for AI models (downloaded on first start)

## One-line install

The interactive installer walks you through the full setup (LLM provider, voice, memory, NVIDIA GPU support) and generates a ready-to-run Docker Compose deployment:

```bash
curl -fsSL https://magec.dev/install | bash
```

{{< callout type="info" >}}
The first start downloads approximately 5 GB of AI models (Qwen 3 8B for LLM, nomic-embed-text for embeddings). This only happens once. Track progress with `docker compose logs -f ollama-setup`.
{{< /callout >}}

## Manual setup

If you prefer to set things up yourself, or want to customize the configuration before starting:

### 1. Download the files

```bash
mkdir magec && cd magec

curl -fsSL https://raw.githubusercontent.com/achetronic/magec/master/docker/compose/docker-compose.yaml \
  -o docker-compose.yaml

curl -fsSL https://raw.githubusercontent.com/achetronic/magec/master/docker/compose/config.yaml \
  -o config.yaml
```

### 2. (Optional) Enable GPU

Edit `docker-compose.yaml` and uncomment the `deploy` section under the `ollama` service:

```yaml
services:
  ollama:
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: all
              capabilities: [gpu]
```

### 3. Start

```bash
docker compose up -d
```

## What gets deployed

| Container | Purpose | Port |
|-----------|---------|------|
| **magec** | Magec server: Admin UI, Voice UI, API, agent runtime | 8080, 8081 |
| **redis** | Session memory storage | 6379 |
| **postgres** | Long-term memory (pgvector) | 5432 |
| **ollama** | LLM and embeddings (Qwen 3 8B, nomic-embed-text) | 11434 |
| **ollama-setup** | Downloads Ollama models on first start, then exits | — |
| **parakeet** | Speech-to-text (NVIDIA Parakeet) | 5092 |
| **tts** | Text-to-speech (OpenAI Edge TTS) | 5050 |

## Set up your first agent

Once everything is running, open the **Admin UI** at `http://localhost:8081`.

### Create backends

You need three backends: one for the LLM/embeddings, one for STT, and one for TTS.

**Ollama (LLM + Embeddings)** (Backends → New):

| Field | Value |
|-------|-------|
| Name | `Ollama` |
| Type | `openai` |
| URL | `http://ollama:11434/v1` |
| API Key | *(leave empty)* |

**Parakeet (STT)** (Backends → New):

| Field | Value |
|-------|-------|
| Name | `Parakeet` |
| Type | `openai` |
| URL | `http://parakeet:5092` |
| API Key | *(leave empty)* |

**Edge TTS (TTS)** (Backends → New):

| Field | Value |
|-------|-------|
| Name | `Edge TTS` |
| Type | `openai` |
| URL | `http://tts:5050` |
| API Key | *(leave empty)* |

### Create memory providers

**Session memory** (Memory → New Session Provider):

| Field | Value |
|-------|-------|
| Type | `redis` |
| URL | `redis://redis:6379` |

**Long-term memory** (Memory → New Long-term Provider):

| Field | Value |
|-------|-------|
| Type | `pgvector` |
| URL | `postgres://magec:magec@postgres:5432/magec?sslmode=disable` |
| Embedding Backend | `Ollama` |
| Embedding Model | `nomic-embed-text` |

### Create an agent

Agents → New:

| Field | Value |
|-------|-------|
| Name | `Assistant` |
| System Prompt | Your agent's personality and instructions |
| LLM Backend | `Ollama` |
| LLM Model | `qwen3:8b` |

Expand the **Voice** section:

| Field | Value |
|-------|-------|
| Transcription Backend | `Parakeet` |
| Transcription Model | `nvidia/parakeet-ctc-0.6b-rnnt` |
| TTS Backend | `Edge TTS` |
| TTS Model | `tts-1` |
| TTS Voice | `es-ES-AlvaroNeural` (or any voice from the Edge TTS catalog) |

### Create a client

Clients → New:

| Field | Value |
|-------|-------|
| Name | `My Voice UI` |
| Type | `Voice UI` |
| Agent | `Assistant` |

Copy the **pairing token**, open `http://localhost:8080`, paste it, and start talking.

## Using cloud providers instead

The Docker Compose includes all the local AI services, but you can use cloud providers by simply creating different backends in the Admin UI. The local services will still be running but unused, or you can stop them.

### OpenAI (handles everything)

Create a single backend:

| Field | Value |
|-------|-------|
| Name | `OpenAI` |
| Type | `openai` |
| API Key | `sk-...` |
| URL | *(leave empty)* |

Use it for the agent's LLM (`gpt-4.1-mini`), transcription (`whisper-1`), and TTS (`tts-1`). For embeddings in long-term memory, use `text-embedding-3-small`.

Then stop the local services you don't need:

```bash
docker compose stop ollama ollama-setup parakeet tts
```

### Anthropic / Gemini (LLM only)

These providers only offer LLM. STT, TTS, and embeddings stay local. Create the cloud backend for the LLM and keep using Parakeet, Edge TTS, and Ollama (for embeddings) as configured above.

| Provider | Backend type | Model example |
|----------|-------------|---------------|
| Anthropic | `anthropic` | `claude-sonnet-4-20250514` |
| Gemini | `gemini` | `gemini-2.0-flash` |

## Managing the deployment

```bash
cd magec                               # your installation directory

docker compose logs -f                 # follow all logs
docker compose logs -f magec           # Magec server only
docker compose logs -f ollama-setup    # model download progress

docker compose down                    # stop everything
docker compose up -d                   # start again

docker compose pull                    # pull latest images
docker compose up -d                   # restart with new versions
```

## Data persistence

All data is stored in Docker volumes:

| Volume | Contains |
|--------|----------|
| `magec_data` | `store.json` (agents, backends, clients), `conversations.json` |
| `redis_data` | Session memory |
| `postgres_data` | Long-term memory (pgvector) |
| `ollama_data` | Downloaded AI models |

Your data survives `docker compose down/up`, image updates, and container recreation. To back up your Magec configuration, copy `data/store.json` from the `magec_data` volume.

To start completely fresh:

```bash
docker compose down -v    # -v removes all volumes
docker compose up -d      # fresh start
```

## Customizing your deployment

### Adding MCP servers as containers

You can extend the Docker Compose file to add MCP servers alongside Magec:

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

```yaml
services:
  magec:
    ports:
      - "3000:8080"   # Voice UI + User API on port 3000
      - "3001:8081"   # Admin UI + Admin API on port 3001
```

### Accessing host services

If your LLM or other services run on the Docker host (not in containers), use `host.docker.internal`:

```yaml
services:
  magec:
    extra_hosts:
      - "host.docker.internal:host-gateway"
```

{{< callout type="info" >}}
On macOS and Windows, `host.docker.internal` works automatically. On Linux, you need the `extra_hosts` mapping above.
{{< /callout >}}

## Next steps

- **[Configuration](/docs/configuration/).** Understand `config.yaml` vs. Admin UI resources
- **[Agents](/docs/agents/).** Customize agent behavior, prompts, and voice
- **[MCP Tools](/docs/mcp/).** Connect external tools (Home Assistant, GitHub, databases, etc.)
- **[Flows](/docs/flows/).** Chain agents into multi-step workflows
