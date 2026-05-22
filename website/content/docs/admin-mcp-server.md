---
title: "Admin MCP Server"
---

Magec ships an embedded MCP server that exposes the entire admin API as MCP tools. Connect any MCP client (Claude Code, Cursor, mcp-cli) and you can list backends, create agents, register MCP servers, wire flows and more, all from your editor or shell.

This page covers the embedded server. For consuming external MCP servers (Home Assistant, GitHub, filesystem, etc.) see [MCP Tools](/docs/mcp/).

## What it does

The server runs alongside the user and admin HTTP servers, on its own port. It speaks Streamable HTTP, so every MCP client that follows the spec works without extra adapters.

The tool catalogue is built at startup by reading the admin OpenAPI spec embedded in the binary. Each admin endpoint becomes one MCP tool. When new admin endpoints land, the next build picks them up automatically.

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
INFO MCP server started addr=0.0.0.0:8082 url=http://0.0.0.0:8082 tools=33
```

## Authentication

The MCP server reuses `server.adminPassword`. Every request must carry the `Authorization: Bearer <adminPassword>` header. The check is constant-time and rate-limited per IP (5 failed attempts per minute), the same as the admin REST API.

If `adminPassword` is empty, the MCP server still starts but logs a warning and accepts every request. Do not expose port 8082 outside your trust boundary in that mode.

Internally the MCP layer forwards each tool call as an HTTP request back to the admin port on the loopback interface. Validation, secret redaction and conversation logging stay in the admin handlers and apply uniformly whether the caller is the admin UI, an HTTP client, or an MCP tool.

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

Tool names use the `magec_` prefix and follow the pattern `<http_method>_<resource>` derived from the admin OpenAPI operation (for example `magec_post_agents`, `magec_delete_clients_id`, `magec_get_flows`). Run `tools/list` on the server for the live list — it always matches the admin REST surface in the running binary.

## What is not exposed

Skill upload and download, plus the backup and restore endpoints, stream binary archives that do not map cleanly to MCP tool inputs and outputs. Those routes are filtered out at startup; use the admin UI or call `POST /api/v1/admin/skills/upload` and `GET /api/v1/admin/settings/backup` directly when you need them. The admin-UI helper endpoints (`/auth/check`, `/overview`) are also skipped because they are not operator actions.

## Security notes

- Secret values are never returned by the secrets endpoints. Admin REST already redacts them; the MCP server inherits that redaction because the tools are thin wrappers over the same handlers.
- The MCP server has the same authority as the admin API. Treat the bearer token with the same care.
- Running a Magec agent against the admin MCP creates a recursive surface: the agent can call destructive tools against the very instance running it. Use sparingly.
- Streamable HTTP allows server-sent events, so the MCP server's `WriteTimeout` is set to zero. The HTTP client manages cancellation through request context, which is what every spec-compliant MCP client does.

## Troubleshooting

| Symptom | Likely cause |
|--------|-------------|
| `tools=0` at startup | Swagger spec missing or empty; rerun `make swagger` and rebuild |
| 401 on every request | Wrong or missing bearer; verify `Authorization: Bearer ...` |
| 429 with `Retry-After: 60` | Hit the 5-failed-attempts-per-minute rate limit; wait one minute |
| 403 from a localhost client | DNS rebinding protection kicked in; make sure the client sends a `Host` header matching `localhost` or `127.0.0.1` |
| `admin API returned 5xx` in a tool response | The admin REST API failed; check the admin server log for the underlying error |
