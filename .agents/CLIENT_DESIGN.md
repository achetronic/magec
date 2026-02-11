# Client Entity — Design Document

## Overview

Replace the `Device` entity with a new `Client` entity that unifies authentication/authorization for all access points: voice-ui tablets, Telegram bots, Discord bots, Slack bots, webhooks, crons, and any future integration.

## Motivation

Currently:
- `Device` handles voice-ui tablet auth (token + allowedAgents)
- `TelegramConfig` lives inside `AgentDefinition` (wrong place — auth config mixed with agent logic)
- No unified model for Discord, Slack, webhooks, or other future clients

The `Client` entity centralizes "who can talk to which agents" in one place.

## Data Model

### ClientDefinition

```go
type ClientDefinition struct {
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
}

type TelegramClientConfig struct {
    BotToken     string  `json:"botToken"`
    AllowedUsers []int64 `json:"allowedUsers"`
    AllowedChats []int64 `json:"allowedChats"`
    ResponseMode string  `json:"responseMode"`
}

type DiscordClientConfig struct {
    BotToken        string   `json:"botToken"`
    GuildID         string   `json:"guildId"`
    AllowedUsers    []string `json:"allowedUsers"`
    AllowedChannels []string `json:"allowedChannels"`
}

type SlackClientConfig struct {
    BotToken        string   `json:"botToken"`
    SigningSecret   string   `json:"signingSecret"`
    AllowedUsers    []string `json:"allowedUsers"`
    AllowedChannels []string `json:"allowedChannels"`
}
```

### Types

| type | config | Use case |
|------|--------|----------|
| `device` | `{}` | voice-ui tablets, webhooks, crons — token-only auth |
| `telegram` | `config.telegram` | Telegram bot |
| `discord` | `config.discord` | Discord bot |
| `slack` | `config.slack` | Slack bot |

### JSON Examples

**voice-ui tablet:**
```json
{
  "name": "tablet-salon",
  "type": "device",
  "token": "mgc_aaa",
  "allowedAgents": ["magec", "cooking"],
  "enabled": true,
  "config": {}
}
```

**Telegram (family):**
```json
{
  "name": "familia-telegram",
  "type": "telegram",
  "token": "mgc_ccc",
  "allowedAgents": ["magec", "itahisa"],
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

**Telegram (private):**
```json
{
  "name": "alby-privado",
  "type": "telegram",
  "token": "mgc_ddd",
  "allowedAgents": ["magec", "privado"],
  "enabled": true,
  "config": {
    "telegram": {
      "botToken": "789012:GHI-JKL...",
      "allowedUsers": [111111],
      "allowedChats": [],
      "responseMode": "mirror"
    }
  }
}
```

**Webhook:**
```json
{
  "name": "webhook-externo",
  "type": "webhook",
  "token": "mgc_ggg",
  "allowedAgents": ["magec"],
  "enabled": true,
  "config": {}
}
```

### Key Decisions

- **`allowedAgents[0]` is the default agent** — no separate `defaultAgent` field
- **`type` is redundant with config key** — explicit is better, self-documents the JSON, Go reads `config.Telegram` directly
- **`device` type covers voice-ui, webhooks, and crons** — they all only need token auth
- **Config is typed structs, not `map[string]interface{}`** — each platform has its own Go struct
- **Telegram config moves OUT of AgentDefinition** — auth belongs in Client, not Agent

## Dynamic Form Fields (Client Type Registry)

Following the same pattern as `server/memory/` provider registry:

### Registry (`server/client/`)

```
server/client/
├── provider.go     — Provider interface, FieldSpec (reuse from memory)
├── registry.go     — Global registry: Register(), Get(), All(), ValidType()
├── device/         — Device provider (no extra fields)
├── telegram/       — Telegram provider (botToken, allowedUsers, etc.)
├── discord/        — Discord provider (future)
└── slack/          — Slack provider (future)
```

### Provider Interface

```go
type Provider interface {
    Type() string
    DisplayName() string
    ConfigFields() []FieldSpec
}
```

- `ConfigFields()` returns field specs for the platform-specific config
- `device` type returns empty fields (no config needed)
- `telegram` type returns botToken, allowedUsers, allowedChats, responseMode
- Admin UI renders forms dynamically from `/clients/types` endpoint

### Admin API Endpoint

`GET /api/v1/admin/clients/types` returns:

```json
[
  {
    "type": "device",
    "displayName": "Device",
    "fields": []
  },
  {
    "type": "telegram",
    "displayName": "Telegram",
    "fields": [
      {"key": "botToken", "label": "Bot Token", "type": "password", "required": true, "placeholder": "123456:ABC-DEF..."},
      {"key": "allowedUsers", "label": "Allowed Users", "type": "text", "placeholder": "Comma-separated Telegram user IDs"},
      {"key": "allowedChats", "label": "Allowed Chats", "type": "text", "placeholder": "Comma-separated Telegram chat IDs"},
      {"key": "responseMode", "label": "Response Mode", "type": "select", "default": "text", "options": "text,voice,mirror,both"}
    ]
  }
]
```

## Admin API Endpoints

Base path: `/api/v1/admin`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/clients` | List all clients |
| POST | `/clients` | Create a client (token auto-generated) |
| GET | `/clients/types` | List registered client types with field specs |
| GET | `/clients/{name}` | Get a client by name |
| PUT | `/clients/{name}` | Update a client |
| DELETE | `/clients/{name}` | Delete a client |
| POST | `/clients/{name}/regenerate-token` | Regenerate auth token |

