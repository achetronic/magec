---
title: "Getting Started"
---

Magec runs as a set of Docker containers. The install script downloads the right Docker Compose configuration for your setup, pulls the images, and starts everything. In a few minutes, you'll have a working AI platform with a voice interface, an admin panel, and your first agent ready to go.

**Requirements:** Docker and Docker Compose installed on your machine.

## Choose your deployment

Magec can run fully local (no cloud accounts, no API keys) or connect to cloud AI providers. The only difference is where the LLM, speech, and embedding models run — everything else (memory, tools, voice processing, the UIs) is always local.

### Fully local (default)

Everything runs on your machine. The LLM (Qwen 3 8B via Ollama), speech-to-text (Parakeet), text-to-speech (OpenAI Edge TTS), and embeddings (nomic-embed-text) are all local. No accounts, no API keys, no data leaving your network.

```bash
curl -fsSL https://raw.githubusercontent.com/achetronic/magec/main/scripts/install.sh | bash
```

{{< callout type="info" >}}
**First start** downloads approximately 5 GB of AI models. This only happens once. You can track the progress with `docker compose logs -f ollama-setup`.
{{< /callout >}}

### Cloud providers

If you prefer to use cloud models (faster, more capable for complex tasks), you can point Magec at your provider. Only Redis and PostgreSQL run locally — the AI inference happens in the cloud.

```bash
# OpenAI (GPT-4, Whisper, TTS)
export OPENAI_API_KEY=sk-...
curl -fsSL https://raw.githubusercontent.com/achetronic/magec/main/scripts/install.sh | bash -s -- --openai

# Anthropic (Claude) — STT and TTS still run locally
export ANTHROPIC_API_KEY=sk-ant-...
curl -fsSL https://raw.githubusercontent.com/achetronic/magec/main/scripts/install.sh | bash -s -- --anthropic

# Google Gemini — STT and TTS still run locally
export GEMINI_API_KEY=AI...
curl -fsSL https://raw.githubusercontent.com/achetronic/magec/main/scripts/install.sh | bash -s -- --gemini
```

{{< callout type="info" >}}
With Anthropic and Gemini, speech-to-text (Parakeet) and text-to-speech (OpenAI Edge TTS) still run locally because those providers don't offer voice APIs. Embeddings also use a local Ollama model. Only the LLM inference goes to the cloud.
{{< /callout >}}

### Install options

| Flag | Description |
|------|-------------|
| `--local` | Fully local deployment — no cloud APIs (default) |
| `--openai` | Use OpenAI for LLM, STT, TTS, and embeddings. Requires `OPENAI_API_KEY` |
| `--anthropic` | Use Anthropic for LLM. STT, TTS, and embeddings remain local. Requires `ANTHROPIC_API_KEY` |
| `--gemini` | Use Google Gemini for LLM. STT, TTS, and embeddings remain local. Requires `GEMINI_API_KEY` |
| `--gpu` | Enable NVIDIA GPU support for Ollama (local mode only). Requires nvidia-container-toolkit |
| `--dir NAME` | Installation directory (default: `magec`) |

## What the installer creates

The installer creates a directory (default: `magec/`) with the Docker Compose files and a `data/` directory where Magec stores its configuration. The structure looks like this:

```
magec/
├── docker-compose.yaml          # Base services
├── docker-compose.override.yaml # Provider-specific config
└── data/
    └── store.json               # Your agents, backends, clients, etc.
```

The `data/store.json` file is pre-populated with a working starter configuration (a "seed") that includes a backend, a memory provider, and a Voice UI client — so you can start talking to an agent immediately after install.

## Once running

Two UIs are available right away:

| URL | What it is |
|-----|-----------|
| **`http://localhost:8081`** | **Admin UI** — Create and manage agents, backends, memory providers, MCP tools, flows, clients, and commands. This is your control panel. |
| **`http://localhost:8080`** | **Voice UI** — Talk to your agents. Wake word ("Oye Magec"), push-to-talk, conversation history, agent switching. Installable as a phone app (PWA). |

The first time you open the Voice UI, it will ask for a pairing token. You can find this token in the Admin UI under **Clients** — the installer already created a Voice UI client for you.

## Your first conversation

1. Open the **Admin UI** at `http://localhost:8081`
2. Go to **Clients** and copy the token of the pre-created Voice UI client
3. Open the **Voice UI** at `http://localhost:8080`
4. Paste the token to pair
5. Tap the microphone or say **"Oye Magec"** and start talking

The default agent uses the LLM and voice configuration from your chosen deployment mode. From here, you can create new agents, change their prompts, add memory, connect MCP tools, build flows — all from the Admin UI.

## Managing your deployment

```bash
cd magec                        # or your --dir path

docker compose logs -f          # follow all logs
docker compose logs -f magec    # follow only the Magec server

docker compose down             # stop everything
docker compose up -d            # start again

docker compose pull             # update to latest images
docker compose up -d            # restart with new images
```

## What to do next

Now that Magec is running, here's the recommended path:

1. **[Configuration](/magec/docs/configuration/)** — Understand the two configuration layers (infrastructure vs. resources)
2. **[Agents](/magec/docs/agents/)** — Create your first custom agent with a specific personality and tools
3. **[AI Backends](/magec/docs/backends/)** — Add more AI providers or mix local and cloud models
4. **[MCP Tools](/magec/docs/mcp/)** — Connect external tools to make your agents actually useful (this is where it gets powerful)
5. **[Flows](/magec/docs/flows/)** — Chain agents into multi-step workflows
