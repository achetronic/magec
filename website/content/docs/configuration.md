---
title: "Configuration"
---

Magec splits configuration into two separate layers:

| Layer | File | Purpose | Managed by |
|-------|------|---------|------------|
| **Infrastructure** | `config.yaml` | Server ports, logging, voice/ONNX settings | You (text editor or env vars) |
| **Resources** | `data/store.json` | Agents, backends, memory, MCP servers, clients, commands, flows | Admin UI or Admin API |

`config.yaml` is read once at startup. `data/store.json` is read/written at runtime — changes made through the Admin UI take effect immediately without restarting the server.

## config.yaml — Infrastructure

This file controls **only** how the server itself runs. It has nothing to do with agents, backends, or any AI resource.

```yaml
server:
  host: 0.0.0.0
  port: 8080          # Voice UI + User API
  adminPort: 8081     # Admin UI + Admin API

voice:
  ui:
    enabled: true     # Toggle Voice UI and voice routes (default: true)
  # onnxLibraryPath: /usr/lib/libonnxruntime.so

log:
  level: info         # debug, info, warn, error
  format: console     # console, json
```

Values support `${VAR}` syntax for environment variable expansion.

### Server block

| Field | Default | Description |
|-------|---------|-------------|
| `host` | `0.0.0.0` | Bind address |
| `port` | `8080` | User-facing port (Voice UI, user API, webhooks) |
| `adminPort` | `8081` | Admin-facing port (Admin UI, admin API) |

### Voice block

| Field | Default | Description |
|-------|---------|-------------|
| `ui.enabled` | `true` | Enable Voice UI and all voice routes. Set to `false` for API-only mode. |
| `onnxLibraryPath` | *auto* | Path to ONNX Runtime shared library. Only needed if auto-detection fails. |

### Log block

| Field | Default | Description |
|-------|---------|-------------|
| `level` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `format` | `console` | Output format: `console` (human), `json` (structured) |

## data/store.json — Resources

Magec uses `data/store.json` as its internal database. Every resource you create, edit, or delete through the Admin UI is persisted to this file automatically. The server keeps it in memory for fast access and writes it to disk on every change — no external database is needed for the store itself.

You should never need to edit `store.json` by hand. All resources are managed at runtime through the **Admin UI** at `http://localhost:8081` or the **Admin REST API**.

| Resource | Description |
|----------|-------------|
| **Backends** | AI provider connections (OpenAI, Anthropic, Gemini, Ollama) |
| **Memory** | Session storage (Redis) and long-term memory (PostgreSQL + pgvector) |
| **MCP Servers** | External tool servers via Model Context Protocol |
| **Agents** | Independent units with their own LLM, memory, voice, tools, and prompts |
| **Commands** | Reusable prompts referenced by cron and webhook clients |
| **Clients** | Access points (Voice UI, Telegram, webhooks, cron) with token-based auth |
| **Flows** | Multi-agent workflows (sequential, parallel, loop) |

On first run, the store starts empty. You configure everything from the Admin UI.

{{< callout type="info" >}}
`store.json` supports `${VAR}` expansion for credentials, so you can use environment variables instead of hardcoding API keys or tokens.
{{< /callout >}}

{{< callout type="info" >}}
To back up your Magec configuration, just copy `data/store.json`. To restore it, put it back and restart the server.
{{< /callout >}}
