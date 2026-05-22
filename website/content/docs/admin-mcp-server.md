---
title: "Admin MCP Server"
---

Magec ships an embedded MCP server that exposes the entire admin API as MCP tools. Connect any MCP client (Claude Code, Cursor, mcp-cli) and you can list backends, create agents, register MCP servers, wire flows and more, all from your editor or shell.

This page covers the embedded server. For consuming external MCP servers (Home Assistant, GitHub, filesystem, etc.) see [MCP Tools](/docs/mcp/).

## What it does

The server runs alongside the user and admin HTTP servers, on its own port. It speaks Streamable HTTP, so every MCP client that follows the spec works without extra adapters.

It exposes one tool per admin operation. Listing, getting, creating, updating and deleting works for every resource the admin REST API understands: backends, memory providers, MCP servers, agents, clients, commands, flows, secrets, settings, conversations, plus the type catalogues for clients, memory and voice providers.

## Enabling it

Set `server.mcp.enabled: true` in `config.yaml` and restart `magec-server`:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  adminPort: 8081
  adminPassword: "your-admin-password"
  mcp:
    enabled: true
    port: 8082
```

The startup log shows a line like:

```
INFO MCP server started addr=0.0.0.0:8082 url=http://0.0.0.0:8082 tools=61
```

## Authentication

The MCP server reuses `server.adminPassword`. Every request must carry the `Authorization: Bearer <adminPassword>` header. The check is constant-time and rate-limited per IP (5 failed attempts per minute), the same as the admin REST API.

If `adminPassword` is empty, the MCP server still starts but logs a warning and accepts every request. Do not expose port 8082 outside your trust boundary in that mode.

## Connecting from Claude Code

Add an entry to `~/.claude/mcp.json`:

```json
{
  "mcpServers": {
    "magec": {
      "type": "streamable-http",
      "url": "http://localhost:8082/",
      "headers": {
        "Authorization": "Bearer your-admin-password"
      }
    }
  }
}
```

Then `/mcp` inside Claude Code lists the `magec` tools. You can ask the agent things like *"list the Magec backends"* or *"create an OpenAI backend called bench"* and it picks the right tool.

## Connecting from mcp-cli

```bash
npx @wong2/mcp-cli streamable-http http://localhost:8082/ \
    --header "Authorization: Bearer your-admin-password" \
    tools/list
```

## Tool catalogue

Tool names use the `magec_` prefix and snake_case.

### Backends

- `magec_list_backends`
- `magec_get_backend`
- `magec_create_backend`
- `magec_update_backend`
- `magec_delete_backend`

### Memory providers

- `magec_list_memory_providers`
- `magec_get_memory_provider`
- `magec_create_memory_provider`
- `magec_update_memory_provider`
- `magec_delete_memory_provider`
- `magec_check_memory_health`
- `magec_list_memory_types`

### MCP servers (the ones agents consume)

- `magec_list_mcp_servers`
- `magec_get_mcp_server`
- `magec_create_mcp_server`
- `magec_update_mcp_server`
- `magec_delete_mcp_server`

### Agents

- `magec_list_agents`
- `magec_get_agent`
- `magec_create_agent`
- `magec_update_agent`
- `magec_delete_agent`
- `magec_list_agent_mcps`
- `magec_link_agent_mcp`
- `magec_unlink_agent_mcp`

### Clients

- `magec_list_clients`
- `magec_get_client`
- `magec_create_client`
- `magec_update_client`
- `magec_delete_client`
- `magec_regenerate_client_token`
- `magec_list_client_types`

### Commands

- `magec_list_commands`
- `magec_get_command`
- `magec_create_command`
- `magec_update_command`
- `magec_delete_command`

### Flows

- `magec_list_flows`
- `magec_get_flow`
- `magec_create_flow`
- `magec_update_flow`
- `magec_delete_flow`

### Skills

- `magec_list_skills`
- `magec_get_skill`
- `magec_delete_skill`

### Settings

- `magec_get_settings`
- `magec_update_settings`

### Secrets

- `magec_list_secrets`
- `magec_get_secret`
- `magec_create_secret`
- `magec_update_secret`
- `magec_delete_secret`

### Conversations

- `magec_list_conversations`
- `magec_get_conversation`
- `magec_delete_conversation`
- `magec_clear_conversations`
- `magec_conversation_stats`
- `magec_update_conversation_summary`
- `magec_find_conversation_pair`
- `magec_reset_conversation_session`

### Voice

- `magec_list_voice_types`

## What is not exposed

Skill upload and download (one SKILL.md plus optional `references/`, `assets/`, `scripts/`) and the backup and restore endpoints stream binary archives. They do not map cleanly to MCP tool inputs and outputs, so they stay on the admin REST API. Use the admin UI or call `POST /api/v1/admin/skills/upload` and `GET /api/v1/admin/settings/backup` directly when you need them.

## Security notes

- Secret values are never returned by `magec_get_secret` or `magec_list_secrets`. The MCP server mirrors the admin REST behaviour: GET responses contain the metadata only. Use the create/update tools to write new values.
- The MCP server has the same authority as the admin API. Treat the bearer token with the same care.
- Running a Magec agent against the admin MCP creates a recursive surface: the agent can call destructive tools (`magec_delete_*`, `magec_clear_conversations`) against the very instance running it. Use sparingly.
- Streamable HTTP allows server-sent events, so the MCP server's `WriteTimeout` is set to zero. The HTTP client manages cancellation through request context, which is what every spec-compliant MCP client does.

## Troubleshooting

| Symptom | Likely cause |
|--------|-------------|
| `tools=0` at startup | Build is stale, restart the server |
| 401 on every request | Wrong or missing bearer; verify `Authorization: Bearer ...` |
| 429 with `Retry-After: 60` | Hit the 5-failed-attempts-per-minute rate limit; wait one minute |
| 403 from a localhost client | DNS rebinding protection kicked in; make sure the client sends a `Host` header matching `localhost` or `127.0.0.1` |
