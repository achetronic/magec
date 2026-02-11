# Multi-Agent Admin API

## Objetivo

Transformar Magec de un sistema mono-agente (configurado vía YAML estático) a un sistema **multi-agente** gestionable en runtime a través de una API de administración.

## Contexto

Actualmente toda la configuración vive en un `config.yaml` que se lee al arrancar:
- Un único agente con un prompt fijo
- Un único bot de Telegram
- MCPs globales compartidos
- Backends globales
- Sesiones sin visibilidad desde fuera

## Qué se pide

1. **Múltiples agentes**, cada uno con:
   - Nombre e ID único
   - System prompt propio
   - Backend LLM propio (referencia a backends globales)
   - MCPs conectados individualmente
   - Configuración de Telegram (token, allowedUsers, responseMode) opcional por agente
   - Configuración de TTS y transcripción propia (o heredada de defaults)

2. **API de administración** (`/api/v1/admin/`) con endpoints CRUD para:
   - **Agentes**: crear, listar, obtener, actualizar, eliminar
   - **Backends**: crear, listar, obtener, actualizar, eliminar
   - **Memory Providers**: crear, listar, obtener, actualizar, eliminar, health check, tipos
   - **MCP Servers**: crear, listar, obtener, actualizar, eliminar (asociados a agente)
   - **Devices**: crear, listar, obtener, actualizar, eliminar, regenerar token
   - **Cron Jobs**: crear, listar, obtener, actualizar, eliminar
   - **Estado global**: health, config overview

3. **Implementado**:
   - Selector de agentes en voice-ui (dropdown con `setAgent()` propagado a todos los componentes)
   - Hot-reload de agentes (crear/modificar sin reiniciar, via `OnChange()` + debounce 500ms)
   - Rename con cascade (renombrar recurso actualiza todas las referencias automáticamente)

4. **Futuro (fuera de scope ahora)**:
   - Autenticación en la API de admin

## Diseño técnico

### Modelo de datos

```
Store (in-memory + JSON persistence)
├── Backends[]          — pools de backends reutilizables (globales)
├── MemoryProviders[]   — proveedores de memoria (globales, extensibles)
│   ├── name, type, category (session | longterm)
│   ├── config: {connectionString, ...extras por tipo}
│   └── embedding: *BackendRef (solo longterm, nil para session)
├── MCPServers[]        — servidores MCP reutilizables (globales)
├── Agents[]            — cada agente es una unidad independiente
│   ├── id, name, description
│   ├── systemPrompt, systemPromptSuffix
│   ├── llm: {backend, model}
│   ├── transcription: {backend, model}
│   ├── tts: {backend, model, voice, speed}
│   ├── memory: {session: "provider-name", longTerm: "provider-name"}
│   ├── mcpServers: ["name1", "name2"]  — referencias a MCPs globales
│   └── telegram: {enabled, token, allowedUsers, allowedChats, responseMode}
├── CronJobs[]          — tareas programadas
├── Devices[]           — puntos de acceso (voice-ui)
└── Server config       — host, port, log level (solo lectura desde YAML)
```

### Jerarquía de recursos

```
Backends (AI infra) → Memory (data infra) → MCPs (tools) → Agents (consumers) → Devices (access) → Crons (automation)
```

## API Reference

Base path: `/api/v1/admin`

All endpoints return JSON. Error responses use `{"error": "message"}`.

---

### Overview

| Method | Path | Description | Response |
|--------|------|-------------|----------|
| GET | `/overview` | Dashboard summary | `{agents, backends, mcps, memoryProviders, devices, cronJobs}` counts + agent summaries |

---

### Backends

CRUD for reusable AI backends (OpenAI, Ollama, Anthropic, Gemini).

| Method | Path | Description | Request Body | Response |
|--------|------|-------------|-------------|----------|
| GET | `/backends` | List all | — | `BackendDefinition[]` |
| POST | `/backends` | Create | `BackendDefinition` | `201` BackendDefinition |
| GET | `/backends/{name}` | Get by name | — | `BackendDefinition` |
| PUT | `/backends/{name}` | Update | `BackendDefinition` | `BackendDefinition` |
| DELETE | `/backends/{name}` | Delete | — | `204` |

