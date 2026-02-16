---
title: "Webhooks"
---

Webhook clients expose an HTTP endpoint that triggers agent invocations. External systems make an HTTP request, Magec processes it through an agent, and the response comes back in the HTTP response body. This is how you integrate Magec with CI/CD pipelines, form handlers, monitoring systems, or any tool that can make HTTP requests.

Each webhook gets a unique URL and its own authentication token. There are two modes depending on where the prompt comes from.

## Command mode

The webhook runs a preconfigured [command](/magec/docs/commands/) — a reusable prompt that you define once and trigger as many times as you want. The request body is ignored; the prompt is always the same command.

This is useful for recurring tasks that always do the same thing:

- **Daily reports** — A monitoring system hits the webhook every morning, the agent queries data and generates a summary
- **CI/CD integration** — A pipeline triggers the webhook after each deploy, the agent reviews the changelog
- **Scheduled checks** — An external scheduler calls the webhook, the agent runs a security audit

The agent, prompt, and behavior are all defined in the command. The webhook is just the trigger.

### Configuration

| Field | Description |
|-------|-------------|
| `name` | Display name |
| `passthrough` | Set to `false` for command mode |
| `commandId` | The command to run when the webhook is triggered |
| `allowedAgents` | Which agents/flows this webhook can access |

## Passthrough mode

The prompt comes from the outside. Whatever is sent in the request body gets forwarded to the agent as the user message. The webhook acts as a bridge between external systems and your agents.

This is the mode you want when the input is dynamic:

- **Form submissions** — A contact form sends the message to the agent for processing or classification
- **Alert handling** — A monitoring system sends an alert, the agent analyzes it and decides on action
- **Chat integrations** — A custom frontend or third-party service sends user messages through the webhook
- **Data processing** — External systems send data for the agent to analyze, summarize, or transform

### Configuration

| Field | Description |
|-------|-------------|
| `name` | Display name |
| `passthrough` | Set to `true` for passthrough mode |
| `allowedAgents` | Which agents/flows this webhook can access |

## Calling a webhook

Every webhook has a URL and a token. The URL format is:

```
POST http://localhost:8080/api/v1/webhooks/{clientID}
```

Authentication uses the client's token as a Bearer token:

```bash
# Command mode — body is ignored
curl -X POST http://localhost:8080/api/v1/webhooks/YOUR_WEBHOOK_ID \
  -H "Authorization: Bearer mgc_your_token"

# Passthrough mode — body contains the prompt
curl -X POST http://localhost:8080/api/v1/webhooks/YOUR_WEBHOOK_ID \
  -H "Authorization: Bearer mgc_your_token" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Analyze this error: connection timeout on service X"}'
```

The agent processes the request synchronously and the response is returned in the HTTP response body. This makes webhooks easy to integrate with any system that can make HTTP requests and read responses.

## Example integrations

### GitHub Actions

Trigger an agent after a deployment to generate release notes:

```yaml
- name: Generate release notes
  run: |
    curl -X POST ${{ secrets.MAGEC_WEBHOOK_URL }} \
      -H "Authorization: Bearer ${{ secrets.MAGEC_TOKEN }}" \
      -H "Content-Type: application/json" \
      -d '{"prompt": "Generate release notes for version ${{ github.ref_name }}"}'
```

### Monitoring alerts

Forward alerts from Prometheus, Grafana, or any monitoring tool:

```bash
curl -X POST http://magec:8080/api/v1/webhooks/alert-handler \
  -H "Authorization: Bearer mgc_..." \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Critical alert: CPU usage above 90% on prod-web-01 for 10 minutes. Analyze possible causes and suggest remediation."}'
```

### Form processing

Process contact form submissions through an agent:

```javascript
const response = await fetch('http://magec:8080/api/v1/webhooks/form-handler', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer mgc_...',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    prompt: `Classify this customer inquiry and draft a response:\n\n${formData.message}`
  })
});
```
