# Magec - TODO

## High Priority

### Device Pairing Authentication

**Problem**: Voice UI and API have no authentication. Need simple auth without OIDC complexity.

**Solution**: Device pairing with one-time code
1. First time user opens voice-ui, shows "Enter pairing code" screen
2. Server generates 6-digit code on startup, logs it to console
3. User enters code in voice-ui
4. Server validates and returns a device token (stored in localStorage)
5. All subsequent requests include device token in header
6. Server validates token on each request
7. Admin can revoke tokens or regenerate pairing code via config/API

**Files to modify**:
- `server/main.go` - Token generation, validation middleware
- `server/config/config.go` - Store active device tokens
- `voice-ui/src/app.js` - Pairing flow UI
- `voice-ui/src/api/AgentClient.js` - Include token in requests

---

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
