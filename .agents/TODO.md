# Magec - TODO

## High Priority

### TTS Real-Time Streaming Playback

**Problem**: Current TTS implementation waits for all audio chunks before starting playback, causing noticeable delay even with SSE streaming enabled.

**Solution**: Implement incremental playback using Web Audio API:

1. Use `AudioContext` with scheduled buffer sources
2. Decode each audio chunk as it arrives
3. Schedule playback immediately after previous chunk
4. Track `scheduledTime` to chain chunks seamlessly

**Technical Details**:
- SSE format: `data: {"type": "speech.audio.delta", "audio": "base64..."}`
- Raw audio streaming: Binary chunks via `Transfer-Encoding: chunked`
- Use `decodeAudioData()` for each chunk
- Handle codec limitations (MP3 chunks need enough data to decode)

**Files to modify**:
- `voice-ui/src/audio/OpenAITTS.js`

**Reference implementation** (not yet applied):
```javascript
async _scheduleAudioChunk(audioBytes) {
    const ctx = this._getAudioContext();
    const audioBuffer = await ctx.decodeAudioData(audioBytes.buffer.slice(0));
    
    const source = ctx.createBufferSource();
    source.buffer = audioBuffer;
    source.connect(ctx.destination);
    
    const startTime = Math.max(ctx.currentTime, this._scheduledTime);
    source.start(startTime);
    
    this._scheduledTime = startTime + audioBuffer.duration;
}
```

---

## Medium Priority

### Add Voice Activity Detection During TTS

**Problem**: On mobile, microphone can pick up speaker output and trigger wake word detection while TTS is playing.

**Possible solutions**:
- Mute microphone during TTS playback
- Implement echo cancellation
- Increase wake word threshold temporarily during playback

---

### Admin API Authentication

**Problem**: The admin API on port 8081 has no authentication. Anyone with network access can modify agents, backends, and devices.

**Possible solutions**:
- Basic auth (simple, config-based)
- API key in header
- Session-based login

---

## Low Priority

### Add More TTS Voices Configuration UI

Currently voice selection is server-side only. Could add UI for users to preview and select voices if backend supports it.

### Offline Mode

- Cache TTS responses for common phrases
- Implement service worker for full offline support
- Store transcription model locally (already partially supported via Transformers.js)

### Multi-Language Wake Words

- Support different wake word models per language
- Auto-switch based on `voice-ui/src/i18n/` selection
- Models configured in `models/wakewords.yaml`

---

## Future Ideas

### Admin UI Framework Migration (Vue/Lit/Preact)

**Problem**: The admin UI (~1000 lines of vanilla JS) is getting verbose. While schema-driven forms solved the dynamic field rendering problem without a framework, the codebase will become harder to maintain as more resource types and features are added.

**When to migrate**: When admin-ui/src/app.js exceeds ~2000 lines, or when complex UI interactions are needed (drag-and-drop, real-time sync, nested component state).

**Candidates**:
- **Vue 3** (via CDN, no build step needed) — good templating, reactivity
- **Lit** (web components, very lightweight) — no build step, native browser
- **Preact** (via CDN) — React-compatible, tiny footprint

**Key constraint**: Avoid introducing a build step (no node_modules, no bundler) if possible. CDN-first approach preferred.

### Credential Management for Connection Strings

**Problem**: Connection strings contain credentials in plain text (`redis://:password@...`, `postgres://user:pass@...`). Currently stored directly in `data/store.json` and visible in the admin UI.

**Possible approaches**:
- Environment variable expansion in connection strings (`redis://:${REDIS_PASS}@localhost:6379/0`)
- Separate secrets store (encrypted at rest)
- Reference external secret managers (Vault, K8s secrets)
- At minimum: mask passwords in API responses, only show `****` in UI

**Status**: TODO — identified during memory provider implementation.

### Speaker Identification

**Goal**: Identify who is speaking to personalize responses or restrict commands.

**Recommended solution**: [WeSpeaker](https://github.com/wenet-e2e/wespeaker)
- Active project (2024), same team as WeNet
- Docker support with Dockerfile for server/client
- ONNX export for CPU inference (~400MB image)
- Speaker embeddings for verification (1:1) or identification (1:N)

**Alternative**: SpeechBrain (heavier, ~1GB+)

### Telegram File/Artifact Support

**Goal**: Allow users to send files via Telegram for the AI to process.

**Depends on**: ADK artifacts implementation in `adk-utils-go`

**Flow**:
1. User sends file via Telegram
2. Bot downloads file
3. Converts to ADK artifact (base64 + mime type)
4. Sends to agent with message

### Database Persistence for Store

**Problem**: `data/store.json` is a single JSON file. Works fine for small setups but won't scale for multi-user or HA deployments.

**When**: When considering horizontal scaling or backup/restore requirements.

---

## Completed

- [x] OpenAI-compatible TTS backend proxy
- [x] SSE streaming support in frontend (collection mode)
- [x] Raw audio streaming support (collection mode)
- [x] Server-side TTS configuration injection
- [x] `streamFormat` config option (audio/sse)
- [x] Fix MCP Tools with Anthropic Backend
- [x] Dynamic wake word model selector from `wakewords.json`
- [x] Canvas full vertical height with capped Magec size
- [x] Footer clock replacing mode indicator
- [x] Own custom wake word models trained (OpenWakeWord)
- [x] Server-side wake word detection (moved from browser to server via WebSocket)
- [x] Server-side VAD detection (Silero VAD via ONNX)
- [x] Configurable ONNX library path (`server.onnxLibraryPath`)
- [x] Telegram voice message support (OGG→WAV conversion via ffmpeg)
- [x] Multi-agent admin API (store, CRUD, admin UI)
- [x] Device pairing authentication (voice-ui)
- [x] Memory Providers: extensible registry system (`server/memory/`)
- [x] Memory Providers: Redis (session) + Postgres (longterm) with health check
- [x] Memory Providers: universal `connectionString` for all providers
- [x] Memory Providers: admin UI with split Session/Long-Term sections
- [x] Memory Providers: dynamic form with per-type extra fields + embedding for longterm
- [x] Store-based agent creation: `agent.New()` accepts store types directly (no config dependency)
- [x] Config split: YAML for infra only (server, log, wakeWord), all resources via admin API/store
- [x] Multi-agent support (server): `agent.New()` accepts `[]AgentDefinition`, `NewMultiLoader` routes by `appName`
- [x] Multi-agent support (voice-ui): `setAgent(agentId)` on AgentClient, SessionService, OpenAITTS, RemoteTranscriber
- [x] Hot-reload agents on store changes: `OnChange()` channel + `agentRouterHandler` rebuild with 500ms debounce
- [x] Voice endpoint redesign: `/api/v1/voice/{agentId}/speech` and `/transcription` (per-agent backend resolution)
- [x] Voice proxy API key forwarding: `serveSpeechProxy` and `serveTranscriptionProxy` forward backend `apiKey`
- [x] Rename with cascade: All 6 resource types support renaming via PUT with cascading reference updates
- [x] Admin UI rename enabled: Name/ID fields editable in edit mode
- [x] Wake word model name in capabilities: `Name` field added to WebSocket capabilities message
- [x] Admin UI modal fix: `formnovalidate` on cancel/close buttons to bypass HTML5 validation
