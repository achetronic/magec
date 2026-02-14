---
title: "Development"
---

## Requirements

- Go 1.21+
- Node.js (for Admin UI build)
- Docker (for infrastructure services)

## Make commands

| Command | Description |
|---------|-------------|
| `make build` | Build Admin UI + server binary |
| `make dev` | Build all and start server |
| `make dev-admin` | Start Admin UI dev server (Vite, hot-reload) |
| `make swagger` | Regenerate Swagger docs |
| `make infra` | Start PostgreSQL + Redis |
| `make ollama` | Start Ollama with qwen3:8b + nomic-embed-text |
| `make clean` | Remove build artifacts |

## Project structure

```
magec/
├── server/                    # Go backend
│   ├── main.go                # Dual HTTP server (:8080 + :8081)
│   ├── agent/                 # Google ADK multi-agent + flow execution
│   ├── api/admin/             # Admin REST API (35+ endpoints)
│   ├── api/user/              # User API (health, client info, speech)
│   ├── clients/               # Client types (direct, telegram, webhook, cron)
│   ├── config/                # YAML config parsing
│   ├── memory/                # Memory providers (Redis, PostgreSQL/pgvector)
│   ├── middleware/             # Auth, CORS, access logging
│   ├── store/                 # JSON persistence with env expansion
│   └── voice/                 # Wake word + VAD (ONNX Runtime)
├── admin-ui/                  # Vue 3 + Vite admin panel
├── voice-ui/                  # Vanilla JS voice interface (PWA)
├── docker/
│   ├── build/                 # Dockerfile + entrypoint
│   └── compose/               # Docker Compose deployments
├── data/seeds/                # Store presets (voice-ui, examples)
├── models/                    # Wake word ONNX models
└── config.example.yaml
```

## Key dependencies

| Dependency | Purpose |
|------------|---------|
| [google.golang.org/adk](https://pkg.go.dev/google.golang.org/adk) | Google Agent Development Kit |
| [achetronic/adk-utils-go](https://github.com/achetronic/adk-utils-go) | ADK providers, session, memory |
| [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) | MCP client |
| [yalue/onnxruntime_go](https://github.com/yalue/onnxruntime_go) | ONNX Runtime for wake word/VAD |
| [mymmrac/telego](https://github.com/mymmrac/telego) | Telegram bot |
