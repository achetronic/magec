---
title: "Memory"
---

Agents can use two types of memory, configured independently per agent.

## Session memory (Redis)

Conversation history stored in Redis with a configurable TTL. Survives server restarts. Each user+session pair gets its own history.

| Field | Description |
|-------|-------------|
| `type` | `redis` |
| `url` | Redis connection URL (e.g. `redis://redis:6379`) |
| `ttl` | Time-to-live for session entries (e.g. `24h`) |

## Long-term memory (PostgreSQL + pgvector)

Semantic memory using vector embeddings. Agents automatically search for relevant past interactions and save important user preferences. Requires a backend configured for embeddings.

| Field | Description |
|-------|-------------|
| `type` | `postgres` |
| `url` | PostgreSQL connection URL |

The Admin UI includes a health check feature to test memory provider connectivity.
