---
title: "Clients"
---

Clients are how users and systems interact with your agents. Every conversation in Magec happens through a client — whether it's a person talking through the Voice UI, a Telegram bot answering messages, or a cron job running a task on a schedule.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-clients.png" alt="Admin UI — Clients list" >}}
</div>

Each client has a **type** that determines how it connects, a set of **allowed agents and flows** that it can interact with, and its own **token** for authentication.

## Client types

| Type | What it does |
|------|-------------|
| **Direct** | Browser-based access. Powers the Voice UI and any direct API call. |
| **Telegram** | Connects a Telegram bot. Users send text or voice messages and get responses. |
| **Webhook** | Receives HTTP requests and either runs a preconfigured command or accepts an external prompt via passthrough. |
| **Cron** | Runs commands on a schedule — like a cron job that talks to your agents. |

## How they work

When you create a client, you choose its type and select which agents and flows it's allowed to use. Magec generates a token that the client uses to authenticate against the API. From there, the client handles its own transport — the Voice UI calls the REST API (with a WebSocket for voice transcription and events), Telegram polls the Bot API, webhooks listen for HTTP requests, and cron fires on a schedule.

All clients end up in the same place: sending a prompt to an agent (or a flow) and returning the response through their own channel.

## Creating a client

Open the Admin UI, go to Clients, and click New. Pick a type, give it a name, select the agents and flows it should have access to, and fill in any type-specific configuration. Each client type has its own page in the docs with the details:

- [Voice UI](/magec/docs/voice-ui/) — Browser-based voice interface with animated orb, push-to-talk, and PWA support
- [Telegram](/magec/docs/telegram/) — Bot with text and voice messages, response modes, and user restrictions
- [Webhooks](/magec/docs/webhooks/) — HTTP endpoint for external system integrations
- [Cron](/magec/docs/cron/) — Scheduled tasks that run commands against agents
