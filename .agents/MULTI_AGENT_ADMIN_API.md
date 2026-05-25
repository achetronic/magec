# Multi-Agent Admin API

## Data Model

```
Store (in-memory + JSON persistence → data/store.json)
├── Backends[]          — reusable backend pools (global)
├── MemoryProviders[]   — memory providers (global, extensible)
│   ├── id, name, type, category (session | longterm)
│   ├── config: {connectionString, ...extras per type}
│   └── embedding: *BackendRef (longterm only, nil for session)
├── MCPServers[]        — reusable MCP servers (global)
│   ├── endpoint, type (http|stdio), headers, insecure
│   ├── command, args, env, workDir (stdio transport)
│   └── systemPrompt
├── Skills[]            — reusable skill packs (instructions + reference files)
│   ├── id, name, description, instructions
│   └── references: [{filename, size}]  — files stored at data/skills/{skillId}/
├── Agents[]            — each agent is an independent unit
│   ├── id, name, description, outputKey
│   ├── systemPrompt
│   ├── llm: {backend, model}
│   ├── transcription: {backend, model}
│   ├── tts: {backend, model, voice, speed}
│   ├── mcpServers: ["id1", "id2"]
│   ├── skills: ["id1", "id2"]
│   ├── tags: ["tag1", "tag2"]
│   ├── contextGuard: {enabled, strategy, maxTurns}
│   └── a2a: {enabled}  — exposes agent via A2A protocol
├── Clients[]           — access points with token-based auth
│   ├── id, name, type, token, allowedAgents, enabled
│   └── config: {telegram?, slack?, cron?, webhook?}  — JSON Schema driven
├── Commands[]          — reusable prompts
│   ├── id, name, description, prompt
├── Flows[]             — multi-agent workflows
│   ├── id, name, description
│   ├── root: FlowStep (recursive tree with responseAgent flag)
│   └── a2a: {enabled}  — exposes flow via A2A protocol
├── Secrets[]           — encrypted key-value pairs for env var injection
│   ├── id, name, key, value, description
└── Settings            — global memory provider selection
    ├── sessionProvider, longTermProvider
```

### Resource Hierarchy

```
Backends (AI infra) → Memory (data infra) → MCPs (tools) → Skills (knowledge) → Agents (consumers) → Commands (prompts) → Clients (access + automation)
                                                                                                                            └── Flows (workflows)
```

## API Reference

Base path: `/api/v1/admin`. All endpoints return JSON. Errors: `{"error": "message"}`.

### Backends

| Method         | Path             | Description           |
| -------------- | ---------------- | --------------------- |
| GET/POST       | `/backends`      | List / Create         |
| GET/PUT/DELETE | `/backends/{id}` | Get / Update / Delete |

Types: `openai`, `anthropic`, `gemini`

### Memory Providers

| Method         | Path                  | Description                       |
| -------------- | --------------------- | --------------------------------- |
| GET/POST       | `/memory`             | List / Create                     |
| GET            | `/memory/types`       | Registered types with JSON Schema |
| GET/PUT/DELETE | `/memory/{id}`        | Get / Update / Delete             |
| GET            | `/memory/{id}/health` | Real-time ping (5s timeout)       |

### MCP Servers

| Method         | Path         | Description           |
| -------------- | ------------ | --------------------- |
| GET/POST       | `/mcps`      | List / Create         |
| GET/PUT/DELETE | `/mcps/{id}` | Get / Update / Delete |

Types: `http` (StreamableClientTransport), `stdio` (CommandTransport)

### Skills

| Method         | Path                                 | Description                                                          |
| -------------- | ------------------------------------ | -------------------------------------------------------------------- |
| GET/POST       | `/skills`                            | List / Create                                                        |
| GET/PUT/DELETE | `/skills/{id}`                       | Get / Update / Delete                                                |
| POST           | `/skills/{id}/references`            | Upload reference file (multipart, 10MB limit)                        |
| GET/DELETE     | `/skills/{id}/references/{filename}` | Download / Delete reference file                                     |
| POST           | `/skills/{id}/package`               | Upload skill package (ZIP/tar.gz, 50MB limit, must contain SKILL.md) |

Skill instructions and reference file contents are injected into the agent system prompt at runtime. Files stored on disk at `data/skills/{skillId}/`, metadata only in store.

### Agents

| Method         | Path                        | Description           |
| -------------- | --------------------------- | --------------------- |
| GET/POST       | `/agents`                   | List / Create         |
| GET/PUT/DELETE | `/agents/{id}`              | Get / Update / Delete |
| GET            | `/agents/{id}/mcps`         | List resolved MCPs    |
| PUT/DELETE     | `/agents/{id}/mcps/{mcpId}` | Link / Unlink MCP     |

### Clients

| Method         | Path                             | Description                                       |
| -------------- | -------------------------------- | ------------------------------------------------- |
| GET/POST       | `/clients`                       | List / Create (token auto-generated as `mgc_...`) |
| GET            | `/clients/types`                 | Registered types with JSON Schema                 |
| GET/PUT/DELETE | `/clients/{id}`                  | Get / Update / Delete                             |
| POST           | `/clients/{id}/regenerate-token` | Regenerate auth token                             |

