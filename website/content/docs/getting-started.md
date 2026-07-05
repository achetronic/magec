---
title: "Getting Started"
---

Magec is a self-contained binary: the server, Admin UI, Voice UI, wake word models, and everything else are baked into a single executable. You just need to give it a `config.yaml` and point it at the services it needs (an LLM, optionally Redis/PostgreSQL for memory, etc.).

There are three ways to get Magec running depending on your needs:

## Installation guides

### Quick start: Docker with OpenAI

**Best for:** Trying Magec for the first time with minimal setup.

A single `docker run` command with your OpenAI API key. No local models, no compose files, no GPU required. You get LLM, STT, TTS, and embeddings from OpenAI out of the box.

→ **[Docker Quick Start (OpenAI)](/docs/install-docker/)**

---

### Docker Compose: fully local

**Best for:** Privacy-first deployments where nothing leaves your network.

Everything runs on your machine: LLM (Ollama), speech-to-text (Parakeet), text-to-speech (OpenAI Edge TTS), embeddings (nomic-embed-text), memory (Redis + PostgreSQL). No API keys, no cloud accounts. You can also swap in cloud providers later by just changing backends in the Admin UI.

→ **[Docker Compose](/docs/install-compose-local/)**

---

### Binary: manual install

**Best for:** Running Magec directly on your machine without Docker. Ideal for local MCP tools (filesystem, git, shell), development, or when you want full control over the process.

Download the binary for your platform, create a `config.yaml`, and run it. You manage the dependencies (LLM, ffmpeg, ONNX Runtime) yourself.

→ **[Binary Installation](/docs/install-binary/)**

---

## What's the same everywhere

Regardless of how you install Magec, you always get:

- **Admin UI** at `http://localhost:8081`: create and manage agents, backends, memory, MCP tools, flows, clients
- **Voice UI** at `http://localhost:8080`: talk to your agents with wake word or push-to-talk
- **User API** at `http://localhost:8080`: webhook and programmatic access
- **Hot-reload:** changes through the Admin UI take effect immediately, no restart needed

## After installation

Once Magec is running, the first thing you need to do is **create a client** in the Admin UI. Clients are access tokens that connect users to agents: the Voice UI needs a pairing token, Telegram needs a bot token, webhooks need an API key.

1. Open the **Admin UI** at `http://localhost:8081`
2. Create a **backend** (your AI provider connection)
3. Create an **agent** (assign it the backend, a system prompt, and optionally voice/memory/tools)
4. Create a **client** (Voice UI, Telegram, webhook, or cron) and assign it to the agent
5. Use the client to talk to your agent

The recommended path through the docs after installation:

1. **[Configuration](/docs/configuration/).** Understand `config.yaml` (infrastructure) vs the Admin UI (resources)
2. **[Agents](/docs/agents/).** Create your first custom agent
3. **[AI Backends](/docs/backends/).** Add and mix AI providers
4. **[MCP Tools](/docs/mcp/).** Connect external tools
5. **[Flows](/docs/flows/).** Chain agents into multi-step workflows
