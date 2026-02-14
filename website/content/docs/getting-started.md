---
title: "Getting Started"
---

## Installation

### Option A: Fully local (no API keys)

Everything runs on your machine — LLM, speech-to-text, text-to-speech, embeddings.

```bash
git clone https://github.com/achetronic/magec.git
cd magec/docker/compose/fully-local
docker compose up -d
```

{{< callout type="info" >}}
**First start** downloads ~5GB of models (Ollama qwen3:8b + nomic-embed-text). For NVIDIA GPU support, uncomment the `deploy` section in `docker-compose.yaml`.
{{< /callout >}}

### Option B: Cloud APIs

Only Redis and PostgreSQL run locally. LLM, STT, TTS, and embeddings use cloud providers.

```bash
git clone https://github.com/achetronic/magec.git
cd magec/docker/compose/remote-openai
export OPENAI_API_KEY=sk-...
docker compose up -d
```

### Option C: From source

```bash
make infra           # Start PostgreSQL + Redis
cp config.example.yaml config.yaml
make dev             # Build and run (includes Admin UI)
```

Once running:

- **Admin UI** → `http://localhost:8081` — configure your agents
- **Voice UI** → `http://localhost:8080` — start chatting
