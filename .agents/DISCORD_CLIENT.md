# Discord Client — Design Document

## Overview

Discord client for Magec using **Gateway WebSocket** (outbound connection, no public URL needed). Users can DM the bot or @mention it in server channels. Supports text messages, voice messages, response modes, and per-channel agent switching. Follows the same pattern as the Telegram and Slack clients.

## Connection Mode: Gateway WebSocket

The bot opens an outbound WebSocket connection to Discord's Gateway API via `discordgo.Session.Open()`. No webhook URL, no public endpoint, no ingress configuration needed. Works behind NAT, firewalls, etc.

**Single token required:**

| Token | Purpose |
|-------|---------|
| **Bot Token** | Everything: establish Gateway connection, send/read messages, manage reactions, upload files |

Simpler than Slack (which needs two tokens). Similar to Telegram (one token), but uses persistent WebSocket instead of long polling.

## Discord App Setup (user steps)

1. Go to [discord.com/developers/applications](https://discord.com/developers/applications) → **New Application**
2. **Bot** section → **Reset Token** → copy the bot token
3. **Bot** → **Privileged Gateway Intents** → enable **Message Content Intent**
4. **OAuth2** → **URL Generator** → select `bot` scope → select permissions (or use pre-built URL below)
5. Open invite URL → select server → authorize
6. Paste bot token in Magec Admin UI → Clients → New Discord client

### OAuth2 Invite URL

```
https://discord.com/oauth2/authorize?client_id=YOUR_APP_ID&permissions=70643622186048&scope=bot
```

### Required Permissions

| Permission | Bit | Value |
|------------|-----|-------|
| Send Messages | `1 << 11` | `2048` |
| Read Message History | `1 << 16` | `65536` |
| Add Reactions | `1 << 6` | `64` |
| Attach Files | `1 << 15` | `32768` |
| Send Messages in Threads | `1 << 38` | `274877906944` |
| Send Voice Messages | `1 << 46` | `70368744177664` |
| **Total** | | **`70643622186048`** |

### Gateway Intents

```go
discordgo.IntentGuildMessages |
discordgo.IntentDirectMessages |
discordgo.IntentMessageContent |        // PRIVILEGED — must enable in Developer Portal
discordgo.IntentGuildMessageReactions |
discordgo.IntentDirectMessageReactions
```

`IntentGuilds` is intentionally omitted. Thread/channel metadata is resolved on demand via `s.Channel()` (HTTP fallback) rather than relying on the state cache being fully populated.

**Message Content Intent** is privileged. Must be toggled ON in Developer Portal → Bot → Privileged Gateway Intents. Without it, `m.Content` is always empty.

## Data Model

### DiscordClientConfig (store/types.go)

```go
type DiscordClientConfig struct {
    BotToken        string   `json:"botToken,omitempty"`
    AllowedUsers    []string `json:"allowedUsers,omitempty"`    // Discord user IDs (snowflakes)
    AllowedChannels []string `json:"allowedChannels,omitempty"` // Discord channel IDs (snowflakes)
    ResponseMode    string   `json:"responseMode,omitempty"`    // "text" (default), "voice", "mirror", "both"
}
```

**Note**: Discord IDs are snowflake strings (e.g. `"123456789012345678"`), not integers.

### JSON Schema (spec.go)

```
botToken:        required, x-format: password, x-placeholder: "MTIz..."
allowedUsers:    optional, array of strings, x-placeholder: "Comma-separated Discord user IDs"
allowedChannels: optional, array of strings, x-placeholder: "Comma-separated Discord channel IDs"
responseMode:    optional, enum: ["text", "voice", "mirror", "both"], default: "text"
```

## Architecture

```
server/clients/discord/
├── spec.go    — Provider interface + JSON Schema (same pattern as telegram/spec.go, slack/spec.go)
└── bot.go     — Gateway client lifecycle + message handling + voice support
```

### Bot lifecycle

- **`New(clientDef, agentURL, agents, logger)`** — validates config, creates `discordgo.Session` with bot token, registers `MessageCreate` handler, sets Gateway intents
- **`Start(ctx)`** — calls `session.Open()` to connect Gateway WebSocket, spawns goroutine waiting on context cancellation
- **`Stop()`** — cancels context, calls `session.Close()` to disconnect

### Message handling

1. Receive `MessageCreate` event → ignore bots (`m.Author.Bot`)
2. Determine DM vs guild: `isDM = (m.GuildID == "")`
3. In guild: check if bot is @mentioned in `m.Mentions`, strip mention from text
4. Permission check: `isAllowed(userID, channelID)`
5. Check for voice message flag (`m.Flags & discordgo.MessageFlagsIsVoiceMessage`) or audio attachment
6. Check for bot commands: `!help`, `!agent`, `!reset`, `!responsemode`
7. Add reactions: 👀 (received)
8. Build MAGEC_META with Discord context
9. Call agent via internal HTTP API (`/api/v1/agent/run_sse`)
10. Add 🧠 reaction (processing), remove 👀
11. Respond based on response mode (text, voice, or both)
12. Remove 🧠, add ✅ (success) or ❌ (failure)

### Emoji reactions

👀 (received) → 🧠 (thinking) → ✅ (success) or ❌ (failure). Intermediate reactions are cleaned up via `MessageReactionRemove` with `"@me"` as the user ID.

### Error sanitization

Redacts bot token from error messages before showing to users.

### Voice message handling

**Incoming (voice messages → STT):**
1. Detect voice via `m.Flags & discordgo.MessageFlagsIsVoiceMessage` or `audio/*` attachment with `DurationSecs > 0`
2. Download audio attachment via `att.URL`
3. Convert to WAV using ffmpeg (`-ar 16000 -ac 1 -f wav`)
4. POST to `/api/v1/voice/{agentId}/transcription` (multipart)
5. Process transcribed text as a normal message (with `inputWasVoice=true`)

**Outgoing (TTS → file upload):**
1. POST to `/api/v1/voice/{agentId}/speech` with `{"input": text, "response_format": "opus"}`
2. Upload audio as `voice.ogg` via `ChannelMessageSendComplex` with file attachment
3. Falls back to text if TTS fails

### Response modes

| Mode | Behavior |
|------|----------|
| `text` | Text only (default) |
| `voice` | TTS voice file only |
| `mirror` | Voice in → voice out, text in → text out |
| `both` | Both text + voice file |

Runtime override via `!responsemode <mode>`. `!responsemode reset` reverts to config default.

### MAGEC_META fields

```json
{
    "source": "discord",
    "discord_user_id": "123456789012345678",
    "discord_channel_id": "987654321098765432",
    "discord_channel_type": "guild|dm",
    "discord_username": "john",
    "discord_name": "John Doe",
    "discord_guild_id": "111222333444555666"
}
```

`discord_guild_id` only present in guild (server) channels.

### Message splitting

Uses `msgutil.SplitMessage(text, msgutil.DiscordMaxMessageLength)` with Discord's 2,000 character limit. First chunk gets the reply reference; subsequent chunks are standalone messages.

### Artifact delivery

1. Before agent call: snapshot existing artifact names via `GET /apps/{agentID}/users/default_user/sessions/{sessionID}/artifacts`
2. After agent response: list artifacts again
3. New artifacts (diff): download individually, send as Discord file attachments via `ChannelMessageSendComplex`
4. Supports both binary (`inlineData` with base64) and text artifacts

### Session ID strategy

`discord_{channelID}_{agentID}` — same pattern as Telegram (`telegram_{chatID}_{agentID}`) and Slack (`slack_{channelID}_{agentID}`).

## Bot commands

| Command | Description |
|---------|-------------|
| `!help` | Show available commands |
| `!agent` | Show/switch active agent |
| `!agent <id>` | Switch to specific agent |
| `!reset` | Reset session (delete ADK session, start fresh) |
| `!responsemode` | Show current response mode |
| `!responsemode <mode>` | Set response mode (text/voice/mirror/both/reset) |

## Access control

```go
func (c *Client) isAllowed(userID string, channelIDs ...string) bool
```

- Both lists empty → open access
- Either list populated → OR allowlist: user allowed if in `AllowedUsers` OR any of the `channelIDs` in `AllowedChannels`
- Neither matches → denied (silently dropped, logged at debug level)

**Thread IDs are different from their parent channel ID.** When a message arrives inside a thread, `m.ChannelID` is the thread's own ID — not the parent channel. `onMessageCreate` resolves the parent ID via `s.Channel()` (state cache + HTTP fallback) and passes both to `isAllowed`, so that a thread created from an allowed channel is itself allowed.

## Wiring (main.go)

Same pattern as Telegram and Slack:

1. Blank import `_ "github.com/achetronic/magec/server/clients/discord"` for registry
2. Explicit import `discordclient "github.com/achetronic/magec/server/clients/discord"` for creating clients
3. Reconcile loop: filter `type == "discord"`, create + start in goroutine
4. `startDiscord()` method follows identical pattern to `startTelegram()` and `startSlack()`
5. Shutdown: call `Stop()` on each

## Dependencies

- `github.com/bwmarrin/discordgo` v0.29.0 — Discord Gateway API + REST client
- `ffmpeg` — required at runtime for audio conversion (same as Telegram and Slack)

## Differences from Telegram and Slack

| Aspect | Discord | Slack | Telegram |
|--------|---------|-------|----------|
| Connection | Gateway WebSocket | Socket Mode (WebSocket) | Long polling |
| Tokens | 1 (botToken) | 2 (botToken + appToken) | 1 (botToken) |
| User IDs | Snowflake strings | String IDs (`U01ABCDEF`) | int64 |
| Voice input | Native voice messages | Audio clips (M4A) | Native voice (OGG) |
| Voice output | `.ogg` file attachment | File upload (Opus) | Native voice message |
| Commands | `!` prefix | `!` prefix | `/` slash commands |
| Channel replies | @mention → reply | @mention → thread reply | Group messages |
| Message limit | 2,000 chars | 39,000 chars | 4,096 chars |
| Privileged setup | Message Content Intent | None | None |

## Out of scope

- **Slash commands** — would require HTTP interaction endpoint, defeats no-public-URL design
- **Interactive components** (buttons, select menus, modals) — future enhancement
- **Forum channels** — future enhancement
- **Voice channels (audio streaming)** — voice messages via text/DM channels only

## Multimodal File Support (pending implementation)

Files from Discord users will be sent as `inlineData` parts in the ADK `/run` request. See `CLIENT_DESIGN.md` for full design.

**Changes needed in `bot.go`**:
- Add handling for non-audio attachments (`image/*`, `application/pdf`, generic mimetypes)
- Download via attachment URL, encode base64, add as `inlineData` part
- File size: check `att.Size` < 20MB before downloading
- `m.Content` alongside attachments becomes the caption/text part