## Rename Cascade

- **Agent renamed** → update `Client.AllowedAgents[]` (was Device.DefaultAgent + Device.AllowedAgents[])
- **Client renamed** → no cascading references (same as Device today)

## Migration from Device

### Store Changes

- `StoreData.Devices []Device` → `StoreData.Clients []ClientDefinition`
- Old `Device` struct removed
- Existing `data/store.json` with `"devices"` key: add migration in `loadFromDisk()` that converts `Device` → `ClientDefinition` with `type: "device"`

### Agent Changes

- Remove `Telegram TelegramConfig` from `AgentDefinition`
- Telegram config now lives in `ClientDefinition.Config.Telegram`
- `agent.New()` no longer receives Telegram config

### main.go Changes

- `deviceAuthMiddleware` → `clientAuthMiddleware` (same logic, reads from `Clients` instead of `Devices`)
- `/api/v1/device/info` → adapt to read from Client (return `allowedAgents` with `[0]` as default)
- Telegram client startup: iterate Clients with `type: "telegram"`, start a bot per client
- `X-Device-Name` header → `X-Client-Name`

### admin/ Changes

- `devices.go` → `clients.go` (CRUD adapted for ClientDefinition)
- Add `listClientTypes` handler (like `listMemoryTypes`)
- Update `handler.go` routes: `/devices/*` → `/clients/*`
- Update `overview.go`: `devicesCount` → `clientsCount`

### voice-ui Changes

- `DeviceAuth.js`: endpoint stays `/api/v1/device/info` (or rename to `/api/v1/client/info`)
- Response format stays the same: `{paired, name, allowedAgents}`

### admin-ui Changes

- Devices tab → Clients tab
- Dynamic form based on `type` selection (fetch fields from `/clients/types`)
- Platform-specific config rendered inside a fieldset

## Future (out of scope)

- **Command entity** — prompts with triggers (cron schedule, webhook). Client handles auth, Command handles execution.
- **Discord provider** — `server/client/discord/`
- **Slack provider** — `server/client/slack/`
- **Enrollment** — `open` / `closed` / `approval` modes for user self-registration

## Implementation Order

1. Write this design doc ✓
2. Add `ClientDefinition` types to `store/types.go`
3. Add Client CRUD + rename to `store/store.go`
4. Create `server/client/` registry (provider.go, registry.go)
5. Create `server/client/device/` provider (empty fields)
6. Create `server/client/telegram/` provider (field specs)
7. Add `admin/clients.go` handlers (CRUD + /types + regenerate-token)
8. Wire routes in `admin/handler.go`
9. Update `admin/overview.go`
10. Adapt `deviceAuthMiddleware` → `clientAuthMiddleware` in `main.go`
11. Adapt `/api/v1/device/info` endpoint
12. Migrate Telegram startup from Agent config to Client config
13. Update rename cascade (Agent → Client.AllowedAgents)
14. Add Device → Client migration in `loadFromDisk()`
15. Update admin-ui (Clients tab with dynamic forms)
16. Update voice-ui if endpoint changes
17. Update AGENTS.md, SESSION_CONTEXT.md, MULTI_AGENT_ADMIN_API.md
18. Build and verify