Client types: `direct`, `telegram`, `discord`, `slack`, `cron`, `webhook`. See [CLIENT_DESIGN.md](CLIENT_DESIGN.md).

### Commands

| Method         | Path             | Description           |
| -------------- | ---------------- | --------------------- |
| GET/POST       | `/commands`      | List / Create         |
| GET/PUT/DELETE | `/commands/{id}` | Get / Update / Delete |

### Flows

| Method         | Path          | Description           |
| -------------- | ------------- | --------------------- |
| GET/POST       | `/flows`      | List / Create         |
| GET/PUT/DELETE | `/flows/{id}` | Get / Update / Delete |

### Secrets

| Method         | Path            | Description                                                        |
| -------------- | --------------- | ------------------------------------------------------------------ |
| GET/POST       | `/secrets`      | List / Create (GET never returns `value`)                          |
| GET/PUT/DELETE | `/secrets/{id}` | Get / Update / Delete (empty `value` on update preserves existing) |

### Settings

| Method  | Path        | Description                                                                              |
| ------- | ----------- | ---------------------------------------------------------------------------------------- |
| GET/PUT | `/settings` | Get / Update global settings (`sessionProvider`, `longTermProvider`, `temporaryDir`)     |

`temporaryDir` is the on-disk path used by tools that hand artifacts off to other filesystem-aware tools (`export_artifact`, see decision #26). Falls back to `os.TempDir()` via `Store.ResolveTemporaryDir()` when unset.

### Voice

| Method | Path           | Description                                                |
| ------ | -------------- | ---------------------------------------------------------- |
| GET    | `/voice/types` | Registered voice providers with TTS/STT JSON Schemas       |

Same registry pattern as `/clients/types` and `/memory/types`. Decision #21.

### Conversations

| Method | Path                                | Description                                             |
| ------ | ----------------------------------- | ------------------------------------------------------- |
| GET    | `/conversations`                    | List (with filters: agent, source, client, perspective) |
| GET    | `/conversations/{id}`               | Get conversation with full messages                     |
| DELETE | `/conversations/{id}`               | Delete conversation                                     |
| DELETE | `/conversations/clear`              | Clear all conversations                                 |
| GET    | `/conversations/stats`              | Aggregated stats (by agent, source, time)               |
| PUT    | `/conversations/{id}/summary`       | Generate AI summary of conversation                     |
| GET    | `/conversations/{id}/pair`          | Find paired perspective (admin↔user)                    |
| POST   | `/conversations/{id}/reset-session` | Delete ADK session for this conversation                |

### Auth

| Method | Path          | Description                                     |
| ------ | ------------- | ----------------------------------------------- |
| GET    | `/auth/check` | Verify admin credentials (returns 200 if valid) |

### Overview

| Method | Path        | Description                         |
| ------ | ----------- | ----------------------------------- |
| GET    | `/overview` | Dashboard: agent counts + summaries |

### Webhook Endpoint (User API, port 8080)

| Method | Path                          | Description                                                      |
| ------ | ----------------------------- | ---------------------------------------------------------------- |
| POST   | `/api/v1/webhooks/{clientId}` | Fire a webhook — Bearer token auth, passthrough or fixed command |

### Backup & Restore

| Method | Path                | Description                                                     |
| ------ | ------------------- | --------------------------------------------------------------- |
| GET    | `/settings/backup`  | Download a `.tar.gz` archive of the entire `data/` directory    |
| POST   | `/settings/restore` | Upload a `.tar.gz` to atomically replace all data (500MB limit) |

The backup archive contains `store.json`, `conversations.json`, and `skills/{id}/` files. On restore, the archive must contain a valid `store.json` at the root level. The current data directory is atomically swapped (rename) and both stores are reloaded in memory.

## Embedded MCP Server

The same admin surface is also exposed as MCP tools at `http://localhost:8082/` when `server.mcp.enabled: true`. Tools live in `server/api/mcp/`; they call `*store.Store` directly without going through the admin HTTP router (decision #30). Authentication reuses `server.adminPassword` as a bearer token. Skill upload/download and backup/restore are intentionally not exposed because they stream binary archives that do not map cleanly to MCP tool inputs. See `website/content/docs/admin-mcp-server.md` for the full tool catalogue and client setup.

## Persistence

- Store persists to `data/store.json` on each write
- If file doesn't exist at startup, store starts empty
- `config.yaml` only contains server infrastructure (ports, log, voice)

## Migration Chain (on store load)

All idempotent. Run on the raw JSON before unmarshal in `server/store/store.go`:

1. `migrateTTSConfig` — moves legacy `tts.speed` into `tts.config.openai.speed` and flat Gemini fields into `tts.config.gemini.*` (introduced with the voice provider registry, decision #21).

Legacy chains (`devices → clients`, `cronJobs → triggers`, `triggers → clients`, `device → direct`, `migrateIDs`) lived in earlier versions and have been removed once their target installations had migrated.

## Future Work

- [ ] Database persistence instead of JSON
- [ ] Multi-tenant (multiple users with their own agents)
- [ ] Conversation `perspective` field documentation (dual admin/user recording)
