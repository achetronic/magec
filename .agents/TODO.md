# Magec - TODO

## ~~Client Config — DefaultAgent + ThreadHistoryLimit~~ ✅

Implemented. `defaultAgent` persisted to store on `!agent <id>` with fallback chain (in-memory → defaultAgent → allowedAgents[0]). `threadHistoryLimit` replaces hardcoded 50 in Discord/Slack `fetchThreadContext`. Shared schema helpers in `server/clients/provider.go`.

---

## ~~Large Message Handling in Telegram and Slack~~ ✅

Implemented. See `server/clients/msgutil/` package.

---

## High Priority

### Conversation Not Split After Session Reset (`!reset`)

**Problem**: When a user runs `!reset` (Discord, Slack) or `/reset` (Telegram), the ADK session is deleted, but the next message's conversation is appended to the **same conversation record** in the admin UI instead of creating a new one. The user sees a single unbroken conversation even though the agent has lost all context.

**Affects**: All chat clients — Discord, Slack, and Telegram. All three use deterministic session IDs (`{platform}_{channelID}_{agentID}`) that don't change after a reset, and the same `FindBySession` recorder logic.

**Root cause**: `FindBySession(sessionID, agentID, perspective)` in `ConversationStore` matches on the session ID, which is stable across resets. Since the session ID is recomputed identically after the reset, the recorder finds the old conversation and calls `AppendMessages` instead of creating a new `Conversation`.

**Flow**:
```
!reset → DELETE /apps/{agent}/users/default_user/sessions/{sessionID} → OK
New message → ensureSession (same ID recreated) → /run_sse → recorder calls FindBySession
  → Finds old conversation (same sessionID) → AppendMessages → new messages land in old record
```

