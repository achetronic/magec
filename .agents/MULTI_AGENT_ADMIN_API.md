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
   - Nombre e ID único (UUID v4)
   - System prompt propio
   - Backend LLM propio (referencia a backends globales)
   - MCPs conectados individualmente
   - Configuración de TTS y transcripción propia

2. **API de administración** (`/api/v1/admin/`) con endpoints CRUD para:
   - **Agentes**: crear, listar, obtener, actualizar, eliminar
   - **Backends**: crear, listar, obtener, actualizar, eliminar
   - **Memory Providers**: crear, listar, obtener, actualizar, eliminar, health check, tipos
   - **MCP Servers**: crear, listar, obtener, actualizar, eliminar
   - **Clients**: crear, listar, obtener, actualizar, eliminar, regenerar token, tipos (JSON Schema)
   - **Commands**: crear, listar, obtener, actualizar, eliminar
   - **Flows**: crear, listar, obtener, actualizar, eliminar
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
│   ├── id, name, type, category (session | longterm)
│   ├── config: {connectionString, ...extras por tipo}
│   └── embedding: *BackendRef (solo longterm, nil para session)
├── MCPServers[]        — servidores MCP reutilizables (globales)
├── Agents[]            — cada agente es una unidad independiente
│   ├── id, name, description, outputKey
│   ├── systemPrompt
│   ├── llm: {backend, model}
│   ├── transcription: {backend, model}
│   ├── tts: {backend, model, voice, speed}
│   ├── memory: {session: "provider-id", longTerm: "provider-id"}
│   └── mcpServers: ["id1", "id2"]  — referencias a MCPs globales
├── Clients[]           — puntos de acceso con auth por token
│   ├── id, name, type, token, allowedAgents, enabled
│   └── config: {telegram?, cron?, webhook?}  — JSON Schema driven
├── Commands[]          — prompts reutilizables
│   ├── id, name, description, prompt
├── Flows[]             — workflows multi-agente
│   ├── id, name, description
│   └── root: FlowStep (recursive tree)
├── CronJobs[]          — legacy, auto-migrado a clients tipo cron
└── Triggers[]          — legacy, auto-migrado a clients tipo cron/webhook
```

### Jerarquía de recursos

```
Backends (AI infra) → Memory (data infra) → MCPs (tools) → Agents (consumers) → Commands (prompts) → Clients (access + automation)
                                                                                                         └── Flows (workflows)
