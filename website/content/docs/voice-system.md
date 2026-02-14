---
title: "Voice System"
---

Server-side voice processing powered by ONNX Runtime. All audio processing happens on the server, not in the browser.

## Components

| Component | Model | Description |
|-----------|-------|-------------|
| Wake Word | OpenWakeWord | Custom "Oye Magec" and "Magec" models |
| VAD | Silero VAD | Detects when you start/stop speaking |
| STT | Whisper-compatible | Parakeet, OpenAI, or any compatible API |
| TTS | OpenAI-compatible | openai-edge-tts, OpenAI, ElevenLabs, etc. |

## WebSocket protocol

The `/api/v1/voice/events` WebSocket endpoint handles real-time audio streaming. The client sends raw audio frames, and the server responds with wake word detection and VAD events. This enables low-latency voice interaction without polling.

## Disabling voice

Set `voice.ui.enabled: false` in your config to disable the Voice UI and all voice-related routes. The Admin UI and API remain fully functional.
