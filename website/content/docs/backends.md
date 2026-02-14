---
title: "AI Backends"
---

Backends represent connections to AI providers. Each backend can serve as LLM, embedding provider, STT (Whisper-compatible), or TTS — and you assign them per-agent.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-backends.png" alt="Admin UI — Backends" >}}
</div>

## Backend types

| Type | Provider | Required fields |
|------|----------|-----------------|
| `openai` | OpenAI, Ollama | `url` and/or `apiKey` |
| `anthropic` | Anthropic Claude | `apiKey` |
| `gemini` | Google Gemini | `apiKey` |

{{< callout type="info" >}}
**Tip:** The `openai` type works with Ollama too. Point it at `http://ollama:11434/v1` and use any local model.
{{< /callout >}}