**BackendDefinition:**
```json
{
  "name": "ollama",
  "type": "openai",
  "url": "http://localhost:11434/v1",
  "apiKey": ""
}
```

Types: `openai`, `anthropic`, `gemini`

---

### Memory Providers

CRUD for memory infrastructure. Schema-driven — field specs declared by each provider.

| Method | Path | Description | Request Body | Response |
|--------|------|-------------|-------------|----------|
| GET | `/memory` | List all | — | `MemoryProvider[]` |
| POST | `/memory` | Create | `MemoryProvider` | `201` MemoryProvider |
| GET | `/memory/types` | Registered provider types | — | `TypeInfo[]` (with field specs) |
| GET | `/memory/{name}` | Get by name | — | `MemoryProvider` |
| PUT | `/memory/{name}` | Update | `MemoryProvider` | `MemoryProvider` |
| DELETE | `/memory/{name}` | Delete | — | `204` |
| GET | `/memory/{name}/health` | Real-time ping (5s timeout) | — | `HealthResult` |

**MemoryProvider:**
```json
{
  "name": "redis-session",
  "type": "redis",
  "category": "session",
  "config": {
    "connectionString": "redis://localhost:6379/0",
    "ttl": "24h"
  }
}
```
```json
{
  "name": "postgres-longterm",
  "type": "postgres",
  "category": "longterm",
  "config": {
    "connectionString": "postgres://user:pass@localhost:5432/magec?sslmode=disable"
  },
  "embedding": {
    "backend": "ollama",
    "model": "nomic-embed-text"
  }
}
```

**TypeInfo** (from `/memory/types`):
```json
[
  {
    "type": "redis",
    "displayName": "Redis",
    "categories": ["session"],
    "fields": [
      {"key": "connectionString", "label": "Connection String", "type": "text", "required": true, "placeholder": "redis://localhost:6379/0"},
      {"key": "ttl", "label": "TTL", "type": "text", "placeholder": "24h", "default": "24h"}
    ]
  },
  {
    "type": "postgres",
    "displayName": "PostgreSQL",
    "categories": ["longterm"],
    "fields": [
      {"key": "connectionString", "label": "Connection String", "type": "text", "required": true, "placeholder": "postgres://user:pass@localhost:5432/db?sslmode=disable"}
    ]
  }
]
```

**HealthResult:**
```json
{"healthy": true, "detail": "connected"}
```

**Validation:**
- `name`, `type`, `category` required on create
- `type` must be registered in the provider registry
- `type` must support the given `category` (checked via `ValidTypeForCategory`)

---

### MCP Servers

CRUD for global MCP (Model Context Protocol) tool servers.

| Method | Path | Description | Request Body | Response |
|--------|------|-------------|-------------|----------|
| GET | `/mcps` | List all | — | `MCPServer[]` |
| POST | `/mcps` | Create | `MCPServer` | `201` MCPServer |
| GET | `/mcps/{name}` | Get by name | — | `MCPServer` |
| PUT | `/mcps/{name}` | Update | `MCPServer` | `MCPServer` |
| DELETE | `/mcps/{name}` | Delete | — | `204` |

**MCPServer:**
```json
{
  "name": "home-assistant",
  "type": "http",
  "endpoint": "http://localhost:8070/mcp",
  "headers": {"Authorization": "Bearer token..."},
  "systemPrompt": "Home automation tools"
}
```
```json
{
  "name": "local-tool",
  "type": "stdio",
  "command": "/path/to/tool",
  "args": ["--flag"],
  "env": {"KEY": "value"},
  "workDir": "/path/to/dir"
}
```

Types: `http` (StreamableClientTransport), `stdio` (CommandTransport)

---

### Agents

CRUD for agent definitions. Each agent is an independent unit with its own LLM, memory, tools, and clients.

| Method | Path | Description | Request Body | Response |
|--------|------|-------------|-------------|----------|
| GET | `/agents` | List all | — | `AgentDefinition[]` |
| POST | `/agents` | Create | `AgentDefinition` | `201` AgentDefinition |
| GET | `/agents/{id}` | Get by ID | — | `AgentDefinition` |
| PUT | `/agents/{id}` | Update | `AgentDefinition` | `AgentDefinition` |
| DELETE | `/agents/{id}` | Delete | — | `204` |
| GET | `/agents/{id}/mcps` | List resolved MCPs | — | `MCPServer[]` |
| PUT | `/agents/{id}/mcps/{name}` | Link MCP to agent | — | `200` |
| DELETE | `/agents/{id}/mcps/{name}` | Unlink MCP from agent | — | `200` |

