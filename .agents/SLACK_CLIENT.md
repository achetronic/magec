# Slack Client — Design Document

## Overview

Slack client for Magec using **Socket Mode** (WebSocket-based, no public URL needed). Users can DM the bot or mention it in channels. Supports text messages, audio clips (voice), response modes, and per-channel agent switching. Follows the same pattern as the Telegram client.

## Connection Mode: Socket Mode

Socket Mode uses an outbound WebSocket connection from Magec to Slack's servers. No webhook URL, no public endpoint, no ingress configuration needed. Works behind NAT, firewalls, etc.

**Two tokens required:**

| Token | Prefix | Purpose |
|-------|--------|---------|
| **Bot Token** | `xoxb-` | Operational: send/read messages, manage channels, download/upload files |
| **App Token** | `xapp-` | Connection: establish the Socket Mode WebSocket tunnel |

Both are configured in the client's config. The App Token only opens the tunnel — it cannot read or send messages.

## Slack App Setup (user steps)

1. Go to [api.slack.com/apps](https://api.slack.com/apps) → **Create New App** (from scratch)
2. **Socket Mode** → Enable → generates `xapp-` token (name it e.g. "magec-socket")
3. **OAuth & Permissions** → Add Bot Token Scopes:
   - `chat:write` — send messages
   - `app_mentions:read` — receive @mentions in channels
   - `files:read` — access audio clips and file attachments
   - `files:write` — upload voice response files
   - `im:history` — read DMs
   - `im:read` — access DM conversations
   - `im:write` — respond in DMs
   - `users:read` — look up user names and display names
4. **Event Subscriptions** → Enable → Subscribe to bot events:
   - `message.im` — DM messages
   - `app_mention` — @mentions in channels
5. **App Home** → Enable:
   - **"Allow users to send Slash commands and messages from the messages tab"** — required for DMs
   - **"Show My Bot as Online"** — shows green presence dot when connected
6. **Install to Workspace** → generates `xoxb-` token
7. Paste both tokens in Magec Admin UI → Clients → New Slack client

## Data Model

### SlackClientConfig (store/types.go)

```go
type SlackClientConfig struct {
    BotToken        string   `json:"botToken,omitempty"`
    AppToken        string   `json:"appToken,omitempty"`
    AllowedUsers    []string `json:"allowedUsers,omitempty"`    // Slack user IDs (e.g. "U01ABCDEF")
    AllowedChannels []string `json:"allowedChannels,omitempty"` // Slack channel IDs (e.g. "C01ABCDEF")
    ResponseMode    string   `json:"responseMode,omitempty"`    // "text" (default), "voice", "mirror", "both"
}
```

**Note**: Slack user/channel IDs are strings (e.g. `U01ABCDEF`, `C01ABCDEF`), not integers like Telegram.

### JSON Schema (spec.go)

```
botToken:        required, x-format: password, x-placeholder: "xoxb-..."
appToken:        required, x-format: password, x-placeholder: "xapp-..."
allowedUsers:    optional, array of strings, x-placeholder: "Comma-separated Slack user IDs"
allowedChannels: optional, array of strings, x-placeholder: "Comma-separated Slack channel IDs"
responseMode:    optional, enum: ["text", "voice", "mirror", "both"], default: "text"
```

## Architecture

```
server/clients/slack/
├── spec.go    — Provider interface + JSON Schema (same pattern as telegram/spec.go)
└── bot.go     — Socket Mode client lifecycle + message handling + voice support
```

### Bot lifecycle

- **`New(clientDef, agentURL, agents, logger)`** — validates config, creates `slack.Client` + `socketmode.Client`
- **`Start(ctx)`** — runs `socketmode.Client.RunContext(ctx)` in a goroutine, listens on `client.Events` channel for:
  - `EventTypeEventsAPI` → `message.im` (DMs) and `app_mention` (channel mentions)
  - Always `Ack()` every event
- **`Stop()`** — cancels context, closes WebSocket

### Message handling

1. Receive event → extract user ID, channel ID, text
2. Permission check: `isAllowed(userID, channelID)`
3. Check for audio clips in `ev.Message.Files` (mimetype `audio/*`)
4. Check for bot commands: `!help`, `!agent`, `!reset`, `!responsemode`
5. Build MAGEC_META with Slack context
6. Call agent via internal HTTP API (`/api/v1/agent/run_sse`)
7. Respond based on response mode (text, voice, or both)

### Emoji reactions

Same pattern as Telegram: `eyes` (received) → `brain` (thinking) → `white_check_mark` (success) or `x` (failure).

### Progress timeout

30-second timer sends "Still working on it..." if the agent hasn't responded yet. Same pattern as Telegram.

### Error sanitization

Redacts `xoxb-` and `xapp-` tokens from error messages before showing to users.

### Voice message handling

**Incoming (audio clips → STT):**
1. Detect files with `audio/*` mimetype in `ev.Message.Files`
2. Download via `url_private_download` with Bearer token auth
3. Convert WebM → WAV using ffmpeg (`-ar 16000 -ac 1 -f wav`)
4. POST to `/api/v1/voice/{agentId}/transcription` (multipart)
5. Process transcribed text as a normal message (with `inputWasVoice=true`)

**Outgoing (TTS → file upload):**
1. POST to `/api/v1/voice/{agentId}/speech` with `{"input": text, "response_format": "opus"}`
2. Upload audio as `voice.ogg` via `UploadFileV2` to the channel
3. Falls back to text if TTS fails

### Response modes

Same as Telegram:

| Mode | Behavior |
|------|----------|
| `text` | Text only (default) |
| `voice` | TTS voice file only |
| `mirror` | Voice in → voice out, text in → text out |
| `both` | Both text + voice file |

Runtime override via `responsemode <mode>` command. `responsemode reset` reverts to config default.

### MAGEC_META fields

```json
{
    "source": "slack",
    "slack_user_id": "U01ABCDEF",
    "slack_channel_id": "C01ABCDEF",
    "slack_channel_type": "im|channel|group",
    "slack_username": "john.doe",
    "slack_name": "John Doe",
    "slack_email": "john.doe@example.com",
    "slack_team_id": "T01ABCDEF",
    "slack_thread_ts": "1234567890.123456"
}
```

### Thread support

When a user mentions the bot in a channel, the response is posted as a **thread reply** (using `slack.MsgOptionTS(threadTS)`) to avoid cluttering the channel. Voice file uploads also use `ThreadTimestamp` for thread replies. In DMs, replies are direct messages.

### Session ID strategy

`slack_{channelID}_{agentID}` — same pattern as Telegram (`telegram_{chatID}_{agentID}`).

## Bot commands (DMs only)

| Command | Description |
|---------|-------------|
| `!help` | Show available commands |
| `!agent` | Show/switch active agent |
| `!agent <id>` | Switch to specific agent |
| `!reset` | Reset session (delete ADK session, start fresh) |
| `!responsemode` | Show current response mode |
| `!responsemode <mode>` | Set response mode (text/voice/mirror/both/reset) |

## Wiring (main.go)

Same pattern as Telegram:

1. Blank import `_ "github.com/achetronic/magec/server/clients/slack"` for registry
2. Explicit import for `slack` package to create clients
3. Boot loop: iterate `dataStore.ListClients()`, filter `type == "slack"`, create + start in goroutine
4. Shutdown: call `Stop()` on each

## Dependencies

- `github.com/slack-go/slack` v0.17.3 — Slack API + Socket Mode client
- `ffmpeg` — required at runtime for WebM → WAV audio conversion (same as Telegram's OGG → WAV)

## Differences from Telegram

| Aspect | Telegram | Slack |
|--------|----------|-------|
| Connection | Long polling | Socket Mode (WebSocket) |
| Tokens | 1 (botToken) | 2 (botToken + appToken) |
| User IDs | int64 | string |
| Voice input | Native voice messages (OGG) | Audio clips (WebM) |
| Voice output | Native voice message | File upload (Opus) |
| Response modes | text/voice/mirror/both | text/voice/mirror/both |
| Thread support | Not applicable | Reply in thread for channel mentions |
| Commands | /start, /help, /agent, /responsemode | help, agent, responsemode (plain text) |

## Out of scope

- **Slash commands** — would require HTTP endpoint, defeats Socket Mode purpose
- **Interactive components** (buttons, modals) — future enhancement
- **Video clips** — Slack has video clips but STT doesn't apply; future enhancement

## Multimodal File Support (pending implementation)

Files from Slack users will be sent as `inlineData` parts in the ADK `/run` request. See `CLIENT_DESIGN.md` for full design.

**Changes needed in `bot.go`**:
- `handleAudioClip()` currently only processes `audio/*` files. Rename/refactor to `handleFiles()` and add handling for `image/*`, `application/pdf`, and generic mimetypes.
- For non-audio files: download via `downloadSlackFile()`, encode base64, add as `inlineData` part.
- `processMessage()`: change parts from `[]map[string]string` to `[]interface{}` to support mixed text+inlineData.
- For `handleAppMention`: files may arrive differently than in DMs — verify `ev.Message.Files` availability in mention events.
- File size: check `file.Size` < 20MB before downloading.
- `ev.Text` alongside files becomes the caption/text part.
