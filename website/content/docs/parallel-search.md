---
title: "Parallel Search MCP"
---

Parallel Search is an optional external service you connect to Magec. Magec does not include or enable it by default. See [MCP Tools](/docs/mcp/) for the general setup and per-agent tool selection.

[Parallel Search MCP](https://docs.parallel.ai/integrations/mcp/search-mcp) gives an agent `web_search` for finding sources and `web_fetch` for reading specific URLs. The remote service has a free tier for exploration and light use, with no Parallel account or API key required. Free-tier rate limits apply; your agent's LLM runs on its configured backend and has its own costs.

## Configure the server

In the Admin UI, go to **MCP Servers**, click **New**, and enter these fields:

```text
Name:     Parallel Search
Type:     HTTP
Endpoint: https://search.parallel.ai/mcp
Headers:  User-Agent = Magec (https://github.com/achetronic/magec)
Prompt:   Use web_search when you need current information.
          Use web_fetch to read relevant source URLs before answering.
          Cite the sources you use.
```

In **Headers**, add `User-Agent` as the key and `Magec (https://github.com/achetronic/magec)` as the value. Leave **Skip TLS verification** off. No `Authorization` header is needed for the anonymous free tier.

## Try it

Save the server, then [enable it on the agent](/docs/mcp/#connecting-mcp-servers-to-agents) you want to use for research. Try asking: "Find the official Model Context Protocol documentation, read its introduction, and explain what MCP connects. Include the source URL."

The agent can search for sources and fetch their content through Magec's HTTP transport. If the tools are missing, check that the server is enabled on that agent and that Magec can reach `https://search.parallel.ai`. If you hit a free-tier rate limit, wait before retrying or consult the [Parallel Search MCP documentation](https://docs.parallel.ai/integrations/mcp/search-mcp) for authenticated usage and higher limits.
