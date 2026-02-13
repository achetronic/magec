# Client Entity — Design Document

## Overview

The `Client` entity unifies authentication/authorization for all access points: voice-ui tablets, Telegram bots, webhooks, cron jobs, and any future integration. Each client type declares its configuration via **JSON Schema**.

## Current State (Implemented)

### Client Types

| Type | Config Schema | Use Case |
|------|--------------|----------|
| `direct` | `{}` (empty) | Voice-UI tablets, apps — token-only auth |
| `telegram` | `botToken`, `allowedUsers`, `allowedChats`, `responseMode` | Telegram bot |
| `cron` | `schedule`, `commandId` | Scheduled automation |
| `webhook` | `passthrough` XOR `commandId` (via `oneOf`) | HTTP endpoint for integrations |

### Data Model

```go
type ClientDefinition struct {
    ID            string       `json:"id"`
    Name          string       `json:"name"`
    Type          string       `json:"type"`
    Token         string       `json:"token"`
    AllowedAgents []string     `json:"allowedAgents"`
    Enabled       bool         `json:"enabled"`
    Config        ClientConfig `json:"config"`
}

type ClientConfig struct {
    Telegram *TelegramClientConfig `json:"telegram,omitempty"`
    Discord  *DiscordClientConfig  `json:"discord,omitempty"`
    Slack    *SlackClientConfig    `json:"slack,omitempty"`
    Cron     *CronClientConfig     `json:"cron,omitempty"`
    Webhook  *WebhookClientConfig  `json:"webhook,omitempty"`
}

type TelegramClientConfig struct {
    BotToken     string  `json:"botToken"`
    AllowedUsers []int64 `json:"allowedUsers"`
    AllowedChats []int64 `json:"allowedChats"`
    ResponseMode string  `json:"responseMode"`
}

type CronClientConfig struct {
    Schedule  string `json:"schedule"`
    CommandID string `json:"commandId"`
}

type WebhookClientConfig struct {
    Passthrough bool   `json:"passthrough"`
    CommandID   string `json:"commandId,omitempty"`
}
```

### JSON Examples

**voice-ui tablet (direct):**
```json
{
  "id": "uuid-v4",
  "name": "tablet-salon",
  "type": "direct",
  "token": "mgc_aaa...",
  "allowedAgents": ["agent-uuid-1", "agent-uuid-2"],
  "enabled": true,
  "config": {}
}
```

**Telegram bot:**
```json
{
  "id": "uuid-v4",
  "name": "familia-telegram",
  "type": "telegram",
  "token": "mgc_ccc...",
  "allowedAgents": ["agent-uuid-1"],
  "enabled": true,
  "config": {
    "telegram": {
      "botToken": "123456:ABC-DEF...",
      "allowedUsers": [111111, 222222],
      "allowedChats": [],
      "responseMode": "both"
    }
  }
}
```

**Cron job:**
```json
{
  "id": "uuid-v4",
  "name": "resumen-diario",
  "type": "cron",
  "token": "mgc_eee...",
  "allowedAgents": ["agent-uuid-1"],
  "enabled": true,
  "config": {
    "cron": {
      "schedule": "0 8 * * *",
      "commandId": "command-uuid-1"
    }
  }
}
```

**Webhook (fixed command):**
```json
{
  "id": "uuid-v4",
  "name": "github-webhook",
  "type": "webhook",
  "token": "mgc_fff...",
  "allowedAgents": ["agent-uuid-1"],
  "enabled": true,
  "config": {
    "webhook": {
      "passthrough": false,
      "commandId": "command-uuid-1"
    }
  }
}
```

**Webhook (passthrough):**
```json
{
  "id": "uuid-v4",
  "name": "api-externa",
  "type": "webhook",
  "token": "mgc_ggg...",
  "allowedAgents": ["agent-uuid-1"],
  "enabled": true,
  "config": {
    "webhook": {
      "passthrough": true
    }
  }
}
```

## Client Type Provider Registry

### Architecture

```
server/client/
├── provider.go         — Provider interface, Schema type alias
├── registry.go         — Global registry: Register(), ValidateConfig() with oneOf
├── direct/direct.go    — Direct provider (empty schema)
├── telegram/telegram.go — Telegram provider (JSON Schema with x-format, enum)
├── cron/cron.go        — Cron provider (JSON Schema with x-entity)
└── webhook/webhook.go  — Webhook provider (JSON Schema with oneOf branches)
```

### Provider Interface

```go
type Schema = map[string]interface{}

type Provider interface {
    Type() string
    DisplayName() string
    ConfigSchema() Schema
}
```

- `ConfigSchema()` returns a full JSON Schema object with `type`, `properties`, `required`, and optional `oneOf`
- JSON Schema extensions: `x-entity`, `x-format`, `x-placeholder`
- Providers register via `init()` + blank imports in `main.go`

### Config Validation

