# Client Entity — Design Document

## Overview

The `Client` entity unifies authentication/authorization for all access points: voice-ui tablets, Telegram bots, Discord bots, webhooks, cron jobs, and any future integration. Each client type declares its configuration via **JSON Schema**.

## Current State (Implemented)

### Client Types

| Type       | Config Schema                                                             | Use Case                                 |
| ---------- | ------------------------------------------------------------------------- | ---------------------------------------- |
| `direct`   | `{}` (empty)                                                              | Voice-UI tablets, apps — token-only auth |
| `telegram` | `botToken`, `allowedUsers`, `allowedChats`, `responseMode`                | Telegram bot                             |
| `discord`  | `botToken`, `allowedUsers`, `allowedChannels`, `responseMode`             | Discord bot (Gateway WebSocket)          |
| `slack`    | `botToken`, `appToken`, `allowedUsers`, `allowedChannels`, `responseMode` | Slack bot (Socket Mode)                  |
| `cron`     | `schedule`, `commandId`                                                   | Scheduled automation                     |
| `webhook`  | `passthrough` XOR `commandId` (via `oneOf`)                               | HTTP endpoint for integrations           |

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

type SlackClientConfig struct {
    BotToken        string   `json:"botToken"`
    AppToken        string   `json:"appToken"`
    AllowedUsers    []string `json:"allowedUsers"`
    AllowedChannels []string `json:"allowedChannels"`
    ResponseMode    string   `json:"responseMode"`
}

