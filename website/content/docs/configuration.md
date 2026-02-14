---
title: "Configuration"
---

Magec uses a YAML config file for server infrastructure. All resources (agents, backends, memory, tools, clients) are managed through the Admin UI and stored in `data/store.json`.

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

Config values support `${VAR}` syntax for environment variable expansion.

## Server block

| Field | Default | Description |
|-------|---------|-------------|
| `host` | `0.0.0.0` | Bind address |
| `port` | `8080` | User-facing port (Voice UI, user API, webhooks) |
| `adminPort` | `8081` | Admin-facing port (Admin UI, admin API) |

## Voice block

| Field | Default | Description |
|-------|---------|-------------|
| `ui.enabled` | `true` | Enable Voice UI and all voice routes. Set to `false` for API-only mode. |
| `onnxLibraryPath` | *auto* | Path to ONNX Runtime shared library. Only needed if auto-detection fails. |

## Log block

| Field | Default | Description |
|-------|---------|-------------|
| `level` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `format` | `console` | Output format: `console` (human), `json` (structured) |
