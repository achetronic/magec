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

### Improve Wake Word Detection Accuracy

- Tune `threshold` parameter in `voice-ui/src/config.js`
- Consider adding multiple wake word model support
- Investigate false positive/negative rates

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

### API Security (Token Rotation)

**Problem**: Voice UI is open (no auth), need to ensure API requests only come from the voice UI.

**Recommended solution**: Session token injected at runtime
1. Server generates random token on startup
2. Token injected into HTML when serving voice UI: `<script>window.MAGEC_TOKEN = "{{.SessionToken}}"</script>`
3. Voice UI sends token in header with every request
4. Server validates token, rejects requests without it
5. Token changes on every server restart

Combine with CORS strict mode and localhost binding for defense in depth.

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
