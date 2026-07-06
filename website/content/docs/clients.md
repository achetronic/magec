---
title: "Clients"
---

Clients are how users and systems interact with your agents. Every conversation in Magec happens through a client, whether it's a person talking through the Voice UI, a Telegram or Discord bot answering messages, or a cron job running a task on a schedule.


<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-clients.png" alt="Clients list" >}}
</div>

## How clients work

Every client in Magec follows the same pattern:

1. **Authentication.** Each client gets a unique token (prefixed with `mgc_`) that authenticates it against the API. The token is generated automatically when you create the client.
2. **Authorization.** Each client has a list of allowed agents and flows. It can only interact with the ones you've explicitly permitted.
3. **Transport.** Each client type handles its own communication channel. The Voice UI calls the REST API, Telegram polls the Bot API, Slack and Discord connect via WebSocket, webhooks listen for HTTP requests, cron fires on a schedule.
4. **Execution.** All clients end up in the same place: sending a prompt to an agent (or flow) and returning the response through their own channel.

This design means you control exactly who can access what. A Voice UI client for the front desk might have access to a customer service agent only. A Telegram bot for your team might have access to all agents and flows. A cron job might only run a specific daily report.

## Client types

| Type | What it does | Use cases |
|------|-------------|-----------|
| **Direct** | Browser-based access via the REST API. Powers the Voice UI and any custom integration. | Voice interface, custom web apps, API integrations |
| **Telegram** | Connects a Telegram bot. Users send text or voice messages and get responses. | Mobile assistant, team bot, customer support |
| **Slack** | Connects a Slack bot via Socket Mode. Users DM the bot or @mention it in channels. | Team workspace assistant, internal tools, ops bot |
| **Discord** | Connects a Discord bot via Gateway WebSocket. Users DM the bot or @mention it in channels. | Community assistant, server bot, moderation helper |
| **Webhook** | Exposes an HTTP endpoint that triggers agent invocations. | CI/CD integration, form processing, alert handling, external automation |
| **Cron** | Runs commands on a schedule, like a cron job that talks to your agents. | Daily reports, periodic health checks, scheduled maintenance |

Each type is covered in detail on its own page:

- [Voice UI](/docs/voice-ui/). Browser-based voice interface with wake word, push-to-talk, agent switching, conversation history, and PWA support
- [Telegram](/docs/telegram/). Bot with text and voice messages, response modes, per-chat agent switching, and user restrictions
- [Slack](/docs/slack/). Bot with Socket Mode, text and voice messages, per-channel agent switching, and thread replies
- [Discord](/docs/discord/). Bot with Gateway WebSocket, text and voice messages, per-channel agent switching, and @mention replies
- [Webhooks](/docs/webhooks/). HTTP endpoint for external system integrations with command and passthrough modes
- [Cron](/docs/cron/). Scheduled tasks that run commands against agents on a configurable schedule

## Creating a client

1. Open the Admin UI at `http://localhost:8081`
2. Go to **Clients** and click **New**
3. Choose a type
4. Give it a name
5. Select which agents and flows it can access
6. Fill in any type-specific settings (Telegram bot token, cron schedule, etc.)
7. Save

Magec generates the authentication token automatically. For Direct and Telegram clients, you'll use this token to connect. For Slack and Discord, authentication uses their own bot tokens. For webhooks, you include it in the `Authorization` header. For cron, authentication is handled internally.

## Token management

Each client has a unique `mgc_` token. If a token is compromised, you can regenerate it from the Admin UI. The old token stops working immediately and a new one is issued. This doesn't affect the client's configuration, just its authentication credential.

## Multiple clients, same agents

You can create as many clients as you need, and multiple clients can access the same agents. For example:

- A **Voice UI client** for the front desk with access to the Customer Service agent
- A **Telegram client** for your team with access to all agents
- A **Slack client** for your workspace with access to work agents
- A **Discord client** for your community server with access to help agents
- A **Webhook client** for your CI pipeline with access to the Code Review agent
- A **Cron client** for daily reports with access to the Analytics flow

Each client has its own token, its own allowed agents, and its own conversation history. They don't interfere with each other.

{{< callout type="info" >}}
When no clients exist (fresh install without a seed), the API runs in open mode: no authentication required. This makes initial setup easier. As soon as you create your first client, authentication is enforced.
{{< /callout >}}
