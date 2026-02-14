---
title: "Commands"
---

Commands are reusable prompts that can be referenced by cron jobs and webhooks. Instead of duplicating the same prompt across multiple triggers, define it once as a command and reference it by ID.

| Field | Description |
|-------|-------------|
| `name` | Display name |
| `prompt` | The text to send to the agent |
| `agentId` | Which agent (or flow) to invoke |
