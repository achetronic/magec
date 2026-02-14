---
title: "Agents"
---

Each agent is an independent AI entity with its own LLM backend, system prompt, memory, voice settings, and tools. Agents are created and managed from the Admin UI.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-agents.png" alt="Admin UI — Agents list" >}}
</div>

Click any agent to open its configuration. The settings are organized in collapsible sections so you can focus on what matters.

## General

Name, description, and tags. The name is how the agent appears across the platform — in the Voice UI agent switcher, in flow steps, in logs. Tags are optional and help you organize agents when the list grows.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-agent-general.png" alt="Agent dialog — General" >}}
</div>

| Field | Description |
|-------|-------------|
| `name` | Display name shown everywhere |
| `description` | Optional note for your own reference |
| `tags` | Labels for filtering and grouping |

## System Prompt

The instructions that define who the agent is and how it behaves. This is the most important field — it shapes every response the agent produces.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-agent-prompt.png" alt="Agent dialog — System Prompt" >}}
</div>

| Field | Description |
|-------|-------------|
| `systemPrompt` | The full prompt text. Supports multi-line, markdown, examples, etc. |
| `outputKey` | Saves the agent's output under a named key. Other agents in the same flow can reference it with `{key_name}`. Useful for passing structured data between steps. |

## LLM

Which AI backend and model the agent uses for inference. The backend is selected from the ones you've configured in the [Backends](/magec/docs/backends/) section.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-agent-llm.png" alt="Agent dialog — LLM" >}}
</div>

| Field | Description |
|-------|-------------|
| `llmBackend` | The AI backend to use (OpenAI, Anthropic, Gemini, Ollama, etc.) |
| `llmModel` | Model identifier — e.g. `gpt-4.1`, `claude-sonnet-4-20250514`, `qwen3:8b` |

## Memory

Optional. Gives the agent the ability to remember across conversations.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-agent-memory.png" alt="Agent dialog — Memory" >}}
</div>

| Field | Description |
|-------|-------------|
| `memorySession` | Redis-backed session memory — stores recent conversation history. Shared by agents that use the same provider. |
| `memoryLongTerm` | PostgreSQL + pgvector long-term memory — stores semantically searchable facts across sessions. |

See [Memory](/magec/docs/memory/) for details on configuring providers.

## MCP Servers

Connect external tools to the agent via [Model Context Protocol](/magec/docs/mcp/). Each toggle enables a configured MCP server, giving the agent access to its tools.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-agent-mcp.png" alt="Agent dialog — MCP Servers" >}}
</div>

An agent with no MCP servers can still chat, but it won't be able to call external tools (file access, web search, database queries, etc.).

## Voice (STT / TTS)

Optional. Only needed if the agent will be used through the Voice UI or if you want voice responses in Telegram.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-agent-voice.png" alt="Agent dialog — Voice" >}}
</div>

**Transcription (STT):**

| Field | Description |
|-------|-------------|
| `transcriptionBackend` | Backend with Whisper-compatible STT endpoint |
| `transcriptionModel` | Model name — e.g. `whisper-1` |

**Text-to-Speech (TTS):**

| Field | Description |
|-------|-------------|
| `ttsBackend` | Backend with TTS endpoint |
| `ttsModel` | Model name — e.g. `tts-1` |
| `ttsVoice` | Voice to use — e.g. `alloy`, `nova`, `shimmer` |
| `ttsSpeed` | Playback speed multiplier — e.g. `1.0` |

{{< callout >}}
**Hot-reload:** Changes to agents take effect immediately — no server restart needed.
{{< /callout >}}
