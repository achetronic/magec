---
title: "MCP Tools"
---

Connect agents to external tools via the [Model Context Protocol](https://modelcontextprotocol.io) (MCP). Each MCP server provides one or more tools that agents can invoke during inference — file access, web search, database queries, smart home control, and anything else you can imagine.

## Transport types

There are two ways to connect to an MCP server, depending on how it runs:

### HTTP

For MCP servers that run as separate services (their own process, a container, a remote server). You just point Magec at the URL and it connects via HTTP/SSE.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-mcp-http.png" alt="Admin UI — New MCP Server (HTTP)" >}}
</div>

Give it a name, set the type to **HTTP**, paste the endpoint URL, and optionally add a system prompt with instructions for the LLM about when and how to use this tool.

### Stdio

For MCP servers that are CLI tools — Magec launches them as subprocesses and communicates over stdin/stdout. Perfect for tools like `uvx`, `npx`, or any local binary.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-mcp-stdio.png" alt="Admin UI — New MCP Server (Stdio)" >}}
</div>

Set the type to **Stdio**, provide the command to execute and its arguments (comma-separated). The system prompt works the same way — extra context for the LLM about what this tool does.

## System Prompt

Every MCP server has an optional system prompt field. Whatever you write here gets injected into the agent's context when this MCP is active. Use it to tell the LLM things like "use this tool to control smart home devices" or "only call this tool when the user explicitly asks for file operations".

## Compatible servers

Works with any MCP-compliant server: [Home Assistant](https://github.com/achetronic/hass-mcp), GitHub, filesystem, databases, Slack, Google Drive, and [hundreds more](https://github.com/modelcontextprotocol/servers).
