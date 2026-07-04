---
title: "A2A Protocol"
---

A2A (Agent-to-Agent) lets your Magec agents talk to external AI systems — and lets external systems talk to yours. When you enable A2A on an agent, it becomes discoverable and callable by any application that speaks the [A2A protocol](https://a2a-protocol.org/latest/specification/).

In practice: you flip a toggle, and your agent gets a URL that other tools can connect to. They discover what the agent can do, authenticate, and send it messages — all without you writing any code.

## Why would I use this?

Think of A2A as giving your agent a phone number that other AI systems can call.

- **Connect agents across platforms.** An agent running in another tool (Claude Desktop, a custom app, another Magec instance) can discover and invoke your agent directly.
- **Build multi-agent systems.** A coordinator agent somewhere else can delegate tasks to your specialized agents — your "database expert" handles queries, your "email writer" drafts responses, each on their own Magec instance.
- **Expose agents as services.** Instead of building a custom API for each agent, A2A gives you a standard protocol that any compatible client already knows how to use.

If you're running Magec for personal use or within a single app, you probably don't need A2A. It becomes useful when you want agents to collaborate across different systems.

## How it works

When you enable A2A on an agent, Magec does three things:

1. **Publishes an agent card** — a JSON document describing the agent's name, capabilities, skills, and how to authenticate. This is the discovery mechanism.
2. **Exposes a JSON-RPC endpoint** — the URL where clients send messages to the agent. Supports both synchronous and streaming responses.
3. **Auto-generates skills from the agent's tools** — MCP servers, built-in tools, and the agent's instructions are all reflected in the card so clients know what the agent can do.

You don't configure any of this manually. The card and skills are derived from the agent's existing setup.

## Enabling A2A

### Agents

1. Open an agent in the Admin UI
2. Toggle **A2A Protocol** on
3. Save

### Flows

Agentic flows can also be exposed via A2A. The process is the same: open the flow in the editor, enable the A2A toggle, and save.

The external client sees the flow as a single agent. The internal orchestration (which steps run in sequence, which run in parallel) is completely transparent. The agent card auto-generates skills that describe the flow's sub-agents and their capabilities, so the client understands what the flow can do without knowing how it's structured internally.

## Endpoints

| Endpoint | Auth | Description |
|----------|------|-------------|
| `/api/v1/a2a/.well-known/agent-card.json` | No | Lists all A2A-enabled agents |
| `/api/v1/a2a/{agentId}/.well-known/agent-card.json` | No | Agent card for a specific agent |
| `/api/v1/a2a/{agentId}` | Bearer token | JSON-RPC invocation endpoint |

Discovery endpoints are public so that external clients can find your agents. The invocation endpoint requires a **Bearer token** — this is a regular Magec client token, the same ones you create in the Admin UI under Clients.

## Connecting a client

Give the external A2A client this URL:

```
https://your-server/api/v1/a2a/{agentId}
```

The client will automatically fetch the agent card from the `.well-known` path relative to that URL, read the security requirements, and use the Bearer token you provide to send messages.

## Public URL

By default, agent cards reference `http://localhost:{port}` as the agent's URL. This works for local development but not when your server is behind a reverse proxy or exposed to the internet.

Set `publicURL` in your config so that agent cards contain the correct address:

```yaml
server:
  publicURL: https://magec.example.com
```

{{< callout type="info" >}}
If you don't set `publicURL`, everything still works locally. You only need it when external systems need to reach your server from outside.
{{< /callout >}}

## What's in an agent card?

The agent card is what clients read to understand your agent. Here's what it contains:

| Field | Description |
|-------|-------------|
| `name` | Agent's display name |
| `description` | What the agent does |
| `url` | JSON-RPC endpoint for invocation |
| `skills` | Auto-generated list of capabilities (from system prompt + tools) |
| `securitySchemes` | How to authenticate (Bearer token) |
| `capabilities` | Protocol features supported (streaming) |

Skills are generated automatically from the agent's configuration. Each MCP tool the agent has access to appears as a separate skill. The agent's system prompt is reflected in the primary skill description, so external clients can understand what the agent is good at without any manual setup.

## Hot-reload

Like everything in Magec, A2A configuration reloads automatically. Enable or disable A2A on an agent, save, and the change is live — no restart needed.
