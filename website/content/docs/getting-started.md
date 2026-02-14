---
title: "Getting Started"
---

## Installation

The install script downloads the Docker Compose files, picks the right configuration for your chosen mode, and starts everything with a single command.

**Requirements:** Docker and Docker Compose.

### Fully local (default)

No API keys, no sign-ups. LLM, speech-to-text, text-to-speech, and embeddings all run on your machine.

```bash
curl -fsSL https://raw.githubusercontent.com/achetronic/magec/main/scripts/install.sh | bash
```

{{< callout type="info" >}}
**First start** downloads ~5GB of models (Ollama qwen3:8b + nomic-embed-text). Track progress with `docker compose logs -f ollama-setup`.
{{< /callout >}}

### Cloud APIs

Only Redis and PostgreSQL run locally. LLM, STT, TTS, and embeddings use your cloud provider.

```bash
# OpenAI
export OPENAI_API_KEY=sk-...
curl -fsSL .../install.sh | bash -s -- --openai

# Anthropic
export ANTHROPIC_API_KEY=sk-ant-...
curl -fsSL .../install.sh | bash -s -- --anthropic

# Google Gemini
export GEMINI_API_KEY=AI...
curl -fsSL .../install.sh | bash -s -- --gemini
```

### Options

| Flag | Description |
|------|-------------|
| `--local` | Fully local deployment (default) |
| `--openai` | Use OpenAI APIs. Requires `OPENAI_API_KEY` |
| `--anthropic` | Use Anthropic APIs. Requires `ANTHROPIC_API_KEY` |
| `--gemini` | Use Google Gemini APIs. Requires `GEMINI_API_KEY` |
| `--gpu` | Enable NVIDIA GPU support for Ollama (local mode only) |
| `--dir NAME` | Installation directory (default: `magec`) |

### Once running

- **Admin UI** → `http://localhost:8081` — configure your agents, backends, memory, and clients
- **Voice UI** → `http://localhost:8080` — start chatting

### Managing the deployment

```bash
cd magec                        # or your --dir path
docker compose logs -f          # follow logs
docker compose down             # stop everything
docker compose up -d            # start again
```