**Possible solutions** (pick one):
1. **`closed` flag on Conversation**: `!reset` (and admin's `reset-session`) marks the conversation as `closed: true`. `FindBySession` skips closed conversations. Cleanest option — no session ID format changes, no client changes.
2. **Variable session ID component**: Add a generation counter or timestamp to the session ID (e.g. `discord_{channelID}_{agentID}_{gen}`). Changes session ID format, requires persisting the counter somewhere.
3. **Recorder-side detection**: On `ensureSession`, if the session was just created (empty history), treat it as a new conversation regardless of `FindBySession` match. Fragile — depends on being able to detect "fresh" sessions reliably.

**Files involved**:
- `server/store/conversations.go` — `FindBySession`, `Conversation` struct
- `server/clients/executor.go` — `LogExternalConversation` (create-vs-append decision)
- `server/clients/discord/bot.go` — `handleCommand` reset case
- `server/clients/slack/bot.go` — `handleCommand` reset case
- `server/clients/telegram/bot.go` — `handleResetCommand`
- `server/api/admin/conversations.go` — `reset-session` endpoint (same issue from admin side)

---

### Telegram Client: Group Channels and Thread/Topic Support

**Problem**: The Telegram client only handles DMs correctly. In groups and supergroups it has two issues:

1. **No @mention filtering in groups** — The bot processes **every** text and voice message it receives in a group (subject to Telegram's bot privacy mode), with no check for `@botname` mention. Discord and Slack both require explicit @mention in channels. Telegram should follow the same pattern for consistency and to avoid noisy behavior in active groups.

2. **Replies don't go to the originating thread/topic** — `msg.MessageThreadID` is correctly used for session scoping (`buildSessionID`) but **never set on outbound `SendMessageParams`**. In supergroups with topics enabled, the bot reads from the correct topic but replies land in the General topic instead of the thread the user wrote in. This affects all outbound calls: `handleMessage` responses, `handleVoice` responses, reactions, typing indicators, tool messages, and artifact delivery.

3. **No channel post handling** — Only `update.Message` is processed. `update.ChannelPost` (messages in Telegram channels, not groups) is never handled. If the bot is added to a Telegram channel, it silently ignores all posts.

**Current state**:
- `isAllowed(userID, chatID)` already supports group/supergroup IDs via `AllowedChats []int64` (negative numbers like `-1001234567890`)
- `buildSessionID` already scopes by `threadID` when non-zero — session isolation works
- `handleMessage` and `handleVoice` have no `msg.Chat.Type` check and no `@mention` detection
- All `SendMessage`/`SendChatAction`/`SetMessageReaction` calls use only `tu.ID(msg.Chat.ID)` without `MessageThreadID`

**Solution**:

**@mention filtering**:
- In groups/supergroups (`msg.Chat.Type` is `"group"` or `"supergroup"`), require the bot to be @mentioned in the message text
- Strip the `@botname` mention from the text before passing to the agent (same as Discord's `stripBotMention`)
- In private chats (`msg.Chat.Type == "private"`), process all messages as today
- The bot's username is available via `bot.GetMe()` at startup — cache it

**Thread/topic replies**:
- Set `MessageThreadID: msg.MessageThreadID` on all outbound `SendMessageParams` when `msg.MessageThreadID != 0`
- Set `MessageThreadID` on `SendChatActionParams` (typing indicator)
- Set `MessageThreadID` on `SetMessageReactionParams` if applicable
- For artifact delivery (`sendNewArtifacts`), pass `threadID` through so `SendDocument` also targets the correct topic

**Channel posts** (lower priority):
- Register a handler for `update.ChannelPost` events
- Same flow as `handleMessage` but using `ChannelPost` fields
- Decide if channel posts should require a specific trigger (e.g. always process, or only with a keyword/command)

**Files to modify**:
- `server/clients/telegram/bot.go` — mention filtering, thread-aware replies, optional channel post handler

---

### Secret Deletion Does Not Invalidate Agents Using It

**Problem**: When a secret is deleted via the admin API, agents that reference the secret's env var placeholder (`${SECRET_KEY}`) continue running with the old credential value. The deletion does not take effect until the server is fully restarted.

**Root cause (two bugs)**:

1. **`os.Unsetenv` is never called** — `DeleteSecret` removes the secret from the store and persists, but the env var set by `os.Setenv` on create/load is never removed from the process. The env var lingers indefinitely.

2. **In-memory expanded data is stale** — The store maintains `s.data` (env-expanded) and `s.rawData` (unexpanded). After deleting the secret and unsetting the env var, `s.data` still holds the previously expanded values. Backends, MCP servers, or client configs referencing `${SECRET_KEY}` in their URLs or tokens keep the old resolved value until `s.data` is re-expanded.

**Current state**:
- `CreateSecret` and `UpdateSecret` both call `os.Setenv(key, value)` — correct
- `DeleteSecret` does **not** call `os.Unsetenv(key)` — bug
- `UpdateSecret` does **not** unset the old key when the key name changes — secondary bug
- `persist()` calls `notifyChange()` → triggers agent rebuild, but the rebuild reads `s.Data()` which returns the stale expanded data
- `os.Unsetenv` appears **zero times** in the entire codebase

**Solution**:
1. In `DeleteSecret`: call `os.Unsetenv(existing.Key)` before removing from the slice
2. In `UpdateSecret`: if the key name changed, call `os.Unsetenv(oldKey)` before `os.Setenv(newKey, value)`
3. After unset, re-expand `s.data` from `s.rawData` via `os.ExpandEnv` so that in-memory values referencing the deleted secret resolve to empty string
4. The existing `notifyChange()` → agent rebuild pipeline then picks up the cleared values automatically

**Files to modify**:
- `server/store/store.go` — `DeleteSecret` (add `os.Unsetenv` + re-expand), `UpdateSecret` (unset old key on rename + re-expand)

---

### Multimodal File/Image Support in Clients

**Problem**: Telegram, Slack, and Discord clients only handle text and voice messages. Users sending images, documents, PDFs, or other files get silently ignored.

**Solution**: Download files from Telegram/Slack, encode as base64, and send as `inlineData` parts alongside text in the ADK `/run` request. The ADK already supports `genai.Part{InlineData: &Blob{Data, MIMEType}}` — zero backend changes needed.

**Adapter support (adk-utils-go v0.3.1)**:
- **Gemini**: passes all `InlineData` transparently to the API. Unsupported types are rejected by Google's API.
- **OpenAI**: translates images (JPEG, PNG, GIF, WebP), audio (WAV, MP3, MPEG, WebM), and files (PDF, text/*). Unsupported types return an error.
- **Anthropic**: translates images (JPEG, PNG, GIF, WebP), PDFs, and text documents (text/*). Unsupported types return an error.
- All three adapters behave the same: if a MIME type can't be translated, the request fails. No silent drops.

**File size limits**: 5MB per file, 10MB total per message, max 10 files per message. Validated client-side before download.

**Supported types (denominator común)**: JPEG, PNG, GIF, WebP. PDF and text/* work on Gemini + Anthropic. Audio works on Gemini + OpenAI.

**Telegram** (`server/clients/telegram/bot.go`):
- Current state: only `Voice` (dedicated handler) and `Text` (predicate at ~line 171 requires `Text != ""` and `Voice == nil`). Everything else is silently dropped.
- Add handler for `Document`, `Photo`, `Video`, `Audio`, `Animation`, `VideoNote`, `Sticker`. All have `FileID` → `bot.GetFile()` → download bytes.
- `Photo` is `[]PhotoSize` — use last element (highest resolution) for its `FileID`.
- `Caption` field accompanies media — include as `{"text": caption}` part alongside `{"inlineData": {...}}`.
- `callAgent()` (~line 803): change `"parts": []map[string]string{{"text": ...}}` to `[]interface{}` to support both text and inlineData parts.

**Slack** (`server/clients/slack/bot.go`):
- Current state: `handleAudioClip()` (~line 213) processes `ev.Message.Files` but only `audio/*` mimetype. Other mimetypes silently skipped.
- Add handling for `image/*`, `application/pdf`, and generic fallback for other types.
- Files are on `ev.Message.Files` (type `[]slack.File`) with `Mimetype`, `Size`, `URLPrivateDownload`/`URLPrivate`.
- Reuse existing `downloadSlackFile()` (~line 658) for all file types.
- `processMessage()` (~line 477): same parts change as Telegram.
- For `handleAppMention`: verify if files arrive via `AppMentionEvent` or need separate handling.

**ADK payload format**:
```json
{
  "parts": [
    {"text": "<!--MAGEC_META:...-->\nCaption or question about the file"},
    {"inlineData": {"mimeType": "image/png", "data": "<base64>"}}
  ]
}
```

**File size validation**: 5MB per file, 10MB total per message, max 10 files. Reject oversized files with user-friendly message.

**LLM limitations**: GPT-4o/Claude/Gemini handle images and PDFs natively. For Word/Excel/CSV, the model may not support them — the user gets a natural "I can't process this format" response from the LLM itself.

**A2A (future, non-blocking)**: `server/a2a/handler.go` declares `DefaultInputModes: []string{"text/plain"}`. When A2A file support is needed, add `"image/*"`, `"application/pdf"`, etc. The A2A handler converts `FilePart` → `genai.Part{InlineData}` before passing to ADK.

**Discord** (`server/clients/discord/bot.go`):
- Same approach as Telegram/Slack: detect non-audio attachments via `m.Attachments`, download via `att.URL`, encode base64, send as `inlineData` parts.
- Check `att.Size` < 5MB before downloading.
- `m.Content` alongside attachments becomes the caption/text part.

**Modify**: `server/clients/telegram/bot.go`, `server/clients/slack/bot.go`, `server/clients/discord/bot.go`
**No changes needed**: `server/agent/agent.go`, `server/api/user/handlers.go`, ADK library

---

### Improve Drag-and-Drop UX in Visual Flow Editor

The visual flow editor's drag-and-drop experience needs polish. Improve feedback, snapping, reordering smoothness, and overall usability when building flows visually.

**Modify**: `frontend/admin-ui/` (flow editor components)

---

### Line Breaks in Voice UI Text Chat

**Problem**: The text input in the Voice UI doesn't support multi-line messages. Pressing Enter sends the message immediately with no way to insert a line break.

**Solution**: Support Shift+Enter (or similar) for inserting line breaks. Switch input from `<input>` to `<textarea>` (or equivalent) with auto-resize behavior. Enter sends, Shift+Enter adds newline.

**Modify**: `frontend/voice-ui/`

---

### Tool Execution Visibility in Clients

**Problem**: When the agent executes tools during a conversation, the user has no visibility into what's happening behind the scenes. This creates a black-box experience that erodes trust.

**Solution**: Show tool calls and their results as collapsible/summarized blocks so users can see what the agent did without cluttering the main conversation flow. Each client adapts to its platform's formatting capabilities.

**Platform collapsible support**:

| Client | Collapsible nativo | Mecanismo |
|--------|-------------------|-----------|
| **Telegram** | **Yes** | `<blockquote expandable>...</blockquote>` (HTML parse mode) — collapsed by default, user taps to expand |
| **Slack** | **No** | No collapsible blocks in mrkdwn or Block Kit. Show a short summary line like `🔧 Tool: search_memory (completed)` without full details |
| **Voice UI** | **Yes** | Custom Vue component — `<details>/<summary>` or click/tap collapsible block |

**Implementation per client**:

**Telegram** (`server/clients/telegram/bot.go`):
- Switch from `Markdown` parse mode to `HTML` parse mode in `sendResponse()`
- Extract tool call events from ADK response (already available as `functionCall`/`functionResponse` parts in the events array)
- Before the main text response, send tool execution info wrapped in `<blockquote expandable>🔧 tool_name\n\nInput: ...\nOutput: ...</blockquote>`
- Collapsed by default — user taps to see full tool details
- If multiple tools were called, group them in a single expandable blockquote or send one per tool

**Slack** (`server/clients/slack/bot.go`):
- No native collapsible support — use a compact summary format
- Before or above the main response text, add a line per tool: `🔧 *tool_name* — completed` (mrkdwn bold)
- Optionally use a Slack `context` block (smaller, muted text) for tool summaries if switching to Block Kit messaging
- Full tool input/output not shown (Slack has no way to hide it behind a toggle)

**Voice UI** (`frontend/voice-ui/src/components/ChatMessage.vue`):
- Add a new message type or section for tool calls in the chat timeline
- Render as a collapsible block: header shows `🔧 tool_name`, body (hidden by default) shows input args and output
- Style: muted colors (`bg-piedra-800`, `text-arena-500`), click to expand/collapse
- Tool events are already present in the ADK `/run` response as `functionCall` and `functionResponse` parts — extract them in `AgentClient.js` `_extractResponses()`

**ADK response events structure** (tool calls are already in the response):
```json
[
  {
    "author": "agent_name",
    "content": {
      "parts": [
        {"functionCall": {"name": "search_memory", "args": {"query": "..."}}}
      ]
    }
  },
  {
    "author": "agent_name",
    "content": {
      "parts": [
        {"functionResponse": {"name": "search_memory", "response": {"result": "..."}}}
      ]
    }
  },
  {
    "author": "agent_name",
    "content": {
      "parts": [
        {"text": "Here is the final answer..."}
      ]
    }
  }
]
```

**Key decisions**:
- Tool visibility is **per-client** — each client renders what its platform allows
- Telegram and Voice UI get full collapsible details; Slack gets a compact summary
- Tool info is sent **alongside** the response, not as a separate message (except Telegram where it may be a preceding message with expandable blockquote)
- No server changes needed — tool events are already in the ADK `/run` response; clients just need to extract and render them

**Discord** (`server/clients/discord/bot.go`):
- Same as Telegram: use `<blockquote expandable>` equivalent if Discord supports it, otherwise compact summary like Slack.
- Discord supports markdown but no native collapsible blocks — use a compact summary line per tool: `🔧 **tool_name** — completed`.

**Modify**: `server/clients/telegram/bot.go`, `server/clients/slack/bot.go`, `server/clients/discord/bot.go`, `frontend/voice-ui/src/components/ChatMessage.vue`, `frontend/voice-ui/src/lib/api/AgentClient.js`

---

### File Upload Support in Voice UI Text Chat

**Problem**: Users can only send text and voice from the Voice UI. There's no way to attach files (images, PDFs, etc.) to a message from the web chat.

**Solution**: Add a file attachment button to the text input area. Upload files, encode as base64, and send as `inlineData` parts alongside text in the `/run` request (same format as the Telegram/Slack multimodal support). Show file previews/thumbnails in the chat.

**Modify**: `frontend/voice-ui/`, `server/api/user/handlers.go` (if multipart upload needed)

---

### Human-in-the-Loop Tool Confirmation

**Problem**: MCP tools can perform sensitive actions (delete data, send emails, execute code). There's no way to require human approval before execution.

**Solution**: Use ADK v0.5.0's native `RequireConfirmationProvider` on `MCPToolset.Config`. This is a dynamic per-tool callback — no need to wrap tools manually or build a custom confirmation layer.

**Design decisions**:
- **Confirmation list lives on the agent, not on the MCP server**. A tool may be dangerous for a public-facing agent but fine for an internal one. The MCP is a shared resource — marking tools there would force the same policy on all agents.
- **Agent config**: new field `toolConfirmation: ["delete_record", "send_email", "execute_*"]` — list of tool names/globs that require confirmation for this agent.
- **Provider in `buildToolsets()`**: when creating each `MCPToolset`, pass a `RequireConfirmationProvider` that checks the agent's `toolConfirmation` list. Signature: `func(toolName string, args any) bool`.
- **Admin UI — per-MCP tool selection**: the agent form shows tools from each connected MCP server (fetched via `client.ListTools()`). The user toggles which ones require confirmation. Stored as a list of tool names/globs per agent.

**"Always Allow" (client-side, not in ADK)**:
- ADK has no built-in "confirm forever" — each invocation is independent.
- Implement a shared `alwaysAllow map[string]bool` behind the `RequireConfirmationProvider`. When a user clicks "Always Allow", the client sends `confirmed: true` AND updates the map — the provider returns `false` for that tool from then on.
- Scope: session-scoped by default (resets on reconnect). Consider persisting per user/workspace later.

**Chat UI confirmation dialog** (all clients):
- Render the `adk_request_confirmation` event as a dialog with three buttons:
  - **Approve** — confirm this invocation only
  - **Reject** — deny this invocation
  - **Always Allow** — confirm + add to `alwaysAllow` map
- Show tool name, hint text, and input args so the user knows what they're approving.

**Client changes (all already use `/run_sse`)**:
- All clients (Telegram, Slack, Discord, executor) call `/run_sse` via `callAgentSSE()` and parse events with `msgutil.ParseSSEStream()`.
- **Telegram**: listen for `adk_request_confirmation` SSE events, show inline keyboard (Approve/Reject/Always Allow), send `FunctionResponse` back.
- **Slack**: show interactive block with buttons, handle callback.
- **Voice UI**: show collapsible confirmation card in chat timeline with Approve/Reject/Always Allow buttons.
- **Executor** (`server/clients/executor.go`): auto-approve or skip (cron/webhook triggers can't wait for a human).

**Protocol flow**:
1. `RequireConfirmationProvider` returns `true` → ADK's `mcpTool.Run()` calls `ctx.RequestConfirmation(hint, payload)` automatically
2. ADK emits `FunctionCall` event with name `adk_request_confirmation` via SSE
3. Client shows confirmation prompt to user (tool name, hint, args)
4. User approves/rejects → client sends `FunctionResponse` with `{confirmed: true/false}`
5. Tool is re-invoked, reads `ctx.ToolConfirmation().Confirmed` and proceeds or aborts

See `.agents/ADK_TOOLS.md` for protocol details.

**Modify**: `server/agent/agent.go` (`RequireConfirmationProvider` in `MCPToolset.Config`), `server/store/types.go` (agent config field), `server/clients/telegram/bot.go`, `server/clients/slack/bot.go`, `server/clients/executor.go`, `frontend/voice-ui/`, `frontend/admin-ui/`

---

### ~~Artifact Management Toolset~~ ✅

Implemented. See `server/agent/tools/artifacts/toolset.go` — provides `save_artifact`, `load_artifact`, and `list_artifacts` tools via `functiontool.New`. Supports text and base64 binary content. Wired into `base_toolset.go` so all agents get it. Filesystem-backed via `adk-utils-go/artifact/filesystem` (persists across restarts). Clients (Telegram, Slack, and Discord) auto-deliver new artifacts as file attachments after each `/run` response using before/after diff of the artifact list REST endpoint.

---

### Voice Provider Registry — Multi-Backend TTS/STT Support

**Problem**: The TTS and STT proxies (`serveSpeechProxy`, `serveTranscriptionProxy` in `main.go`) are hardcoded to OpenAI-compatible endpoints (`/v1/audio/speech`, `/v1/audio/transcriptions`). `BackendDefinition.Type` (`openai`, `anthropic`, `gemini`) is completely ignored — a Gemini backend assigned to TTS gets requests sent to `/v1/audio/speech`, which Gemini doesn't serve. Users who want to use Gemini's native TTS/STT have no path today.

**Provider landscape**:

| Provider | TTS | STT | API shape |
|----------|-----|-----|-----------|
| OpenAI (+ compatible: edge-tts, parakeet, etc.) | `/v1/audio/speech` | `/v1/audio/transcriptions` | Already supported |
| Gemini | `generateContent` with `speechConfig` + `responseModalities: ["AUDIO"]` | `generateContent` with audio `inlineData` | Needs translation |
| Anthropic | No API | No API | N/A |
| xAI | `/v1/tts` (own format) | No standalone endpoint | Future candidate |

**Solution**: Extract TTS/STT into a provider interface with per-backend-type implementations. Same pattern as the client and memory provider registries — `init()` + blank imports.

**Interface design**:

```go
// server/voice/provider.go

type TTSRequest struct {
    Input          string
    ResponseFormat string // "opus", "mp3", "wav", etc.
    Model          string
    Voice          string
    Speed          float64
}

type TTSProvider interface {
    SynthesizeSpeech(ctx context.Context, req TTSRequest, backend store.BackendDefinition) (io.ReadCloser, string, error)
    // Returns: audio stream, content-type, error
}

type STTRequest struct {
    Audio       io.Reader
    ContentType string // original multipart content-type
    Model       string
}

type STTProvider interface {
    TranscribeAudio(ctx context.Context, req STTRequest, backend store.BackendDefinition) (string, error)
    // Returns: transcribed text, error
}
```

**Implementations**:

```
server/voice/
├── provider.go         — TTSProvider + STTProvider interfaces, registry
├── registry.go         — Register(), GetTTS(), GetSTT() by backend type
├── openai/
│   └── openai.go       — OpenAI-compatible TTS + STT (extract from current main.go)
└── gemini/
    └── gemini.go       — Gemini TTS + STT via generateContent
```

**OpenAI provider** (`server/voice/openai/openai.go`):
- TTS: POST `{url}/v1/audio/speech` with `{input, model, voice, speed, response_format}` → stream raw audio back
- STT: POST `{url}/v1/audio/transcriptions` with multipart form + model → parse `{"text": "..."}` response
- Direct extraction of current `serveSpeechProxy`/`serveTranscriptionProxy` logic

**Gemini provider** (`server/voice/gemini/gemini.go`):

TTS:
- POST `{url}/v1beta/models/{model}:generateContent` with API key via `x-goog-api-key` header (or Bearer if using Vertex AI)
- Request body:
  ```json
  {
    "contents": [{"role": "user", "parts": [{"text": "..."}]}],
    "generationConfig": {
      "responseModalities": ["AUDIO"],
      "speechConfig": {
        "voiceConfig": {
          "prebuiltVoiceConfig": {"voiceName": "Kore"}
        }
      }
    }
  }
  ```
- Response: `candidates[0].content.parts[0].inlineData` with base64 PCM audio (24kHz 16-bit)
- Provider decodes base64, wraps in WAV header (or converts to requested format via ffmpeg), returns stream
- Voice mapping: Gemini has 30 voices (Kore, Charon, Leda, Puck, etc.) — user sets `voice` in TTSRef as usual

STT:
- POST same `generateContent` endpoint with the audio as `inlineData` base64
- Request body:
  ```json
  {
    "contents": [{"role": "user", "parts": [
      {"text": "Transcribe this audio."},
      {"inlineData": {"mimeType": "audio/wav", "data": "base64..."}}
    ]}]
  }
  ```
- Response: extract text from `candidates[0].content.parts[0].text`
- Provider reads multipart audio from client, base64-encodes it, builds generateContent request, parses text response

**Proxy refactor** (`main.go`):

```go
func serveSpeechProxy(w http.ResponseWriter, r *http.Request, agentDef store.AgentDefinition, dataStore *store.Store) {
    backend, ok := dataStore.GetBackend(agentDef.TTS.Backend)
    // ...
    provider := voiceregistry.GetTTS(backend.Type) // falls back to openai
    if provider == nil {
        http.Error(w, `{"error":"TTS not supported for this backend type"}`, http.StatusBadRequest)
        return
    }
    // parse client body → TTSRequest
    stream, contentType, err := provider.SynthesizeSpeech(r.Context(), req, backend)
    // write stream to response
}
```

**Auth handling**: Gemini uses `x-goog-api-key: {apiKey}` instead of `Authorization: Bearer`. Each provider handles its own auth from `backend.APIKey`.

**Gemini URL**: The `BackendDefinition.URL` for Gemini backends is already `https://generativelanguage.googleapis.com` (or a custom Vertex AI URL). The provider appends `/v1beta/models/{model}:generateContent`.

**No store/type changes needed**: `BackendDefinition.Type` already exists with `gemini` as a value. `TTSRef` and `BackendRef` already have `backend`, `model`, `voice`, `speed` — all applicable to Gemini. The provider just reads them differently.

**Admin UI**: No changes. User creates a Gemini backend (type `gemini`, URL + API key), assigns it as the agent's TTS/STT backend with a Gemini TTS model (`gemini-2.5-flash-preview-tts`) and a voice name (`Kore`). The proxy dispatches to the right provider automatically.

**Future extensibility**: Adding xAI would be `server/voice/xai/xai.go` implementing the same interfaces. Register in `init()`, done.

**Files to modify**:
- `server/voice/provider.go` (new — interfaces)
- `server/voice/registry.go` (new — registry)
- `server/voice/openai/openai.go` (new — extracted from main.go)
- `server/voice/gemini/gemini.go` (new — Gemini implementation)
- `server/main.go` — refactor `serveSpeechProxy`/`serveTranscriptionProxy` to use registry
- `server/main.go` — blank imports for `voice/openai` and `voice/gemini`

**Files NOT modified**: `store/types.go`, `config/config.go`, `api/admin/`, `frontend/` — zero data model or UI changes.

---

### TTS Real-Time Streaming Playback

**Problem**: Current TTS waits for all audio chunks before playback. Noticeable delay.

**Solution**: Incremental playback using Web Audio API — decode and schedule each chunk as it arrives.

**Modify**: `frontend/voice-ui/src/lib/audio/OpenAITTS.js`

---

## Medium Priority

### Filter Tool Messages from `fetchThreadContext`

**Background**: `fetchThreadContext` is confirmed necessary. ADK only accumulates messages that go through the bot — it has no visibility into messages other users post in the channel/thread. Without `fetchThreadContext`, the agent answers as if it has no prior context from the conversation.

**Problem**: When `!showtools` is active, tool call/result messages posted by the bot into the thread are picked up by `fetchThreadContext` and injected into the next turn's `THREAD_HISTORY` as noise. The LLM already has that tool activity in its ADK session.

**Solution**: In both `fetchThreadContext` implementations, skip messages from the bot itself that look like tool output. These are identifiable because they are sent by the bot (`msg.Author.ID == botID` in Discord, `msg.BotID != ""` in Slack) and their text starts with the tool formatting prefix (`⚡` for tool calls, `✅` for tool results, `⚙️` for the counter).

**Files**: `server/clients/slack/bot.go`, `server/clients/discord/bot.go`

---

### Admin UI: Strip Metadata from Messages in ConversationDetail

**Problem**: `msg.content` in the conversation detail view is rendered raw via `renderMarkdown()`. If a message contains `<!--MAGEC_META:...:MAGEC_META-->` or `<!--MAGEC_THREAD_HISTORY:...:MAGEC_THREAD_HISTORY-->`, those tags and their contents are visible to the admin — noise that adds nothing for a human reader.

**Solution**: Import `stripMetadata` from `src/lib/metadata.js` (already exists and is already used in `ConversationsList.vue`) and apply it in `renderMarkdown()` or directly on the `v-html` binding in `ConversationDetail.vue` (line 210).

Also apply strip in the PDF export (`handleExportPDF`, line 551 — `marked.parse(m.content)` also renders content without stripping).

**Files**: `frontend/admin-ui/src/views/conversations/ConversationDetail.vue`

---

### Composable Flows (flow-as-step)

**Problem**: Flows can only reference agents in their steps. To build complex pipelines (e.g. a content pipeline that includes a review sub-pipeline), users have to flatten everything into a single flow, which becomes unwieldy.

**Solution**: Allow a flow step to reference another flow ID, not just an agent ID. Since flows already compile to ADK agents (`SequentialAgent`, `ParallelAgent`, `LoopAgent`) and register in `adkAgentMap`, a step pointing to a flow ID resolves to the sub-flow's compiled agent.

**Key design decisions**:

- **Compilation order**: Build flows in topological order (leaf flows first). Flows with no flow-dependencies compile first, then flows that reference them.
- **Cycle detection**: Reject flow A → flow B → flow A at save time (admin API validation) and at compile time (`BuildFlowAgent`).
- **responseAgent inheritance**: A step pointing to a sub-flow has a toggle: "inherit responseAgents" (default) or "silence" (step produces no public output). No partial override — inherit all or none. To change which agents are responseAgents, edit the sub-flow directly.
- **Output key**: Sub-flow's output key passes through as the step's output, same as a regular agent step.
- **UI**: Flow step agent selector shows both agents and flows (distinguished by `type` field). Cycle detection prevents selecting a flow that would create a circular dependency.

**Implementation**:

1. `BuildFlowAgent` (`server/agent/flow.go`): when resolving a step's agent ID, look in `adkAgentMap` which already contains both agents and compiled flows
2. Add topological sort in `agent.go` `New()` before the flow compilation loop (~line 181)
3. Add cycle detection: build dependency graph from flow steps, reject if cycle found
4. `FlowStep` gains `InheritResponseAgents *bool` (default true). When false, the step's sub-flow responseAgents are excluded from the parent flow's `ResponseAgentIDs()`
5. Admin API: validate no cycles on flow create/update
6. Admin UI: flow step selector includes flows, visual indicator for flow vs agent

**Modify**: `server/agent/flow.go`, `server/agent/agent.go`, `server/store/types.go`, `server/api/admin/flows.go`, `frontend/admin-ui/`

---

### Refactor MemoryCard to use Card component

`MemoryCard.vue` duplicates hover styles from `Card.vue`. Should wrap `<Card color="green">` instead.

**Modify**: `frontend/admin-ui/src/views/memory/MemoryCard.vue`

---

### Voice Activity Detection During TTS

On mobile, microphone picks up speaker output and triggers wake word during TTS playback. Options: mute mic during TTS, echo cancellation, or increase threshold temporarily.

---

### Move `response_format` Out of Clients

TTS `response_format` (opus, mp3, wav) is hardcoded per client. Could be per-agent in `TTSRef`, per-client in config, or documented as client contract. **Decision**: TBD. Related to the voice provider registry — Gemini returns PCM and the provider handles format conversion, so `response_format` may need to become a provider-level concern rather than a client one.

---

### Remote A2A Agents as Tools (orchestration mode)

**Problem**: A user may have multiple A2A agents deployed across their network (e.g. researcher, architect, code reviewer). They want a local "header" agent (MetaMagecAgent) that can call those remote agents when it decides, consolidate their responses, and deliver a unified answer to the user.

**Solution**: Use ADK's `agent/remoteagent` + `tool/agenttool` to wrap each remote A2A agent as a tool callable by the orchestrator's LLM. The orchestrator maintains full control — it decides which remotes to call, can call multiple, and consolidates before responding.

**How it works**:
```
User → MetaMagecAgent (LLM + remote agent tools)
           ├── ask_architect("design this system") → A2A call → response
           ├── ask_researcher("find prior art")    → A2A call → response
           └── LLM consolidates both → responds to user
```

**ADK native support** (already available in v0.4.0):
```go
import (
    "google.golang.org/adk/agent/remoteagent"
    "google.golang.org/adk/tool/agenttool"
)

remote, _ := remoteagent.NewA2A(remoteagent.A2AConfig{
    Name:            "architect",
    AgentCardSource: "http://architect-agent:8080",
})
architectTool := agenttool.New(remote, nil)
```

**What to implement in magec**:
1. New entity `RemoteAgent` in the store: `{id, name, agentCardURL, credentials}`
2. In `buildToolsets()` (`server/agent/agent.go`): for each remote agent configured on the agent, create `remoteagent.NewA2A()` + `agenttool.New()` and add to toolsets
3. Agent config: new field `remoteAgents []string` (list of RemoteAgent IDs), similar to how `mcpServers` works
4. Admin UI: section for managing remote agents (CRUD), agent form gains a "Remote Agents" multi-select like MCPs
5. System prompt guidance: the orchestrator agent's prompt should describe what each remote agent does so the LLM knows when to use them

**Characteristics**:
- Orchestrator always keeps control
- Can call multiple remotes per turn
- Can compare, filter, reformulate remote responses
- User always talks to the orchestrator, never directly to remotes
- Works as a flow step — the flow doesn't know or care that it uses remote agents internally

**Modify**: `server/agent/agent.go`, `server/store/types.go`, `server/api/admin/`, `frontend/admin-ui/`

---

### Remote A2A Agents as Sub-agents (transfer mode)

**Problem**: In some cases, a remote A2A agent needs to interact directly with the user — ask clarifying questions, iterate on a solution, have a multi-turn conversation — without the orchestrator in the middle adding latency and losing context.

**Solution**: Use ADK's `agent/remoteagent` to create the remote as a proper sub-agent. The orchestrator's LLM can "transfer" the conversation to the remote agent. The remote then talks directly with the user until it's done, then control returns to the orchestrator.

**How it works**:
```
User → MetaMagecAgent: "I need a system architecture"
  MetaMagecAgent → transfer to architect
    User ↔ Architect (direct multi-turn conversation)
    Architect: "done, here's the design"
  ← control returns to MetaMagecAgent
MetaMagecAgent → continues with next step
```

**ADK native support**:
```go
remote, _ := remoteagent.NewA2A(remoteagent.A2AConfig{
    Name:            "architect",
    AgentCardSource: "http://architect-agent:8080",
})
// Pass as sub-agent directly, no agenttool wrapper
orchestrator, _ := llmagent.New(llmagent.Config{
    SubAgents: []agent.Agent{remote},
    // ...
})
```

**Characteristics**:
- Remote gets full conversation context and direct user interaction
- No orchestrator latency/tokens in the middle during the transfer
- Remote can use all its own tools and personality
- Orchestrator loses visibility during the transfer
- One transfer at a time (can't talk to two remotes simultaneously)
- Better for deep specialization tasks that need multi-turn interaction

**When to use which**:

| Scenario | Use |
|---|---|
| "Ask the researcher for X and the architect for Y, then combine" | Tool mode |
| "Hand this off to the architect, let them work it out with the user" | Transfer mode |
| Quick factual queries to remotes | Tool mode |
| Complex tasks needing clarification/iteration | Transfer mode |

**Implementation**: Same `RemoteAgent` entity as tool mode. The agent config would specify per-remote whether it's a tool or sub-agent (or both — ADK allows it). Can be implemented after tool mode as an incremental addition.

**Modify**: Same files as tool mode, plus sub-agent wiring in `agent.go`

---

### Evaluate Flow Subagent Invocation Model

Should clients target sub-agents within flows? Should flows support conditional routing? Should execution include per-step metadata? Should flows be composable (reference other flows)?

Design evaluation for when more complex workflows are needed.

---

### Evaluate Subagent-as-Tool Pattern

ADK supports agents as tools — orchestrator decides at runtime which specialists to call. More flexible than static flows but harder to represent in the UI. Design evaluation for when sequential/parallel model feels too rigid.

---

### ContextGuard Summary Tier Migration (app/user scope)

**Problem**: ContextGuard summaries are currently session-scoped. When a user switches client (Discord → Telegram) or starts a new thread, the agent loses all conversation context from previous sessions. The summary dies with the session.

**Blocked by**: All users are currently `default_user`. Moving summaries to `app:` tier would share them across **all** clients/channels for that agent — if a user asks about Kubernetes deployments on Discord and networking on Telegram, both contexts contaminate each other's summary.

**Solution (requires real user identity)**:
1. Implement per-client user identity: each client generates a meaningful `userID` (e.g. `discord_123456`, `slack_U0ABC`, `telegram_98765`) instead of `default_user`
2. Move ContextGuard state keys to `user:` tier (`session.KeyPrefixUser` prefix) so summaries are scoped per-user across all that user's sessions with a given agent
3. The `user:` tier in `adk-utils-go` v0.7.0 already supports differentiated TTL (defaults to no expiration), so summaries survive indefinitely

**What's already in place**:
- `adk-utils-go` v0.7.0 has full tier support (`app:`, `user:`, `temp:`) with independent TTLs for app/user state (default: no expiration, matching canonical ADK DatabaseService behaviour)
- ContextGuard state keys are simple string constants in `adk-utils-go/plugin/contextguard/contextguard.go` — adding the prefix is a one-line change per key
- The Redis session service stores `user:` state in a dedicated HASH (`userstate:{appName}:{userID}`) separate from session data

**Cross-client identity (future)**:
If a single person uses Discord AND Telegram, they'd have two `userID`s and two separate summaries — which is actually correct (different conversational contexts). True cross-client identity (linking `discord_123` and `telegram_456` as the same person) is a separate, larger problem.

**Modify**: `adk-utils-go/plugin/contextguard/contextguard.go` (state key prefixes), `server/clients/telegram/bot.go`, `server/clients/slack/bot.go`, `server/clients/discord/bot.go`, `server/clients/executor.go`

---

## Low Priority

### ~~Skill Card View Formatter~~ ✅

Implemented. Frontend parses YAML frontmatter from `instructions` field via `lib/frontmatter.js` (uses `js-yaml`). Canonical skills (valid frontmatter with `name`) render structured cards with description, license/compatibility badges, and file count. Non-canonical skills fall back to store name/description. Store-level name/description always takes priority over frontmatter — frontmatter values are fallback only.

---

### ~~Skill Package Upload (ZIP/tar.gz)~~ ✅

Implemented. `POST /skills/{id}/package` extracts ZIP or tar.gz, requires `SKILL.md` at root (or one level deep — auto-stripped). Preserves directory structure in `data/skills/{id}/`. If SKILL.md has valid frontmatter, `name` and `description` are extracted for the store; otherwise name defaults to archive filename. `SkillDialog.vue` uses `SegmentedControl` for Manual | Package toggle — Package mode shows a drop zone for compressed files.

---

### More TTS Voices Configuration UI

Voice selection is server-side only. Could add UI for preview and selection.

### Offline Mode

Cache TTS, service worker, local transcription model.

### Multi-Language Wake Words

Different models per language, auto-switch based on i18n selection.