`ValidateConfig(providerType, configBlock)` walks the JSON Schema recursively:
- Checks `required` fields exist in the data
- Validates `properties` types
- For `oneOf`: uses `matchOneOf()` to find the matching branch by comparing `const` values against actual data, then validates that branch's requirements

### Admin API Endpoint

`GET /api/v1/admin/clients/types` returns:

```json
[
  {
    "type": "direct",
    "displayName": "Direct",
    "configSchema": {}
  },
  {
    "type": "telegram",
    "displayName": "Telegram",
    "configSchema": {
      "type": "object",
      "required": ["botToken"],
      "properties": {
        "botToken": {"type": "string", "x-format": "password", "x-placeholder": "123456:ABC-DEF..."},
        "allowedUsers": {"type": "string", "x-placeholder": "Comma-separated user IDs"},
        "allowedChats": {"type": "string", "x-placeholder": "Comma-separated chat IDs"},
        "responseMode": {"type": "string", "enum": ["text", "voice", "mirror", "both"], "default": "text"}
      }
    }
  },
  {
    "type": "cron",
    "displayName": "Cron",
    "configSchema": {
      "type": "object",
      "required": ["schedule", "commandId"],
      "properties": {
        "schedule": {"type": "string", "x-placeholder": "0 8 * * *"},
        "commandId": {"type": "string", "x-entity": "commands"}
      }
    }
  },
  {
    "type": "webhook",
    "displayName": "Webhook",
    "configSchema": {
      "type": "object",
      "properties": {
        "passthrough": {"type": "boolean", "default": false},
        "commandId": {"type": "string", "x-entity": "commands"}
      },
      "oneOf": [
        {
          "properties": {"passthrough": {"const": false}},
          "required": ["commandId"]
        },
        {
          "properties": {"passthrough": {"const": true}}
        }
      ]
    }
  }
]
```

### JSON Schema Extensions

| Extension | Purpose | Example |
|-----------|---------|---------|
| `x-entity` | UI renders a `<select>` populated from the named store collection | `"x-entity": "commands"` |
| `x-format` | UI renders password input instead of text | `"x-format": "password"` |
| `x-placeholder` | Placeholder text for input fields | `"x-placeholder": "0 8 * * *"` |

### Frontend Rendering (ClientDialog.vue)

The dialog renders forms dynamically from JSON Schema:
- `currentSchema` computed: finds the matching type's `configSchema`
- `activeOneOfBranch` computed: evaluates `oneOf` branches by matching `const` values against form data
- `visibleProperties` computed: shows/hides fields based on active `oneOf` branch
- Renders: `boolean` → toggle, `x-entity` → select from store, `enum` → select, default → text/password input
- `isFieldRequired()`: checks both top-level and branch-level `required`
- `onTypeChange()`: resets config and applies defaults from schema

## Admin API Endpoints

Base path: `/api/v1/admin`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/clients` | List all clients |
| POST | `/clients` | Create a client (token auto-generated as `mgc_...`) |
| GET | `/clients/types` | List registered client types with JSON schemas |
| GET | `/clients/{id}` | Get a client by ID |
| PUT | `/clients/{id}` | Update a client |
| DELETE | `/clients/{id}` | Delete a client |
| POST | `/clients/{id}/regenerate-token` | Regenerate auth token |

## Webhook Endpoint (User API)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/webhooks/{clientId}` | Fire a webhook client |

- Auth: `Authorization: Bearer <mgc_token>` (client's own token)
- Passthrough body: `{"prompt": "your text here"}`
- Fixed command: body empty or ignored
- Executes against all `allowedAgents`
- Bypasses `clientAuthMiddleware` (own auth in webhook.go)

## Key Decisions

- **`allowedAgents[0]` is the default agent** — no separate `defaultAgent` field
- **JSON Schema replaces FieldSpec** — full OpenAPI JSON Schema per type with extensions for UI rendering
- **`oneOf` for exclusive config** — webhook's passthrough XOR commandId enforced at schema level
- **Client token for webhook auth** — no separate `secret` field. One auth mechanism everywhere
- **Cron/webhook execute against ALL allowedAgents** — not a single agentId
- **Config is typed Go structs** — each platform has its own struct in `ClientConfig`
- **Telegram config lives in Client, not Agent** — auth/access belongs in Client entity

## Migration Chain

On `loadFromDisk()`, these migrations run in order (all idempotent):

1. `devices → clients` — Legacy Device entities become type `direct` clients
2. `cronJobs → triggers` — Legacy CronJob becomes Command + Trigger
3. `triggers → clients` — Trigger entities become cron/webhook client types with own tokens
4. `device → direct` — Client type `device` renamed to `direct`
5. `migrateIDs` — Generates UUID v4 for any entity missing one

## Future (out of scope)

- **Discord provider** — `server/client/discord/`
- **Slack provider** — `server/client/slack/`
- **Memory provider migration to JSON Schema** — `server/memory/` still uses FieldSpec
- **Enrollment** — `open` / `closed` / `approval` modes for user self-registration