```

## API Reference

Base path: `/api/v1/admin`

All endpoints return JSON. Error responses use `{"error": "message"}`.

---

### Overview

| Method | Path | Description | Response |
|--------|------|-------------|----------|
| GET | `/overview` | Dashboard summary | Counts + agent summaries |

---

### Backends

CRUD for reusable AI backends (OpenAI, Ollama, Anthropic, Gemini).

| Method | Path | Description |
|--------|------|-------------|
| GET | `/backends` | List all |
| POST | `/backends` | Create |
| GET | `/backends/{id}` | Get by ID |
| PUT | `/backends/{id}` | Update |
| DELETE | `/backends/{id}` | Delete |

**BackendDefinition:**
```json
{
  "id": "uuid-v4",
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

| Method | Path | Description |
|--------|------|-------------|
| GET | `/memory` | List all |
| POST | `/memory` | Create |
| GET | `/memory/types` | Registered provider types (with FieldSpec) |
| GET | `/memory/{id}` | Get by ID |
| PUT | `/memory/{id}` | Update |
| DELETE | `/memory/{id}` | Delete |
| GET | `/memory/{id}/health` | Real-time ping (5s timeout) |

**MemoryProvider:**
```json
{
  "id": "uuid-v4",
  "name": "redis-session",
  "type": "redis",
  "category": "session",
  "config": {
    "connectionString": "redis://localhost:6379/0",
    "ttl": "24h"
  }
}
```

---

### MCP Servers

CRUD for global MCP (Model Context Protocol) tool servers.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/mcps` | List all |
| POST | `/mcps` | Create |
| GET | `/mcps/{id}` | Get by ID |
| PUT | `/mcps/{id}` | Update |
| DELETE | `/mcps/{id}` | Delete |

**MCPServer:**
```json
{
  "id": "uuid-v4",
  "name": "home-assistant",
  "type": "http",
  "endpoint": "http://localhost:8070/mcp",
  "headers": {"Authorization": "Bearer token..."},
  "systemPrompt": "Home automation tools"
}
```

Types: `http` (StreamableClientTransport), `stdio` (CommandTransport)

---

### Agents

CRUD for agent definitions. Each agent is an independent unit with its own LLM, memory, tools, and prompts.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/agents` | List all |
| POST | `/agents` | Create |
| GET | `/agents/{id}` | Get by ID |
| PUT | `/agents/{id}` | Update |
| DELETE | `/agents/{id}` | Delete |
| GET | `/agents/{id}/mcps` | List resolved MCPs |
| PUT | `/agents/{id}/mcps/{mcpId}` | Link MCP to agent |
| DELETE | `/agents/{id}/mcps/{mcpId}` | Unlink MCP from agent |

**AgentDefinition:**
```json
{
  "id": "uuid-v4",
  "name": "Magec",
  "description": "Default voice assistant agent",
  "systemPrompt": "You are a helpful assistant...",
  "outputKey": "",
  "llm": {"backend": "backend-uuid", "model": "qwen3:8b"},
  "transcription": {"backend": "backend-uuid", "model": "whisper-1"},
  "tts": {"backend": "backend-uuid", "model": "tts-1", "voice": "alloy", "speed": 1.0},
  "memory": {"session": "memory-uuid", "longTerm": "memory-uuid"},
  "mcpServers": ["mcp-uuid-1", "mcp-uuid-2"]
}
```

---

### Clients

CRUD for access points. Each client has a type with config defined by JSON Schema.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/clients` | List all |
| POST | `/clients` | Create (token auto-generated as `mgc_...`) |
| GET | `/clients/types` | List registered types with JSON schemas |
| GET | `/clients/{id}` | Get by ID |
| PUT | `/clients/{id}` | Update |
| DELETE | `/clients/{id}` | Delete |
| POST | `/clients/{id}/regenerate-token` | Regenerate auth token |

**ClientDefinition:**
```json
{
  "id": "uuid-v4",
  "name": "tablet-salon",
  "type": "direct",
  "token": "mgc_auto-generated",
  "allowedAgents": ["agent-uuid-1"],
  "enabled": true,
  "config": {}
}
```

**Client Types** (from `GET /clients/types`):

| Type | Config Schema | Description |
|------|--------------|-------------|
| `direct` | `{}` | Voice-UI, apps — token-only |
| `telegram` | `botToken` (req), `allowedUsers`, `allowedChats`, `responseMode` (enum) | Telegram bot |
| `cron` | `schedule` (req), `commandId` (req, x-entity:commands) | Scheduled task |
| `webhook` | `passthrough` (bool) + `commandId` (oneOf exclusive) | HTTP endpoint |

See [CLIENT_DESIGN.md](CLIENT_DESIGN.md) for full JSON Schema details and examples.

---

### Commands

CRUD for reusable prompts. Referenced by cron and webhook clients via `commandId`.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/commands` | List all |
| POST | `/commands` | Create |
| GET | `/commands/{id}` | Get by ID |
| PUT | `/commands/{id}` | Update |
| DELETE | `/commands/{id}` | Delete |

**Command:**
```json
{
  "id": "uuid-v4",
  "name": "resumen-diario",
  "description": "Genera resumen del día",
  "prompt": "Dame un resumen de las noticias de hoy"
}
```

---

### Flows

CRUD for multi-agent workflows. Recursive tree structure.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/flows` | List all |
| POST | `/flows` | Create |
| GET | `/flows/{id}` | Get by ID |
| PUT | `/flows/{id}` | Update |
| DELETE | `/flows/{id}` | Delete |

**FlowDefinition:**
```json
{
  "id": "uuid-v4",
  "name": "Test Flow",
  "description": "Sequential then parallel",
  "root": {
    "type": "sequential",
    "steps": [
      {"type": "agent", "agentId": "agent-uuid-1"},
      {
        "type": "parallel",
        "steps": [
          {"type": "agent", "agentId": "agent-uuid-2"},
          {"type": "agent", "agentId": "agent-uuid-3"}
        ]
      }
    ]
  }
}
```

---

### Cron Jobs (Legacy)

CRUD for legacy cron jobs. **Auto-migrated** to Commands + cron-type Clients on store load.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/crons` | List all |
| POST | `/crons` | Create |
| GET | `/crons/{id}` | Get by ID |
| PUT | `/crons/{id}` | Update |
| DELETE | `/crons/{id}` | Delete |

---

## Webhook Endpoint (User API, port 8080)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/webhooks/{clientId}` | Fire a webhook client |

- Auth: `Authorization: Bearer <mgc_token>`
- Passthrough: `{"prompt": "text"}`
- Fixed command: body empty/ignored
- Swagger docs: `server/api/user/docs/`

---

## Persistencia

- El store persiste a un fichero JSON (`data/store.json`) en cada escritura
- Si no existe `data/store.json` al arrancar, el store empieza vacío — todo se configura vía Admin API
- `config.yaml` solo contiene configuración de infraestructura del servidor (puertos, log)

## Estructura de ficheros

```
server/
├── store/
│   ├── store.go        — Store struct, Load/Save, CRUD operations, migration chain
│   └── types.go        — All entity types + legacy types for migration
├── client/
│   ├── provider.go     — Provider interface: Type(), DisplayName(), ConfigSchema()
│   ├── registry.go     — Global registry: Register(), ValidateConfig() with oneOf support
│   ├── direct/         — Direct provider (empty schema)
│   ├── telegram/       — Telegram provider (JSON Schema)
│   ├── cron/           — Cron provider (JSON Schema with x-entity)
│   └── webhook/        — Webhook provider (JSON Schema with oneOf)
├── memory/
│   ├── provider.go     — Provider interface, FieldSpec, Category, HealthResult
│   ├── registry.go     — Global registry: Register(), Get(), All(), ValidTypeForCategory()
│   ├── redis/redis.go  — Redis provider (session), ConfigFields, Ping via ParseURL
│   └── postgres/postgres.go — Postgres provider (longterm), ConfigFields, Ping via sql.Open
├── clients/
│   ├── executor.go     — RunClient() — executes commands against all allowedAgents
│   ├── cron/
│   │   ├── cron.go     — Cron expression parser
│   │   └── scheduler.go — Cron scheduler — filters cron-type clients
│   ├── webhook/webhook.go — Webhook HTTP handler — Bearer token auth
│   └── telegram/telegram.go — Telegram bot client
├── api/
│   ├── admin/
│   │   ├── handler.go  — Router + helpers (writeJSON, writeError)
│   │   ├── agents.go   — Agent CRUD + MCP linking handlers
│   │   ├── backends.go — Backend CRUD handlers
│   │   ├── clients.go  — Client CRUD + /types (JSON Schema) + token regeneration
│   │   ├── commands.go — Command CRUD handlers
│   │   ├── memory.go   — Memory CRUD + health check + /types handler
│   │   ├── mcps.go     — MCP Server CRUD handlers
│   │   ├── flows.go    — Flow CRUD handlers + recursive validation
│   │   └── docs/       — Generated swagger (swagger.json/yaml)
│   └── user/
│       ├── handlers.go — Health, DeviceInfo, Voice stubs, Webhook swagger types
│       ├── doc.go      — Swagger metadata
│       └── docs/       — Generated swagger (userapi_swagger.json/yaml)
```

## Migración desde config YAML anterior

La migración del YAML legacy ya no es automática. Los recursos que antes se definían en `config.yaml` ahora se gestionan exclusivamente vía la Admin API. Si se viene de una instalación anterior, hay que recrear los recursos manualmente a través del Admin UI en `:8081`.

## Cadena de migración automática (store)

Al cargar `data/store.json`, estas migraciones se ejecutan en orden (todas idempotentes):

1. `devices → clients` — Devices legacy se convierten en clients tipo `device`
2. `cronJobs → triggers` — CronJobs legacy se convierten en Command + Trigger
3. `triggers → clients` — Triggers se convierten en clients tipo `cron` o `webhook`
4. `device → direct` — Client type `device` se renombra a `direct`
5. `migrateIDs` — Genera UUID v4 para entidades sin ID

## Impacto en código existente

- `server/config/config.go` — solo parsea server, log, wakeWord del YAML
- `server/agent/agent.go` — acepta store types directamente (`AgentDefinition`, `BackendDefinition`, etc.)
- `server/main.go` — lee recursos del store, router de admin en puerto 8081, blank imports para provider registration
- `server/clients/telegram/` — se conecta al client con config telegram

## Fases

### Fase 1 — Multi-Agent Store + Admin API ✅
- [x] Store con tipos multi-agente
- [x] Admin API con todos los endpoints CRUD
- [x] Refactor de agent.go para aceptar config por agente
- [x] Wiring en main.go

### Fase 1.5 — Memory Providers (extensible) ✅
- [x] Registry-based provider system (`server/memory/`)
- [x] Redis provider (session) + Postgres provider (longterm)
- [x] Admin UI: schema-driven forms from `ConfigFields()`
- [x] Health check endpoint

### Fase 1.6 — Swagger/OpenAPI Documentation ✅
- [x] Admin API swagger + Swagger UI
- [x] User API swagger (webhook endpoint)

### Fase 2 — Multi-Agent Runtime + Voice Endpoints ✅
- [x] Hot-reload de agentes via `OnChange()` + debounce
- [x] Multi-agent ADK routing via `NewMultiLoader`
- [x] Voice endpoints per-agent
- [x] Rename con cascade

### Fase 2.5 — Commands + Clients Consolidation ✅
- [x] Commands entity (reusable prompts)
- [x] Triggers→Clients consolidation (cron/webhook as client types)
- [x] Client type registry migrated from FieldSpec to JSON Schema
- [x] `device` → `direct` rename
- [x] Migration chain: CronJobs→Triggers→Clients
- [x] Cron scheduler + Webhook HTTP handler
- [x] Webhook endpoint with Bearer token auth + Swagger docs

### Fase 3 (futuro)
- [ ] Autenticación en admin API
- [ ] Persistencia en base de datos en vez de JSON
- [ ] Multi-tenant (múltiples usuarios con sus propios agentes)
