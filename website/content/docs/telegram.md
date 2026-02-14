---
title: "Telegram"
---

Connect to Magec via Telegram — send text or voice messages and get responses.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-clients.png" alt="Admin UI — Clients" >}}
</div>

## Setup

1. Create a bot with [@BotFather](https://t.me/BotFather) and get the token
2. Get your user ID from [@userinfobot](https://t.me/userinfobot)
3. Create a Telegram client in the Admin UI with your bot token and allowed users
4. Start chatting with your bot

## Response modes

| Mode | Behavior |
|------|----------|
| `text` | Always reply with text (default) |
| `voice` | Always reply with voice (requires TTS) |
| `mirror` | Reply in the same format as input |
| `both` | Reply with both text and voice |

## Bot commands

- `/responsemode <mode>` — Change response mode at runtime
- `/agent` — Switch the active agent for this chat

{{< callout >}}
**Security:** Always set `allowedUsers` and/or `allowedChats` to restrict who can use your bot.
{{< /callout >}}
