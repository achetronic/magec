---
title: "Docker"
---

## Deployment options

| Deployment | Location | Description |
|---|---|---|
| Fully local | `docker/compose/fully-local/` | Ollama + Parakeet + Edge TTS. No API keys. |
| Cloud APIs | `docker/compose/remote-openai/` | Only Redis + PostgreSQL locally. LLM/STT/TTS via cloud. |

## Seed data

Set `MAGEC_SEED` to pre-populate your store on first run:

```bash
MAGEC_SEED=voice-ui docker compose up -d    # Minimal voice assistant
MAGEC_SEED=examples docker compose up -d    # Full demo with all examples
docker compose up -d                        # Empty — configure yourself
```

## Standalone container

```bash
docker run -d --name magec \
  -p 8080:8080 -p 8081:8081 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v magec_data:/app/data \
  -e MAGEC_SEED=voice-ui \
  ghcr.io/achetronic/magec:latest
```

## Building locally

```bash
docker build -f docker/build/Dockerfile -t magec .
```
