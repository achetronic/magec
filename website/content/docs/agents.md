---
title: "Agents"
---

Each agent is an independent AI entity with its own LLM backend, system prompt, memory, voice settings, and tools. Agents are created and managed from the Admin UI.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-agents.png" alt="Admin UI — Agents list" >}}
</div>

## Agent properties

| Field | Description |
|-------|-------------|
| `name` | Display name shown in UIs and logs |
| `model` | Model identifier (e.g. `gpt-4.1`, `claude-sonnet-4-20250514`, `qwen3:8b`) |
| `systemPrompt` | Instructions that define the agent's behavior and personality |
| `backendId` | Which AI backend to use for LLM inference |
| `embeddingsBackendId` | Backend for embeddings (long-term memory search) |
| `embeddingsModel` | Embedding model (e.g. `nomic-embed-text`, `text-embedding-3-small`) |
| `sttBackendId` | Backend for speech-to-text |
| `sttModel` | STT model (e.g. `whisper-1`) |
| `ttsBackendId` | Backend for text-to-speech |
| `ttsModel` | TTS model (e.g. `tts-1`) |
| `ttsVoice` | Voice name for TTS output |
| `ttsSpeed` | TTS playback speed (e.g. `1.0`) |
| `sessionMemoryId` | Redis memory provider for conversation history |
| `longTermMemoryId` | PostgreSQL/pgvector provider for semantic memory |
| `mcpIds` | List of MCP server IDs to connect as tools |

{{< callout >}}
**Hot-reload:** Changes to agents take effect immediately — no server restart needed.
{{< /callout >}}