type DiscordClientConfig struct {
    BotToken        string   `json:"botToken"`
    AllowedUsers    []string `json:"allowedUsers"`
    AllowedChannels []string `json:"allowedChannels"`
    ResponseMode    string   `json:"responseMode"`
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

**Slack bot:**

```json
{
  "id": "uuid-v4",
  "name": "team-slack",
  "type": "slack",
  "token": "mgc_ddd...",
  "allowedAgents": ["agent-uuid-1"],
  "enabled": true,
  "config": {
    "slack": {
      "botToken": "xoxb-...",
      "appToken": "xapp-...",
      "allowedUsers": ["U01ABCDEF"],
      "allowedChannels": [],
      "responseMode": "mirror"
    }
  }
}
```

**Discord bot:**

```json
{
  "id": "uuid-v4",
  "name": "team-discord",
  "type": "discord",
  "token": "mgc_ddd...",
  "allowedAgents": ["agent-uuid-1"],
  "enabled": true,
  "config": {
    "discord": {
      "botToken": "MTIz...",
      "allowedUsers": ["123456789012345678"],
      "allowedChannels": [],
      "responseMode": "text"
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
server/clients/
├── provider.go          — Provider interface, Schema type alias
├── registry.go          — Global registry: Register(), ValidateConfig() with oneOf
├── executor.go          — Shared execution logic (webhook + cron)
├── msgutil/
│   ├── msgutil.go       — Shared message validation and splitting utilities
│   └── msgutil_test.go  — Tests for validation and splitting
├── direct/spec.go       — Direct provider (empty schema)
├── telegram/
│   ├── spec.go          — Telegram provider (JSON Schema with x-format, enum)
│   └── bot.go           — Telegram bot (long polling, voice, emoji reactions)
├── discord/
│   ├── spec.go          — Discord provider (JSON Schema with x-format, enum)
│   └── bot.go           — Discord bot (Gateway WebSocket, voice, reactions, ! commands)
├── slack/
│   ├── spec.go          — Slack provider (JSON Schema with x-format, enum, array)
│   └── bot.go           — Slack bot (Socket Mode, audio clips, ! commands)
├── cron/
│   ├── spec.go          — Cron provider (JSON Schema with x-entity)
│   ├── cron.go          — Cron runtime
│   └── scheduler.go     — Cron scheduler
└── webhook/
    ├── spec.go          — Webhook provider (JSON Schema with oneOf branches)
    └── handler.go       — Webhook HTTP handler
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
        "botToken": {
          "type": "string",
          "x-format": "password",
          "x-placeholder": "123456:ABC-DEF..."
        },
        "allowedUsers": {
          "type": "string",
          "x-placeholder": "Comma-separated user IDs"
        },
        "allowedChats": {
          "type": "string",
          "x-placeholder": "Comma-separated chat IDs"
        },
        "responseMode": {
          "type": "string",
          "enum": ["text", "voice", "mirror", "both"],
          "default": "text"
        }
      }
    }
  },
  {
    "type": "slack",
    "displayName": "Slack",
    "configSchema": {
      "type": "object",
      "required": ["botToken", "appToken"],
      "properties": {
        "botToken": {
          "type": "string",
          "x-format": "password",
          "x-placeholder": "xoxb-..."
        },
        "appToken": {
          "type": "string",
          "x-format": "password",
          "x-placeholder": "xapp-..."
        },
        "allowedUsers": {
          "type": "array",
          "items": { "type": "string" },
          "x-placeholder": "Comma-separated Slack user IDs"
        },
        "allowedChannels": {
          "type": "array",
          "items": { "type": "string" },
          "x-placeholder": "Comma-separated Slack channel IDs"
        },
        "responseMode": {
          "type": "string",
          "enum": ["text", "voice", "mirror", "both"],
          "default": "text"
        }
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
        "schedule": { "type": "string", "x-placeholder": "0 8 * * *" },
        "commandId": { "type": "string", "x-entity": "commands" }
      }
    }
  },
  {
    "type": "webhook",
    "displayName": "Webhook",
    "configSchema": {
      "type": "object",
      "properties": {
        "passthrough": { "type": "boolean", "default": false },
        "commandId": { "type": "string", "x-entity": "commands" }
      },
      "oneOf": [
        {
          "properties": { "passthrough": { "const": false } },
          "required": ["commandId"]
        },
        {
          "properties": { "passthrough": { "const": true } }
        }
      ]
    }
  }
]
```

### JSON Schema Extensions

| Extension       | Purpose                                                           | Example                        |
| --------------- | ----------------------------------------------------------------- | ------------------------------ |
| `x-entity`      | UI renders a `<select>` populated from the named store collection | `"x-entity": "commands"`       |
| `x-format`      | UI renders password input instead of text                         | `"x-format": "password"`       |
| `x-placeholder` | Placeholder text for input fields                                 | `"x-placeholder": "0 8 * * *"` |

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

| Method | Path                             | Description                                         |
| ------ | -------------------------------- | --------------------------------------------------- |
| GET    | `/clients`                       | List all clients                                    |
| POST   | `/clients`                       | Create a client (token auto-generated as `mgc_...`) |
| GET    | `/clients/types`                 | List registered client types with JSON schemas      |
| GET    | `/clients/{id}`                  | Get a client by ID                                  |
| PUT    | `/clients/{id}`                  | Update a client                                     |
| DELETE | `/clients/{id}`                  | Delete a client                                     |
| POST   | `/clients/{id}/regenerate-token` | Regenerate auth token                               |

## Webhook Endpoint (User API)

| Method | Path                          | Description           |
| ------ | ----------------------------- | --------------------- |
| POST   | `/api/v1/webhooks/{clientId}` | Fire a webhook client |

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
- **Slack uses Socket Mode** — WebSocket, no public URL needed. Two tokens: `xoxb-` (Bot) + `xapp-` (App)
- **Slack commands use `!` prefix** — no slash commands (would need HTTP endpoint). IRC-style: `!help`, `!agent`, `!responsemode`
- **Voice API routes always registered** — STT/TTS proxy endpoints (`/api/v1/voice/`) are available regardless of Voice UI toggle, since Slack/Telegram clients need them

## Migration Chain

On `loadFromDisk()`, these migrations run in order (all idempotent):

1. `devices → clients` — Legacy Device entities become type `direct` clients
2. `cronJobs → triggers` — Legacy CronJob becomes Command + Trigger
3. `triggers → clients` — Trigger entities become cron/webhook client types with own tokens
4. `device → direct` — Client type `device` renamed to `direct`
5. `migrateIDs` — Generates UUID v4 for any entity missing one

## Future (out of scope)

- ~~**Discord provider**~~ ✅ — Implemented. See `server/clients/discord/`
- **WhatsApp provider** — `server/clients/whatsapp/`
- **Memory provider migration to JSON Schema** — `server/memory/` still uses FieldSpec
- **Enrollment** — `open` / `closed` / `approval` modes for user self-registration

## Multimodal File Support (implemented)

### Overview

Telegram, Slack and Discord clients accept user-uploaded files and forward them to the ADK `/run_sse` request. The adapter processes them using helpers in `server/clients/msgutil/attachments.go`:

- **Artifact** — files are persisted through the ADK `artifact.Service` with a session-unique filename (`{messageID}_{originalName}`). The prompt is amended with a `MAGEC_ATTACHED_ARTIFACTS` HTML-comment block listing each file (name, MIME, human size) and instructing the model to call `load_artifact` on demand.

See decision #24 in `.agents/DECISIONS.md` for the rationale and "do not" rules.

### Helper primitives (`server/clients/msgutil/attachments.go`)

- `StoreAsArtifact(ctx, svc, appName, userID, sessionID, filename, mimeType, data) (descriptor, error)` — persists via `artifact.Service.Save` and returns a single descriptor line meant to be wrapped by `AttachedArtifactsBlock`.
- `AttachedArtifactsBlock(lines []string) string` — wraps descriptor lines in the `MAGEC_ATTACHED_ARTIFACTS` block; returns `""` on empty input so callers can concatenate unconditionally.

Each client has its own short loop over platform-specific attachment types (`msg.Photo/Document/...`, `ev.Message.Files`, `m.Attachments`) because download APIs, size caps, and MIME discovery differ per platform. The DRY surface is the helper set, not the loop.

### Flow

```
User sends file in Telegram/Slack/Discord
  → Client downloads bytes from platform API (per-platform size cap: 20–25 MiB)
  → For each file:
      descriptor := StoreAsArtifact(ctx, artifacts, agentID, userID, sessionID, name, mime, data)
      artifactLines = append(artifactLines, descriptor)
  → prompt += AttachedArtifactsBlock(artifactLines)
  → POST /run_sse with {parts: [{text: prompt}]}
  → LLM calls load_artifact when it needs to read the file contents
```

### Fallback

When `artifacts == nil` (no agents configured yet, or intentionally unset), the file is dropped with a warning. When a `StoreAsArtifact` call fails (disk error, service down), the client logs a warning and drops that file — the text part of the user message is still delivered.

### Platform specifics

**Telegram**: `Photo` (highest-resolution), `Document`, `Video`, `Audio` (non-voice), `Animation`, `VideoNote`, `Sticker`. All resolve via `bot.GetFile(FileID)` + HTTP GET. `Caption` becomes the text part. Voice messages keep their own STT path (`handleVoice`).

**Slack (DMs)**: every non-audio entry in `ev.Message.Files`. Downloaded via `URLPrivateDownload` (fallback: `URLPrivate`) with Bearer auth. `AppMentionEvent` does not surface files in slack-go v0.17.3; channel-mention files remain a pending item in `TODO.md`.

**Discord**: every non-audio `m.Attachments[i]`. Downloaded via `att.URL`. Voice messages flow through `handleVoice`. `m.Content` with stripped mention becomes the text part.

### Parts serialization

Clients construct parts as raw `map[string]interface{}` (not `genai` types) since they serialize directly to JSON for the HTTP call to ADK:

```go
parts := []interface{}{
    map[string]string{"text": fullMessage},
}
```

### A2A (future)

A2A agent cards (`server/a2a/handler.go`) currently declare `DefaultInputModes: []string{"text/plain"}`. When A2A file support is added, include additional MIME types (`image/*`, `application/pdf`, etc.) and convert A2A `FilePart` → `genai.Part{InlineData}` in the executor.

## Message Size Handling (Implemented)

### Package: `server/clients/msgutil/`

Shared utility package for message validation and splitting. Both Telegram and Slack clients import it. Clients stay decoupled — each calls the utility with its own platform limits.

### Constants

| Constant                   | Value | Usage                             |
| -------------------------- | ----- | --------------------------------- |
| `TelegramMaxMessageLength` | 4096  | Telegram API limit per message    |
| `SlackMaxMessageLength`    | 39000 | Slack API limit per message block |
| `DefaultMaxInputLength`    | 16000 | Max inbound user message length   |

### Functions

**`ValidateInputLength(text string, maxLen int) (string, bool)`**

- Truncates at `maxLen` runes (unicode-safe), appends `\n\n[message truncated]`
- Returns the (possibly truncated) text and whether truncation occurred
- Applied at client entry points before calling the agent

**`SplitMessage(text string, maxLen int) []string`**

- Splits into chunks respecting `maxLen` runes per chunk
- Split priority: paragraph (`\n\n`) > line (`\n`) > word (space) > hard cut
- Returns `[]string` — all chunks non-empty, within limit

### Where it's applied

| Client       | Inbound                                                                            | Outbound                                                                          |
| ------------ | ---------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| **Telegram** | `handleMessage()` validates `msg.Text`; `handleVoice()` validates transcribed text | `sendResponse()` splits via `SplitMessage(text, 4096)`, sends chunks sequentially |
| **Discord**  | `handleTextMessage()` validates text; `handleVoice()` validates transcribed text   | `sendResponse()` splits via `SplitMessage(text, 2000)`, sends chunks sequentially |
| **Slack**    | `processMessage()` validates text (covers DMs + audio clips)                       | `postMessage()` splits via `SplitMessage(text, 39000)`, posts chunks sequentially |
| **Voice UI** | No validation (browser input is bounded)                                           | No splitting (browser has no render limit)                                        |
| **Executor** | No validation (prompts are from commands/webhooks, admin-controlled)               | No splitting (returns string to HTTP caller)                                      |

---

## Artifact Delivery

When an LLM uses `save_artifact` during a `/run` call, clients automatically deliver the new artifact as a file attachment. The flow:

1. **Before `/run`**: Client calls `GET /apps/{agent}/users/{user}/sessions/{session}/artifacts` to snapshot existing artifact names
2. **After `/run`**: Client calls the same endpoint again and diffs the two lists
3. **New artifacts**: Each new name is downloaded via `GET .../artifacts/{name}` and sent as a file

### Delivery per client

| Client       | Method                                                | Details                                            |
| ------------ | ----------------------------------------------------- | -------------------------------------------------- |
| **Telegram** | `ctx.Bot().SendDocument()` with `tu.FileFromReader()` | Artifact name used as filename                     |
| **Discord**  | `s.ChannelMessageSendComplex()` with `discordgo.File` | Artifact name used as filename                     |
| **Slack**    | `c.api.UploadFileV2()`                                | Artifact name as filename + title, respects thread |
| **Voice UI** | Not yet implemented                                   | Would need download button in UI                   |

### Artifact REST response format

The ADK artifact endpoint returns a `genai.Part` JSON:

- **Text artifacts**: `{"text": "content..."}`
- **Binary artifacts**: `{"inlineData": {"mimeType": "...", "data": "<base64>"}}`

### Key files

| File                                      | Role                                                                            |
| ----------------------------------------- | ------------------------------------------------------------------------------- |
| `server/agent/tools/artifacts/toolset.go` | Toolset with save/load/list tools                                               |
| `server/agent/base_toolset.go`            | Wires artifact toolset into all agents                                          |
| `server/agent/agent.go`                   | Creates `artifactfs.NewFilesystemService()`, sets `launcherCfg.ArtifactService` |
| `server/clients/telegram/bot.go`          | `listArtifacts()`, `downloadArtifact()`, `sendNewArtifacts()`                   |
| `server/clients/slack/bot.go`             | Same three methods, adapted for Slack API                                       |