**AgentDefinition:**
```json
{
  "id": "default",
  "name": "Magec",
  "description": "Default voice assistant agent",
  "systemPrompt": "You are a helpful assistant...",
  "systemPromptSuffix": "Additional instructions...",
  "llm": {"backend": "ollama", "model": "qwen3:8b"},
  "transcription": {"backend": "parakeet", "model": "whisper-1"},
  "tts": {"backend": "openai", "model": "tts-1", "voice": "alloy", "speed": 1.0},
  "memory": {"session": "redis-session", "longTerm": "postgres-longterm"},
  "mcpServers": ["home-assistant", "local-tool"],
  "telegram": {
    "enabled": true,
    "token": "bot-token",
    "allowedUsers": [123456789],
    "allowedChats": [-100123456789],
    "responseMode": "both"
  }
}
```

---

### Devices

CRUD for voice-UI access points (tablets, phones, kiosks). Token-based pairing.

| Method | Path | Description | Request Body | Response |
|--------|------|-------------|-------------|----------|
| GET | `/devices` | List all | — | `Device[]` |
| POST | `/devices` | Create | `Device` | `201` Device |
| GET | `/devices/{name}` | Get by name | — | `Device` |
| PUT | `/devices/{name}` | Update | `Device` | `Device` |
| DELETE | `/devices/{name}` | Delete | — | `204` |
| POST | `/devices/{name}/regenerate-token` | Regenerate auth token | — | `Device` (new token) |

**Device:**
```json
{
  "name": "kitchen-tablet",
  "token": "auto-generated-uuid",
  "defaultAgent": "default",
  "allowedAgents": ["default", "cooking-assistant"],
  "enabled": true
}
```

---

### Cron Jobs

CRUD for scheduled tasks that send prompts to agents.

| Method | Path | Description | Request Body | Response |
|--------|------|-------------|-------------|----------|
| GET | `/crons` | List all | — | `CronJob[]` |
| POST | `/crons` | Create | `CronJob` | `201` CronJob |
| GET | `/crons/{name}` | Get by name | — | `CronJob` |
| PUT | `/crons/{name}` | Update | `CronJob` | `CronJob` |
| DELETE | `/crons/{name}` | Delete | — | `204` |

**CronJob:**
```json
{
  "name": "daily-summary",
  "schedule": "0 8 * * *",
  "agentId": "default",
  "prompt": "Give me a summary of today's calendar",
  "description": "Morning briefing",
  "enabled": true
}
```

---

## Persistencia

- El store persiste a un fichero JSON (`data/store.json`) en cada escritura
- Si no existe `data/store.json` al arrancar, el store empieza vacío — todo se configura vía Admin API
- `config.yaml` solo contiene configuración de infraestructura del servidor (puertos, log)

## Estructura de ficheros

```
server/
├── store/
│   ├── store.go        — Store struct, Load/Save, CRUD operations
│   ├── types.go        — AgentDefinition, BackendDefinition, MemoryProvider, etc.
├── memory/
│   ├── provider.go     — Provider interface, FieldSpec, Category, HealthResult
│   ├── registry.go     — Global registry: Register(), Get(), All(), ValidTypeForCategory()
│   ├── redis/redis.go  — Redis provider (session), ConfigFields, Ping via ParseURL
│   └── postgres/postgres.go — Postgres provider (longterm), ConfigFields, Ping via sql.Open
├── admin/
│   ├── handler.go      — Router + helpers (writeJSON, writeError)
│   ├── agents.go       — Agent CRUD + MCP linking handlers
│   ├── backends.go     — Backend CRUD handlers
│   ├── memory.go       — Memory CRUD + health check + /types handler
│   ├── mcps.go         — MCP Server CRUD handlers
│   ├── devices.go      — Device CRUD + token regeneration handlers
│   ├── crons.go        — Cron Job CRUD handlers
│   └── overview.go     — Overview/dashboard handler
```

## Migración desde config YAML anterior

