---
title: "Store Seeds"
---

When deploying with Docker, the `MAGEC_SEED` environment variable pre-populates the store with example data on first run.

| Value | What you get |
|-------|-------------|
| *(empty)* | Empty store — configure everything from the Admin UI |
| `voice-ui` | 1 agent (Magec) + 3 backends (Ollama, Parakeet STT, Edge TTS) + Redis/Postgres memory + VoiceUI client |
| `examples` | 19 agents + 5 backends + 3 flows (Research Pipeline, Debate Arena, Software Factory) + webhook clients |

```bash
MAGEC_SEED=voice-ui docker compose up -d    # Minimal voice assistant
MAGEC_SEED=examples docker compose up -d    # Full demo
docker compose up -d                        # Empty store
```

Seeds use `${ENV_VARS}` for credentials (no secrets hardcoded). The seed is only applied on first run — once `store.json` exists, the seed is ignored.
