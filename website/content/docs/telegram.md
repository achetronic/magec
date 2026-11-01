---
title: "Telegram"
---

Magec can connect to Telegram through a bot. Users send text or voice messages to the bot, and the bot responds using your configured agents. It supports multiple response modes, per-chat agent switching, and voice messages in both directions.

This is a great way to access your agents from anywhere (your phone, desktop, or any Telegram client) without opening the Voice UI or writing API calls.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-clients-telegram.png" alt="Telegram client" >}}
</div>

## Setup

### 1. Create a Telegram bot

Open Telegram and talk to [@BotFather](https://t.me/BotFather):

1. Send `/newbot`
2. Choose a name (e.g., "My Magec Assistant")
3. Choose a username (e.g., `my_magec_bot`)
4. BotFather gives you a token. Copy it.

### 2. Get your user ID

Talk to [@userinfobot](https://t.me/userinfobot). It replies with your Telegram user ID (a number). You'll need this to restrict who can use your bot.

### 3. Create a Telegram client in Magec

In the Admin UI, go to **Clients** → **New** → **Telegram**:

| Field | Description |
|-------|-------------|
| `name` | Display name for this client |
| `botToken` | The token from BotFather |
| `allowedUsers` | Comma-separated list of Telegram user IDs that can use this bot |
| `allowedChats` | Comma-separated list of chat IDs where the bot can respond (for group chats) |
| `groupPolicy` | When the bot answers in a group (see below) |
| `responseMode` | How the bot responds (see below) |
| `allowedAgents` | Which agents and flows this bot can access |

### 4. Start chatting

Open your bot in Telegram and send a message. The bot responds using the first allowed agent.

## Group policy

In a private chat the bot answers every message. In a group or
supergroup, `groupPolicy` decides when it speaks:

| Policy | Behavior |
|--------|----------|
| `mention` | Answer only when the bot is named, for example `@mybot what is the deadline?`. This is the default, and the behaviour of every version before this setting existed. |
| `all` | Answer every message in the group. Useful in a small working group where the bot is a participant rather than a tool. |
| `observe` | Never answer in a group. The messages are still received and kept as context, so the bot can be asked later — in a private chat, or by an agent — what was said. |

`observe` exists because receiving is not answering. A bot that reads a
busy group and comments on every line becomes unbearable; the same bot,
silent, can still summarise the discussion when someone asks for it.

An unrecognised value falls back to `mention` and logs a warning. A bot
that speaks because its configuration was not understood is worse than
one that stays quiet.

Two notes if group messages do not reach the bot at all:

- Telegram's privacy mode must be disabled in BotFather
  (`/setprivacy` → `Disable`), otherwise the bot only receives messages
  that start with `/` or name it — regardless of `groupPolicy`.
- `allowedChats`, when set, still applies. `groupPolicy` decides *when*
  to answer in a chat the bot is already allowed to use; it does not
  grant access to a chat.

## Response modes

The response mode controls the format of the bot's replies:

| Mode | Behavior |
|------|----------|
| `text` | Always respond with text (default). Simple and reliable. |
| `voice` | Always respond with a voice message. Requires the agent to have TTS configured. |
| `mirror` | Mirror the user's format: text replies to text, voice replies to voice messages. |
| `both` | Respond with both a text message and a voice message. |

Users can change the response mode at runtime with the `/responsemode` command.

## Voice messages

Telegram voice messages work in both directions:

**Incoming voice:** When a user sends a voice message, Magec:
1. Downloads the audio from Telegram
2. Converts it from OGG to WAV (using ffmpeg inside the container)
3. Sends it to the agent's STT backend for transcription
4. The agent processes the transcribed text and responds

**Outgoing voice:** When the response mode requires voice, Magec:
1. Sends the agent's text response to the TTS backend
2. Gets the audio back
3. Sends it as a Telegram voice message

This means you can have a fully voice-based conversation through Telegram: speak a question, hear the answer.

## Bot commands

| Command | Description |
|---------|-------------|
| `/start` | Welcome message |
| `/help` | List available commands |
| `/agent` | Switch the active agent for this chat |
| `/responsemode <mode>` | Change the response mode (`text`, `voice`, `mirror`, `both`, `reset`) |

The `/agent` command shows a list of all agents and flows the bot has access to. Select one and all subsequent messages in that chat go to the selected agent. Different chats can use different agents simultaneously.

## Context metadata

When a message arrives from Telegram, Magec injects metadata about the source: the Telegram user ID, username, display name, and chat ID. You can use this in system prompts to personalize responses (e.g., *"Address the user by their first name when available"*).

## Security

{{< callout >}}
**Always restrict access.** Set `allowedUsers` and/or `allowedChats` to control who can use your bot. Without restrictions, anyone who discovers your bot's username can interact with your agents, and through them, any MCP tools those agents have access to.
{{< /callout >}}

If a user not in the allowed list tries to message the bot, the message is silently ignored. The same applies for chats not in the allowed chats list.

## Multiple bots

You can create multiple Telegram clients, each with its own bot, its own set of allowed users, and its own agent access. For example:

- A **personal bot** with full access to all agents, restricted to your user ID
- A **team bot** with access to specific work agents, restricted to team member IDs
- A **customer bot** with access to a customer service agent, available to specific chat groups
