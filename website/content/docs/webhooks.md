---
title: "Webhooks"
---

Webhook clients expose an HTTP endpoint that triggers agent invocations. Each webhook gets a unique URL and its own authentication token. External systems hit the endpoint, the agent processes the request, and the response comes back in the HTTP response body.

There are two ways to use them depending on where the prompt comes from.

## Command mode

The webhook runs a preconfigured [command](/magec/docs/commands/) — a reusable prompt that you define once and trigger as many times as you want. The request body is ignored; the prompt is always the same.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-clients-webhook-command.png" alt="Admin UI — Webhook client (command mode)" >}}
</div>

This is useful for recurring tasks: "summarize today's news", "generate the daily report", "check for security alerts". Wire it to an external scheduler, a CI pipeline, or any system that can make an HTTP request.

## Passthrough mode

The prompt comes from the outside. Whatever is sent in the request body gets forwarded to the agent as the user message. The webhook acts as a bridge between external systems and your agents.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-clients-webhook-passthrough.png" alt="Admin UI — Webhook client (passthrough mode)" >}}
</div>

This is the mode you want when integrating with tools that need to send dynamic content — form submissions, alerts, notifications, or anything where the input changes every time.

## Calling a webhook

```bash
curl -X POST http://localhost:8080/api/v1/webhooks/WEBHOOK_ID \
  -H "Authorization: Bearer mgc_your_token" \
  -H "Content-Type: application/json" \
  -d '{"message": "Summarize today news"}'
```
