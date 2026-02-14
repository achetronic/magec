---
title: "API Reference"
---

Magec exposes two HTTP servers with full Swagger documentation.

## User API (`:8080`)

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/agent/run` | POST | Run agent (blocking response) |
| `/api/v1/agent/run_sse` | POST | Run agent (SSE streaming) |
| `/api/v1/voice/{agentId}/transcription` | POST | Speech-to-text proxy |
| `/api/v1/voice/{agentId}/speech` | POST | Text-to-speech proxy |
| `/api/v1/voice/events` | WebSocket | Real-time wake word + VAD events |
| `/api/v1/webhooks/{id}` | POST | Webhook trigger |
| `/api/v1/client/info` | GET | Client pairing info |
| `/api/v1/health` | GET | Health check |
| `/swagger/` | GET | Swagger UI |

## Admin API (`:8081`)

Full CRUD for 7 resource types (35+ endpoints):

| Resource | Endpoints | Notes |
|----------|-----------|-------|
| Backends | CRUD | OpenAI, Anthropic, Gemini |
| Memory | CRUD + health check | Redis (session), PostgreSQL (long-term) |
| MCP Servers | CRUD | HTTP and Stdio transport |
| Agents | CRUD + MCP linking | Per-agent LLM, STT, TTS, memory, tools |
| Flows | CRUD | Multi-agent workflow definitions |
| Commands | CRUD | Reusable prompts for cron/webhooks |
| Clients | CRUD + token regen | Direct, Telegram, Webhook, Cron |

Swagger UI available at `http://localhost:8081/swagger/`.

## Example: Talk to an agent

```bash
curl -X POST http://localhost:8080/api/v1/agent/run \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer mgc_your_client_token" \
  -d '{
    "app_name": "AGENT_ID",
    "user_id": "user1",
    "session_id": "session1",
    "new_message": {
      "role": "user",
      "parts": [{ "text": "Hello!" }]
    }
  }'
```
