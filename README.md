# Magec

<p align="center">
  <img src="docs/img/banner.png" alt="Magec" width="800">
</p>

<p align="center">
  <strong>Self-hosted multi-agent AI platform with voice, visual workflows, and tool integration.</strong>
</p>

<p align="center">
  <a href="https://magec.dev">Website</a> ·
  <a href="https://magec.dev/docs/">Docs</a> ·
  <a href="#quick-start">Quick Start</a>
</p>

---

Define multiple AI agents, each with its own LLM, memory, and tools. Chain them into multi-step workflows. Access via voice, Telegram, webhooks, or cron. Manage it all from a visual admin panel.

Your server, your data, your rules.

<p align="center">
  <img src="docs/img/architecture.svg" alt="Architecture" width="860">
</p>

## Quick Start

### One-line install (fully local, no API keys)

```bash
curl -fsSL https://raw.githubusercontent.com/achetronic/magec/master/scripts/install.sh | bash
```

Downloads a Docker Compose file with everything: LLM (Ollama), STT (Parakeet), TTS (Edge TTS), embeddings, Redis, PostgreSQL. Add `--gpu` for NVIDIA acceleration.

### Docker with OpenAI (minimal)

```bash
docker run -d --name magec \
  -p 8080:8080 -p 8081:8081 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v magec_data:/app/data \
  ghcr.io/achetronic/magec:latest
```

Create backends, agents, and clients from the Admin UI. See the [Docker Quick Start](https://achetronic.github.io/magec/docs/install-docker/) guide.

### Binary

Download from [Releases](https://github.com/achetronic/magec/releases), extract, and run:

```bash
./magec --config config.yaml
```

Ideal for local MCP tools (filesystem, git, shell). See the [Binary Installation](https://achetronic.github.io/magec/docs/install-binary/) guide.

---

**Admin UI** → http://localhost:8081 · **Voice UI** → http://localhost:8080

## Highlights

- **Multi-agent**: per-agent LLM, memory, voice, and tools. Hot-reload from the Admin UI.
- **Agentic Flows**: visual graph editor. Agents, routers, joins, loops, parallel fan-out, and code blocks (CEL and Starlark).
- **Run auditing**: every execution recorded and browsable as a timeline; conversations projected from the same data.
- **Secrets**: encrypted at rest, usable by agents and flows through placeholders the models never see resolved.
- **Any backend**: OpenAI, Anthropic, Gemini, Ollama, or any OpenAI-compatible API.
- **MCP tools**: Home Assistant, GitHub, databases, and hundreds more via Model Context Protocol.
- **Memory**: session (Redis) + long-term semantic (PostgreSQL/pgvector).
- **Voice**: wake word, VAD, STT, TTS. All server-side via ONNX Runtime. Privacy-first.
- **Clients**: Voice UI (PWA), Telegram, Discord, Slack, webhooks, cron, REST API.
- **A2A**: expose agents and flows to other agent platforms via the Agent-to-Agent protocol.

## Screenshots

See all screenshots in the [documentation](https://achetronic.github.io/magec/docs/screenshots/).

## Roadmap

- [x] Multi-agent system with per-agent LLM, memory, and tools
- [x] Visual flow editor
- [x] MCP tool integration (HTTP + stdio transports)
- [x] Voice UI with wake word detection and VAD
- [x] Telegram client with voice support
- [x] Long-term semantic memory (pgvector)
- [x] Session memory (Redis)
- [x] Webhook and cron clients
- [x] Admin UI with hot-reload
- [x] Secrets management (encrypted storage for API keys and sensitive credentials)
- [x] Slack client
- [x] Context window management: automatic summarization when approaching token limits (experimental)
- [x] Expose agents and flows with A2a (Agent-to-agent) protocol
- [x] Discord client
- [x] Flows as real graphs (routers, joins, loops, fan-out, CEL/Starlark blocks)
- [x] Run auditing with per-block timelines
- [x] Secrets usable by agents and flows without exposing values to the models

## Documentation

Full docs at **[magec.dev/docs](https://magec.dev/docs/)**: installation, configuration, agents, flows, backends, memory, MCP tools, clients, voice system, and API reference.

## Development

### Requirements

- Go 1.25+
- Node.js 22+ (for UI builds)
- Docker (for infrastructure services)

### Make commands

| Command              | Description                                               |
| -------------------- | --------------------------------------------------------- |
| `make build`         | Build frontend UIs + embed models + compile server binary |
| `make dev`           | Build all and start server                                |
| `make dev-admin`     | Start Admin UI dev server (Vite, hot-reload)              |
| `make dev-voice`     | Start Voice UI dev server (Vite, hot-reload)              |
| `make swagger`       | Regenerate Swagger docs                                   |
| `make infra`         | Start PostgreSQL + Redis                                  |
| `make ollama`        | Start Ollama with qwen3:8b + nomic-embed-text             |
| `make docker-build`  | Build Docker image (current arch)                         |
| `make docker-buildx` | Build multi-arch image (amd64 + arm64)                    |
| `make clean`         | Remove build artifacts                                    |

### Key dependencies

| Dependency                                                                    | Purpose                          |
| ----------------------------------------------------------------------------- | -------------------------------- |
| [google.golang.org/adk](https://pkg.go.dev/google.golang.org/adk)             | Google Agent Development Kit     |
| [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) | MCP client                       |
| [yalue/onnxruntime_go](https://github.com/yalue/onnxruntime_go)               | ONNX Runtime for wake word / VAD |
| [mymmrac/telego](https://github.com/mymmrac/telego)                           | Telegram bot                     |
| [achetronic/adk-utils-go](https://github.com/achetronic/adk-utils-go)         | ADK providers, session, memory   |

## Special Mentions

| Who                                      | What                                                                                                                                                                                                                                                                                                                                             |
| ---------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [@travisvn](https://github.com/travisvn) | Built the ARM64 Docker image for [OpenAI Edge TTS](https://github.com/travisvn/openai-edge-tts) in record time. This is the local TTS service we recommend: it exposes an OpenAI-compatible API (`/v1/audio/speech`) that uses Microsoft Edge's free neural voices under the hood, so Magec can use it as a drop-in replacement for OpenAI TTS. |

## Contributors

<a href="https://github.com/achetronic/magec/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=achetronic/magec" />
</a>

## License

[Apache 2.0](LICENSE) · Alby Hernández

---

<p align="center">
  If you find Magec useful, please ⭐ star this repo, it helps a lot.
</p>
