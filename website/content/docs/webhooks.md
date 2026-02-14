---
title: "Webhooks"
---

HTTP endpoints that trigger agent invocations. Each webhook gets a unique URL and authentication token.

## Modes

| Mode | Behavior |
|------|----------|
| `passthrough` | The request body is sent as the user message to the agent |
| `commandId` | A fixed command (predefined prompt) is executed regardless of request body |

## Example

```bash
curl -X POST http://localhost:8080/api/v1/webhooks/WEBHOOK_ID \
  -H "Authorization: Bearer mgc_your_token" \
  -H "Content-Type: application/json" \
  -d '{"message": "Summarize today news"}'
```