La migración del YAML legacy ya no es automática. Los recursos que antes se definían en `config.yaml` (backends, agents, MCPs, memory, telegram) ahora se gestionan exclusivamente vía la Admin API. Si se viene de una instalación anterior, hay que recrear los recursos manualmente a través del Admin UI en `:8081`.

## Impacto en código existente

- `server/config/config.go` — solo parsea server, log, wakeWord del YAML
- `server/agent/agent.go` — acepta store types directamente (`AgentDefinition`, `BackendDefinition`, etc.)
- `server/main.go` — lee recursos del store, router de admin en puerto 8081, blank imports para provider registration
- `server/clients/telegram/` — sin cambios por ahora (se conecta al agente que tenga telegram configurado)

## Fases

### Fase 1 — Multi-Agent Store + Admin API
- [x] Store con tipos multi-agente
- [x] Admin API con todos los endpoints CRUD
- [x] Migración automática de config.yaml → store
- [x] Refactor de agent.go para aceptar config por agente
- [x] Wiring en main.go

### Fase 1.5 — Memory Providers (extensible)
- [x] Registry-based provider system (`server/memory/`)
- [x] Provider interface: `Type()`, `DisplayName()`, `SupportedCategories()`, `ConfigFields()`, `Ping(ctx, config)`
- [x] `FieldSpec` struct: declarative field definitions per provider (key, label, type, required, placeholder, default)
- [x] Redis provider (`server/memory/redis/`) — session, Ping via `ParseURL`
- [x] Postgres provider (`server/memory/postgres/`) — longterm, Ping via `sql.Open`
- [x] Blank imports in `main.go` trigger `init()` registration
- [x] `MemoryProvider.Category` (per-instance) vs `SupportedCategories()` (per-type capability)
- [x] Config as opaque `map[string]interface{}` — provider-specific fields
- [x] Universal `connectionString` for all providers (Redis via `redis://`, Postgres via `postgres://`)
- [x] Provider-specific extra fields in config map (e.g. `ttl` for Redis) declared via `ConfigFields()`
- [x] `Embedding *BackendRef` (pointer) — omitted for session providers, present for longterm
- [x] Admin UI: Memory tab split into Session / Long-Term sections
- [x] Admin UI: Schema-driven forms — fields rendered dynamically from `ConfigFields()` specs via `/memory/types`
- [x] Admin UI: Zero hardcoded fields per provider type — new providers get forms for free
- [x] Admin UI: Embedding fieldset shown only for longterm category
- [x] Admin UI: Test Connection button (card + dialog) via `/memory/{name}/health`
- [x] Admin API: `GET /memory/types` returns registered types with categories + field specs
- [x] Admin API: `GET /memory/{name}/health` real-time Ping with 5s timeout

### Fase 1.6 — Swagger/OpenAPI Documentation
- [x] `swaggo/swag` annotations on all 34 admin API handlers
- [x] `swaggo/http-swagger` Swagger UI wired at `/swagger/`
- [x] Generated `server/admin/docs/` (docs.go, swagger.json, swagger.yaml)
- [x] `make swagger` Makefile target for regeneration
- [x] `ErrorResponse` type for consistent error schemas
- [x] `MemoryTypeInfo` exported type for `/memory/types` response model
- [x] All store types auto-discovered from json struct tags

### Fase 2 — Multi-Agent Runtime + Voice Endpoints
- [x] Hot-reload de agentes (crear agente → arrancarlo sin reiniciar) via `OnChange()` channel + `agentRouterHandler`
- [x] Selector de agentes en voice-ui (dropdown + `setAgent()` en todos los componentes)
- [x] Multi-agent ADK: `agent.New()` acepta `[]AgentDefinition`, `NewMultiLoader` enruta por `appName`
- [x] Voice endpoints: `/api/v1/voice/{agentId}/speech` y `/transcription` (resolución dinámica de backend por agente)
- [x] Voice proxy API key forwarding al upstream
- [x] Rename con cascade: 6 métodos en store, actualizaciones atómicas de referencias
- [x] Admin UI rename: campos name/ID editables en modo edición
- [x] Wake word model name en capabilities WebSocket

### Fase 3 (futuro)
- [ ] Autenticación en admin API
- [ ] Persistencia en base de datos en vez de JSON
- [ ] Multi-tenant (múltiples usuarios con sus propios agentes)
