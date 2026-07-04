---
title: "Configuration"
---

Magec uses two separate configuration layers. Understanding this split is important because they work differently and serve different purposes.

## The two layers

| Layer | What it controls | Where it lives | How you edit it |
|-------|-----------------|----------------|----------------|
| **Infrastructure** | Server ports, logging, voice toggle, ONNX paths | `config.yaml` | Text editor or environment variables |
| **Resources** | Agents, backends, memory, MCP tools, clients, commands, flows | `data/store.json` | Admin UI or Admin API (never by hand) |

**Infrastructure** is about how the server runs. **Resources** are about what the server does.

Think of it like this: `config.yaml` is the wiring of your house (ports, protocols, switches). `store.json` is the furniture and appliances inside (your agents, their tools, who can talk to them).

## config.yaml — Infrastructure

This file is read once at startup. Changes require a server restart. It controls only the server process itself — nothing about agents, backends, or AI.

```yaml
server:
  host: 0.0.0.0
  port: 8080          # User API + Voice UI
  adminPort: 8081     # Admin API + Admin UI
  # adminPassword: ${MAGEC_ADMIN_PASSWORD}
  # encryptionKey: ${MAGEC_ENCRYPTION_KEY}

voice:
  ui:
    enabled: true     # Toggle Voice UI and voice routes
  # onnxLibraryPath: /usr/lib/libonnxruntime.so

log:
  level: info         # debug, info, warn, error
  format: console     # console, json
```

### Server

Controls where the server listens. Magec runs two HTTP servers on separate ports — one for users (API, Voice UI, webhooks) and one for administration (Admin UI, management API).

| Field | Default | Description |
|-------|---------|-------------|
| `host` | `0.0.0.0` | Bind address. Use `127.0.0.1` to restrict to localhost only. |
| `port` | `8080` | User-facing port. The Voice UI, user API, webhook endpoints, and voice WebSocket all live here. |
| `adminPort` | `8081` | Admin-facing port. The Admin UI and management API live here. In production, restrict access to this port. |
| `adminPassword` | *(empty)* | Password for admin API authentication. When set, all `/api/` requests require `Authorization: Bearer <password>`. See [Admin Password](/docs/admin-password/). |
| `encryptionKey` | *(empty)* | Key used to encrypt [Secrets](/docs/secrets/) at rest (AES-256-GCM, PBKDF2-derived). When not set, secrets are stored in plain text with a warning. Independent from `adminPassword`. |

### Voice

Controls voice features at the server level. This is independent of per-agent voice settings — it's a global switch.

| Field | Default | Description |
|-------|---------|-------------|
| `ui.enabled` | `true` | When `true`, the Voice UI is served at the user port and all voice routes (STT proxy, TTS proxy, WebSocket) are active. Set to `false` for API-only deployments where you don't need voice. |
| `onnxLibraryPath` | *auto-detect* | Path to the ONNX Runtime shared library (`.so` / `.dylib`). Magec tries to find it automatically. Only set this if auto-detection fails. |

### Log

| Field | Default | Description |
|-------|---------|-------------|
| `level` | `info` | Minimum log level. Use `debug` during development to see detailed agent execution, MCP tool calls, and voice events. |
| `format` | `console` | `console` produces human-readable colored output. `json` produces structured logs for log aggregation systems. |

### Environment variables

All values in `config.yaml` support `${VAR}` syntax for environment variable substitution. This is useful for Docker deployments where you want to configure the server without modifying the file:

```yaml
server:
  port: ${MAGEC_PORT:-8080}

log:
  level: ${LOG_LEVEL:-info}
```

## data/store.json — Resources

This is Magec's internal database. Every agent, backend, memory provider, MCP server, client, command, and flow is stored here. The server keeps it in memory for fast access and writes to disk on every change.

**You should never edit this file by hand.** Use the Admin UI at `http://localhost:8081` or the Admin REST API. Both provide full CRUD for all resource types with validation, health checks, and immediate hot-reload.

### What's inside

| Resource | What it is | More info |
|----------|-----------|-----------|
| **Backends** | Connections to AI providers (OpenAI, Anthropic, Gemini, Ollama) | [AI Backends](/docs/backends/) |
| **Memory Providers** | Session storage (Redis) and long-term memory (PostgreSQL + pgvector) | [Memory](/docs/memory/) |
| **MCP Servers** | External tool connections via Model Context Protocol | [MCP Tools](/docs/mcp/) |
| **Agents** | AI entities with their own LLM, prompt, memory, voice, and tools | [Agents](/docs/agents/) |
| **Commands** | Reusable prompts referenced by cron and webhook clients | [Commands](/docs/commands/) |
| **Clients** | Access points — Voice UI, Telegram, webhooks, cron — each with its own token | [Clients](/docs/clients/) |
| **Flows** | Multi-agent workflows as directed graphs: routers, joins, transforms, loops | [Flows](/docs/flows/) |
| **Secrets** | Encrypted key-value pairs for API keys, tokens, and passwords | [Secrets](/docs/secrets/) |

### Hot-reload

Changes made through the Admin UI or API take effect **immediately** — no server restart needed. When you update an agent's prompt, add a new backend, or modify a flow, the change is persisted to `store.json` and the server picks it up in real time.

### Environment variables in store.json

Like `config.yaml`, the store also supports `${VAR}` substitution. This is particularly useful for keeping API keys and tokens out of the file:

```json
{
  "apiKey": "${OPENAI_API_KEY}"
}
```

The installer uses this for cloud deployments, so your API keys come from environment variables rather than being stored in plain text.

For a more convenient way to manage sensitive values, see [Secrets](/docs/secrets/) — they let you store API keys through the Admin UI and reference them with the same `${VAR}` syntax.

### Backup and restore

To back up your entire Magec configuration, copy `data/store.json`. To restore it, put it back and restart the server. That single file contains every resource you've configured.

{{< callout type="info" >}}
If you're running Magec via Docker Compose, the `data/` directory is mounted as a volume. Your configuration persists across container restarts and image updates.
{{< /callout >}}
