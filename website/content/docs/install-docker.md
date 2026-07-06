---
title: "Docker Quick Start (OpenAI)"
---

The fastest way to try Magec. One command, one API key, no local models. OpenAI handles everything: LLM, speech-to-text (Whisper), text-to-speech, and embeddings.

**Requirements:**
- Docker installed
- An OpenAI API key (`sk-...`)

## 1. Create a config file

Create a `config.yaml` with the default server settings:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  adminPort: 8081

voice:
  ui:
    enabled: true

log:
  level: info
  format: console
```

## 2. Run Magec

```bash
docker run -d \
  --name magec \
  -p 8080:8080 \
  -p 8081:8081 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v magec_data:/app/data \
  ghcr.io/achetronic/magec:latest
```

That's it. Magec is now running:

| URL | What it is |
|-----|-----------|
| **`http://localhost:8081`** | **Admin UI.** Create agents, backends, clients, and everything else |
| **`http://localhost:8080`** | **Voice UI.** Talk to your agents (wake word + push-to-talk) |

## 3. Set up your first agent

Open the **Admin UI** at `http://localhost:8081` and create the resources in this order:

### Create a backend

Go to **Backends → New**:

| Field | Value |
|-------|-------|
| Name | `OpenAI` |
| Type | `openai` |
| API Key | Your OpenAI API key (`sk-...`) |
| URL | *(leave empty: uses the default OpenAI endpoint)* |

### Create an agent

Go to **Agents → New**:

| Field | Value |
|-------|-------|
| Name | `Assistant` |
| System Prompt | Write whatever personality/instructions you want |
| LLM Backend | Select `OpenAI` |
| LLM Model | `gpt-4.1-mini` (or `gpt-4.1`, `gpt-4o`, etc.) |

To enable voice, expand the **Voice** section:

| Field | Value |
|-------|-------|
| Transcription Backend | `OpenAI` |
| Transcription Model | `whisper-1` |
| TTS Backend | `OpenAI` |
| TTS Model | `tts-1` |
| TTS Voice | `nova` (or `alloy`, `shimmer`, `echo`, `onyx`, `fable`) |

### Create a client

Go to **Clients → New**:

| Field | Value |
|-------|-------|
| Name | `My Voice UI` |
| Type | `Voice UI` |
| Agent | Select `Assistant` |

Save. Copy the **pairing token** that appears.

### Connect the Voice UI

1. Open `http://localhost:8080`
2. Paste the pairing token
3. Tap the microphone or say **"Oye Magec"** and start talking

## Optional: Add memory

To give your agent conversation memory, you need Redis (for session history) and/or PostgreSQL with pgvector (for long-term semantic memory).

The quickest way to add Redis:

```bash
docker run -d --name magec-redis -p 6379:6379 redis:alpine
```

Then in the Admin UI, go to **Memory → New Session Provider**:

| Field | Value |
|-------|-------|
| Type | `redis` |
| URL | `redis://host.docker.internal:6379` |

{{< callout type="info" >}}
On Linux, you need `--add-host=host.docker.internal:host-gateway` in your `docker run` command for the Magec container to reach services on the host. On macOS and Windows, `host.docker.internal` works automatically.
{{< /callout >}}

For the full memory setup (including long-term memory with PostgreSQL + pgvector), see the [Docker Compose: Local](/docs/install-compose-local/) guide which includes everything pre-configured.

## Data persistence

The `-v magec_data:/app/data` flag creates a Docker volume for Magec's data. Your agents, backends, clients, and conversation history persist across container restarts and updates.

To update Magec:

```bash
docker pull ghcr.io/achetronic/magec:latest
docker rm -f magec
# Run the same docker run command again
```

Your data is safe in the `magec_data` volume.

## Next steps

This setup is great for trying Magec, but for a full deployment with memory, local models, or multiple services, check:

- **[Docker Compose: Local](/docs/install-compose-local/).** Fully local with Ollama, Parakeet, memory, everything pre-wired
- **[Docker Compose: Cloud](/docs/install-compose-cloud/).** Cloud providers with full infrastructure
- **[Configuration](/docs/configuration/).** Understand the two configuration layers
- **[MCP Tools](/docs/mcp/).** Connect external tools to your agents
