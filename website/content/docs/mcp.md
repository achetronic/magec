---
title: "MCP Tools"
---

Connect agents to external tools via the [Model Context Protocol](https://modelcontextprotocol.io) (MCP). Each MCP server provides one or more tools that agents can invoke during inference.

## Transport types

| Type | Description | Use when |
|------|-------------|----------|
| `http` | Connect to a remote MCP server via HTTP/SSE | MCP server runs as a separate service |
| `stdio` | Launch a local MCP server as a subprocess | MCP server is a CLI tool |

## Configuration fields

| Field | Description |
|-------|-------------|
| `name` | Display name |
| `transport` | `http` or `stdio` |
| `url` | Server URL (HTTP transport) |
| `command` | Executable path (Stdio transport) |
| `args` | Command arguments (Stdio transport) |
| `headers` | Custom HTTP headers (auth tokens, etc.) |
| `systemPrompt` | Additional context injected into the agent's system prompt when this MCP is active |

Works with any MCP server: [Home Assistant](https://github.com/achetronic/hass-mcp), GitHub, filesystem, databases, Slack, Google Drive, and [hundreds more](https://github.com/modelcontextprotocol/servers).
