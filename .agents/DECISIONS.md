# Technical Decisions

Technical decisions made by the project owner. Any AI tool working on this repository
**must respect these decisions** and not revert them without explicit approval.

---

## Middlewares in their own package with httpsnoop

**Date**: 2026-02-13
**Status**: Implemented

HTTP middlewares (`AccessLog`, `CORS`, `ClientAuth`) live in `server/middleware/`,
**not** in `main.go`.

To capture status code and bytes written in the access log, we use
[`httpsnoop.CaptureMetrics`](https://github.com/felixge/httpsnoop) instead of a custom
wrapper over `http.ResponseWriter`. httpsnoop correctly handles `Hijacker`, `Flusher`,
`CloseNotifier`, `Pusher` and any other optional interface without additional code.

**Do not use**: custom `responseRecorder` or manual wrappers over `ResponseWriter`.

---

## Unified clients in server/clients/

**Date**: 2026-02-13
**Status**: Implemented

All client types live under `server/clients/`. Each subtype has its own
subdirectory with `spec.go` (JSON Schema) alongside its runtime:

```
server/clients/
├── provider.go          ← Provider interface + Schema type
├── registry.go          ← Register(), ValidateConfig(), All()
├── executor.go          ← shared logic (webhook + cron)
├── direct/
│   └── spec.go          ← Schema (no runtime)
├── telegram/
│   ├── spec.go
│   └── bot.go
├── webhook/
│   ├── spec.go
│   └── handler.go
└── cron/
    ├── spec.go
    ├── cron.go
    └── scheduler.go
```

The `server/client/` (singular) package was absorbed. The previous separation
(schemas in `client/`, runtime in `clients/`) was not consistent.
The `server/trigger/` package was removed. Webhook and cron are clients just like
Telegram — the previous separation was not consistent with the domain.

---

## JSON Schema validation with google/jsonschema-go

**Date**: 2026-02-13
**Status**: Implemented

Client type config validation uses
[`google/jsonschema-go`](https://github.com/google/jsonschema-go) instead of manual
logic. The library:

- Is from Google, no external dependencies (stdlib only)
- Supports draft-07 and 2020-12 fully (`oneOf`, `const`, `required`, `enum`,
  `pattern`, `minLength`, `if/then/else`...)
- Validates directly on `map[string]any`
- Includes `ApplyDefaults` to populate default values

**Do not use**: manual validators for `required`/`oneOf` or helpers like `matchOneOf`
or `jsonEqual`. Always delegate to the library.

---

## Voice configuration as an independent block

**Date**: 2026-02-14
**Status**: Implemented

Voice-related configuration (UI, ONNX runtime) lives in its own `voice` block
in the YAML, **not** inside `server`. ONNX Runtime is used for voice models of
different types (wake word, VAD, embeddings), so it belongs to the voice domain,
not the HTTP infrastructure domain.

```yaml
voice:
  ui:
    enabled: true # Enable/disable Voice UI, routes and static files
  onnxLibraryPath: "" # Path to libonnxruntime.so (default: /usr/lib/libonnxruntime.so)
```

The Go struct uses sub-structs: `Config.Voice.UI.Enabled` (\*bool, default true)
and `Config.Voice.OnnxLibraryPath` (string).

**Do not put**: voice fields inside `Server` — that block is for network/ports only.

---

## Documentation website (Hugo)

**Date**: 2026-02-14
**Status**: Implemented

Project documentation lives in `website/` as a Hugo static site, deployed to
GitHub Pages. Uses a custom `magec` theme with the project palette: piedra,
atlántico, lava, sol, arena. Dark mode only.

```
website/
├── hugo.toml               ← Site config + sidebar navigation
├── content/docs/           ← Markdown docs (getting-started, install-*, configuration, etc.)
└── themes/magec/           ← Custom theme (layouts, css, js, shortcodes)
```

Build: `cd website && hugo`. Dev: `cd website && hugo server`.

The README.md is simplified — highlights and quick start, pointing to the website
for detailed documentation.

---

## Admin UI never accesses the User API

**Date**: 2026-02-14
**Status**: Implemented

The Admin UI (port 8081) **must never** access the User API (port 8080) to
perform operations. All logic must go directly through internal access
(Go structs, services, stores).

Example: to delete an ADK session, the admin handler calls
`sessionService.Delete()` directly — it does not make HTTP calls to port 8080. To list conversations,
it reads from `ConversationStore` — it does not call REST endpoints.

**Reason**: The admin is an internal component with privileged access. It must not
depend on client authentication (`clientAuthMiddleware`) or the User API's
availability. If the User API is down or misconfigured,
the admin must continue working.

**Do not**: `http.Get("http://127.0.0.1:8080/api/v1/...")` from the admin handler.
Always pass direct references to internal services (session, memory, store).

---

## Centralized memory in the launcher

**Date**: 2026-02-14
**Status**: Implemented

Session and long-term memory configuration is **global**, not per agent.
The ADK launcher accepts a single `session.Service` and a single `memory.Service`,
so configuring memory individually per agent is an illusion — in practice
they all use the same one.

Global config lives in `StoreData.Settings`:

```go
type Settings struct {
    SessionProvider  string `json:"sessionProvider,omitempty"`
    LongTermProvider string `json:"longTermProvider,omitempty"`
}
```

The `AgentDefinition.Memory.Session` and `AgentDefinition.Memory.LongTerm`
fields are kept in the struct for backwards compatibility but are ignored. The UI no
longer shows them in the agent form.

**Do not**: Configure session/longterm memory at the individual agent level.
If ADK improves the launcher to support multiple session services in the future,
it can be decentralized.

---

## Voice config is an agent capability, not a flow property

**Date**: 2026-02-15
**Status**: Implemented

TTS and STT are configured **per agent** (`AgentDefinition.TTS`, `AgentDefinition.Transcription`).
Flows do not and should not have their own voice configuration. A flow orchestrates agents,
and each agent has its own "voice" (TTS) and "ear" (STT), like a person.

Analogy: in a meeting, everyone speaks and understands. The flow decides who participates and in
what order. What voice each one has is intrinsic to the agent, not the flow.

**Do not**: Add voice fields to `FlowDefinition` or `FlowStep`. Do not create a
`voiceAgent` boolean in steps. If an agent participates in a flow and needs voice,
configure it on the agent.

---

## Spokesperson is a consumer (voice-ui) decision, not an admin one

**Date**: 2026-02-15
**Status**: Implemented

When the voice-ui works with a flow, multiple agents can be marked as
`responseAgent` (they speak publicly). The voice-ui needs to know which agent to send
audio to for STT and which agent's voice to use for TTS. This choice is called
**spokesperson** and is a **user preference** of the voice-ui, not an admin
configuration.

Layers of responsibility:

| Layer                 | Responsibility                | Where                   |
| --------------------- | ----------------------------- | ----------------------- |
| Voice capability      | TTS/STT config per agent      | Admin (AgentDefinition) |
| Who responds publicly | `responseAgent` per flow step | Admin (FlowStep)        |
| Who is spokesperson   | Selector among responseAgents | Voice-ui (user chooses) |

The spokesperson is persisted in localStorage by flow ID (`SettingsManager`). If there is no
saved spokesperson, the first `responseAgent` of the flow is used as fallback. If there
is no `responseAgent` marked, the first agent of the flow is used.

**Do not**: Put spokesperson selection in the admin UI. Do not add
`voiceAgent` or `spokesperson` fields to the server data model.

---

## /client/info exposes type and internal agents of flows

**Date**: 2026-02-15
**Status**: Implemented

The `GET /api/v1/client/info` endpoint returns `allowedAgents` with enriched
information so that clients (voice-ui, future ones) can distinguish agents
from flows and know the internal composition:

```json
{
  "allowedAgents": [
    { "id": "...", "name": "Magec", "type": "agent" },
    {
      "id": "...",
      "name": "Software Factory",
      "type": "flow",
      "agents": [
        {
          "id": "...",
          "name": "Architect",
          "type": "agent",
          "responseAgent": true
        },
        {
          "id": "...",
          "name": "Developer",
          "type": "agent",
          "responseAgent": true
        },
        { "id": "...", "name": "Planner", "type": "agent" }
      ]
    }
  ]
}
```

`AgentSummary` fields:

- `type`: `"agent"` or `"flow"` — previously indistinguishable
- `agents`: only in flows, list of unique agents from the tree (via `FlowDefinition.AgentIDs()`)
- `responseAgent`: only in nested agents of a flow, indicates if they are marked as `responseAgent` in some step

**Do not**: Expose TTS/STT config in this endpoint. The voice-ui does not need
to know the technical details — it only needs the agent ID to pass it to the
voice endpoints (`/voice/{agentId}/speech`, `/voice/{agentId}/transcription`).

---

## Voice errors as notifications, not blockers

**Date**: 2026-02-15
**Status**: Implemented

When a spokesperson does not have TTS or STT configured, the voice-ui shows a
friendly notification instead of failing silently or blocking the UI:

- STT fails → "It seems the agent can't understand you. Check that it has transcription configured."
- TTS fails → "It seems the agent can't speak. Check that it has text-to-speech configured."

**Do not**: Filter spokespersons by whether they have voice configured (couples the endpoint
to the concept of voice). Do not block the selection — the user can choose any
responseAgent and receive feedback if it doesn't work.

---

## Admin password authentication (v0.2.0)

**Date**: 2026-02-14
**Status**: Implemented

Admin API (port 8081) is protected by a password configured in `server.adminPassword`.
Authentication uses `Authorization: Bearer <password>` header with constant-time
comparison (`crypto/subtle.ConstantTimeCompare`) and per-IP rate limiting (5 attempts/minute).

If no password is set, the admin remains open (backwards compatible) with a warning log.

The middleware bypasses auth for `OPTIONS` preflight and static files (non-`/api/` paths).
A dedicated `/api/v1/admin/auth/check` endpoint allows the UI to verify credentials
without hitting a real resource.

The Admin UI shows a login screen when auth is required. The password is stored in
memory only (not localStorage) — closing the tab requires re-authentication.

**Do not**: Use cookies or sessions. Do not store the password in localStorage.
Do not use `X-Admin-Password` custom header — use standard `Authorization: Bearer`.

---

## Remote A2A agents use ADK native remoteagent, not custom protocol

**Date**: 2026-02-22
**Status**: Decided (pending implementation)

When connecting to remote A2A agents, magec uses ADK's `agent/remoteagent` package (`remoteagent.NewA2A()`) instead of implementing custom A2A client logic. This gives us:

- Agent card discovery, `tasks/send`, SSE streaming — all handled by the library
- Two composition modes for free: as tool (`agenttool.New(remote)`) for orchestration, or as sub-agent (transfer) for direct user interaction
- Same entity pattern as MCPs: a `RemoteAgent` in the store with URL + credentials, referenced by agent ID

The orchestrator agent (e.g. MetaMagecAgent) is a regular `AgentDefinition` with a system prompt and `remoteAgents` list. From a flow perspective, it's just another step — the flow doesn't know it delegates to remote A2A agents internally.

**Do not**: Implement custom A2A client calls in `callAgent()` or `executor.go`. Do not bypass ADK's remoteagent — it handles protocol details, retries, and streaming.

---

## Multimodal files use inlineData, not fileData

**Date**: 2026-02-22
**Status**: Decided (pending implementation)

When clients (Telegram, Slack) receive files from users, the files are sent to the ADK `/run_sse` endpoint as `inlineData` (base64-encoded bytes + mimetype) in `newMessage.parts[]`, not as `fileData` (URI reference).

**Reasoning**:

- `fileData` with URI is a Gemini-specific concept (Google Files API). OpenAI and Anthropic do not support fetching from URIs — they expect content inline.
- Magec supports all three backend types. `inlineData` is the common denominator that works everywhere.
- Files from Telegram/Slack chats are typically small (photos, screenshots, short PDFs) — base64 overhead (~33%) is acceptable.
- No external storage dependency (S3, MinIO) needed.

**Size limit**: 20MB per file (Telegram bot API limit is 20MB, Gemini inline limit is 20MB, Anthropic is 5MB for images). Enforced client-side before download.

**ADK already supports this**: `genai.Part{InlineData: &Blob{Data: []byte, MIMEType: string}}` deserializes natively from the JSON payload. Zero backend changes.

**Do not**: Use `fileData` for client-uploaded files. Do not add object storage for this use case.
If a future need arises for large files (>20MB), evaluate `fileData` with Gemini Files API as a Gemini-specific path, keeping `inlineData` as fallback for other backends.

---

## Secrets with env var injection and encryption at rest (v0.2.0)

**Date**: 2026-02-14
**Status**: Implemented

Secrets are key-value pairs stored in `StoreData.Secrets`. Each secret has:
`{id, name, key, value, description}` where `key` is the env var name (e.g. `OPENAI_API_KEY`).

**Injection flow**: On store load, secrets are extracted first (raw unmarshal), decrypted
if encrypted, injected via `os.Setenv()`, then the full store is expanded via `os.ExpandEnv()`.
This allows `${OPENAI_API_KEY}` in any store field (backend URLs, API keys, bot tokens).

**Encryption**: When `encryptionKey` is configured in `config.yaml`, secret values are encrypted with
AES-256-GCM, key derived via PBKDF2 (100k iterations, SHA-256). Stored as `enc:v1:<base64>`.
Without encryption key, secrets are stored in cleartext with a warning log.
`encryptionKey` is independent from `adminPassword` — one handles encryption, the other authentication.

**API**: Secrets CRUD at `/api/v1/admin/secrets`. GET responses never include the `value`
field — values are write-only from the API perspective. Updates with empty `value` preserve
the existing value.

**Recovery**: If encryption key is lost, encrypted secrets are unrecoverable. Delete them
and recreate. Non-secret entities remain intact.

**Do not**: Return secret values in GET responses. Do not store secrets in config.yaml.

---

## Message size validation and splitting via shared msgutil package

**Date**: 2026-02-23
**Status**: Implemented

Large inbound messages and oversized outbound responses are handled by a shared utility package `server/clients/msgutil/`. Both Telegram and Slack clients import it — the logic is DRY and testable, while clients remain decoupled from each other and from the ADK API.

**Inbound validation**: `ValidateInputLength(text, maxLen)` truncates messages exceeding 16K runes (unicode-safe) and appends `[message truncated]`. Applied in both clients before calling the agent.

**Outbound splitting**: `SplitMessage(text, maxLen)` breaks responses into platform-safe chunks. Split priority: paragraph boundaries (`\n\n`) > line boundaries (`\n`) > word boundaries (space) > hard cut. Telegram uses 4096, Slack uses 39000.

**Platform constants**:

- `TelegramMaxMessageLength = 4096`
- `SlackMaxMessageLength = 39000`
- `DefaultMaxInputLength = 16000`

**Where validation happens**:

- Telegram: `handleMessage()` validates `msg.Text`, `handleVoice()` validates transcribed text — both before `callAgent()`
- Slack: `processMessage()` validates `text` before building the request — covers both DMs and audio clips (which flow through `processMessage`)
- Voice UI: no splitting needed (browser has no render limit)
- Executor: no splitting needed (returns string to HTTP caller)

**Do not**: Validate or split inside `callAgent()` / the ADK request path. Keep it at the client entry/exit points so each client controls its own limits. Do not add platform-specific logic to the shared package — it only provides generic split/validate functions with configurable limits.

---

## 17. Artifact Toolset — Universal via Base Toolset, No Delete

**Date**: 2025-02-23

All agents get the artifact toolset (save/load/list) unconditionally via `base_toolset.go`, not opt-in per agent. This avoids config complexity and ensures every agent can produce files for users.

**No delete tool**: ADK's `agent.Artifacts` interface (exposed via `tool.Context`) has Save, Load, List, and LoadVersion — but no Delete. Delete exists only on `artifact.Service` directly. Rather than breaking the abstraction by passing the raw service into tools, we omit delete. Artifacts are versioned and session-scoped, so stale artifacts are naturally cleaned up when sessions expire.

**Storage**: `adk-utils-go/artifact/filesystem` — filesystem-backed `artifact.Service` implementation. Stores artifacts as JSON at `data/artifacts/{appName}/{userID}/{sessionID}/{fileName}/{version}.json`. Supports versioning and user-scoped artifacts. Data persists across restarts.

**Client delivery**: Telegram and Slack clients list artifacts before and after each `/run` call, diff the lists, and deliver new artifacts as file attachments (Telegram: `SendDocument`, Slack: `UploadFileV2`). Artifacts are always files, never inlined in chat text.

**Files**: `server/agent/tools/artifacts/toolset.go` (toolset), `server/agent/base_toolset.go` (wiring), `server/agent/agent.go` (FilesystemService + launcher config), `server/clients/telegram/bot.go` and `server/clients/slack/bot.go` (delivery).

---

## 18. Multimodal Adapter Parity — Error on Unsupported Types

**Date**: 2026-02-23

When an adapter receives `genai.Part{InlineData}` with a MIME type it can't translate, it returns an error — **not** `nil` (silent drop). This matches Gemini's native behavior where unsupported types cause the API request to fail.

**Rationale**: Silent drops are a bug — the user sends a file, the LLM never sees it, and nobody gets feedback. With errors, either the client validates beforehand (preferred) or the user sees an explicit failure. All three providers behave identically: unsupported = fail.

**Supported types per adapter (adk-utils-go v0.3.1)**:

| Type                          | Gemini      | OpenAI          | Anthropic              |
| ----------------------------- | ----------- | --------------- | ---------------------- |
| Images (JPEG, PNG, GIF, WebP) | ✅ (native) | ✅ (data URI)   | ✅ (Base64ImageSource) |
| PDF                           | ✅ (native) | ✅ (FileParam)  | ✅ (Base64PDFSource)   |
| Text (text/\*)                | ✅ (native) | ✅ (FileParam)  | ✅ (PlainTextSource)   |
| Audio (WAV, MP3, WebM)        | ✅ (native) | ✅ (InputAudio) | ❌ error               |
| Video, other                  | ✅ (native) | ❌ error        | ❌ error               |

**Do not**: Silently drop unsupported `InlineData` parts. Do not convert them to text descriptions. Return `fmt.Errorf("unsupported inline data MIME type for %s: %s")`.

**Files**: `adk-utils-go/genai/openai/openai.go` (`convertInlineDataToPart`), `adk-utils-go/genai/anthropic/anthropic.go` (`convertInlineDataToBlock`).

---

## 19. InstructionProvider bypasses ADK's `{variable}` substitution

**Date**: 2026-03-12
**Status**: Implemented

ADK's `InjectSessionState` automatically replaces `{anything}` in agent instructions with session state values. This breaks any prompt containing curly braces — JSON examples, scripts, code snippets, templates, etc. The regex (`{+[^{}]*}+`) is hardcoded with no escape mechanism and no configuration options.

**Solution**: All agents use `InstructionProvider` (a callback) instead of the static `Instruction` string. When `InstructionProvider` is set, ADK skips `InjectSessionState` entirely — curly braces become plain text.

**Custom substitution syntax**: `{{agent.output:variable_name}}`. This is the only pattern Magec resolves from session state. The regex `\{\{agent\.output:([a-zA-Z_][a-zA-Z0-9_]*)\}\}` is specific enough to never collide with real prompt content.

**Fast path**: If the instruction contains no `{{agent.output:` patterns, the provider returns the string as-is with zero regex overhead.

**Flow output keys**: Work exactly as before — an agent's `outputKey` saves to session state, and downstream agents reference it with `{{agent.output:key}}` in their system prompt. The `SessionStateSeed` middleware pre-seeds empty values so first-invocation doesn't fail.

**Do not**: Use ADK's static `Instruction` field on `llmagent.Config`. Always use `InstructionProvider` via `makeInstructionProvider()`. Do not introduce new substitution patterns — `{{agent.output:key}}` is the only one.

**Files**: `server/agent/agent.go` (`makeInstructionProvider`, `stateVarRegex`).

---

## 20. Recursive snake_case→camelCase normalization for ADK REST API

**Date**: 2026-04-09
**Status**: Implemented

ADK's REST decoder calls `json.Decoder.DisallowUnknownFields()`, which rejects any JSON key that doesn't match a struct tag exactly. All ADK and genai JSON tags are camelCase (`appName`, `sessionId`, `inlineData`, `mimeType`, etc.) with zero exceptions. API clients sending snake_case keys get a 400.

**Solution**: `SnakeCaseNormalize` middleware intercepts POST `/run` and `/run_sse`, parses the JSON body, and recursively converts all snake_case keys to camelCase at every nesting level. Generic `snakeToCamel` conversion (not a hardcoded key list) so it handles any current and future fields without maintenance.

**Key behaviors**:

- When both `app_name` and `appName` coexist in the same object, camelCase wins (explicit client intent)
- Single-word keys (`text`, `role`, `parts`, `data`) are never modified
- If the body is not valid JSON or already all camelCase, the original bytes pass through unchanged
- `Content-Length` is updated after rewriting

**Scope**: Only `/run` and `/run_sse` — the only ADK endpoints with multi-word JSON body fields. Session create uses single-word fields (`state`, `events`). Session/artifact paths use URL parameters (already snake_case by ADK design).

**Middleware position**: Outermost in the chain, before `ConversationRecorder`, so all downstream middlewares see the normalized camelCase body.

**Do not**: Use a fixed key list (fragile, misses nested genai fields). Do not normalize non-`/run` paths (unnecessary, could interfere with other handlers). Do not modify response bodies — only request normalization.

**Files**: `server/middleware/normalize.go`, `server/main.go` (wiring).

---

## 21. Voice provider registry for multi-backend TTS/STT

**Date**: 2026-04-09
**Status**: Implemented

TTS and STT proxies were hardcoded to OpenAI-compatible endpoints (`/v1/audio/speech`, `/v1/audio/transcriptions`). A Gemini backend assigned to TTS would fail because Gemini doesn't serve those endpoints.

**Solution**: `voice.Provider` interface with per-backend-type implementations. Same registry pattern as clients and memory — `init()` + blank imports in `main.go`. The proxy resolves the provider via `voice.Get(backend.Type)` and delegates.

**Providers**:

- **OpenAI** (`voice/openai/`) — extracted from the previous inline code in `main.go`. Passthrough to `/v1/audio/speech` and `/v1/audio/transcriptions`.
- **Gemini** (`voice/gemini/`) — translates to `generateContent` API with `speechConfig` for TTS, `inlineData` for STT. Handles PCM→WAV wrapping, base64 encoding/decoding, and API key auth via query parameter.

**Provider-specific config**: `TTSRef.Config` and `BackendRef.Config` use typed Go structs (`TTSConfig`, `STTConfig`) with per-provider namespaces — same pattern as `ClientConfig` (e.g. `config.telegram`, `config.discord`). Each provider has its own struct: `OpenAITTSConfig` (`speed`), `GeminiTTSConfig` (`languageCode`, `temperature`, `stylePrompt`). The `speed` field was removed from `TTSRef` (it was OpenAI-specific) and moved into `config.openai.speed`. A store migration in `loadFromDisk` handles the transition from the old format.

**Admin API**: `GET /voice/types` returns all registered providers with their schemas, same pattern as `/clients/types` and `/memory/types`.

**Do not**: Hardcode endpoint paths in the proxy handlers. Do not add provider-specific logic to `main.go` — all translation lives in the provider package. Do not put provider-specific fields as top-level fields on `TTSRef` or `BackendRef` — use the typed config structs.

**Files**: `server/voice/provider.go`, `server/voice/registry.go`, `server/voice/openai/openai.go`, `server/voice/gemini/gemini.go`, `server/api/admin/voice.go`, `server/main.go` (proxy refactor + blank imports), `server/store/types.go` (typed config structs), `server/store/store.go` (migration), `frontend/admin-ui/src/views/agents/AgentDialog.vue`.

---

## 22. Conversation close on reset is a middleware concern, not a client one

**Date**: 2026-04-18
**Status**: Implemented (since 75cc2de, 2026-04-10)

When a chat client (Telegram/Slack/Discord) or the admin UI triggers a session reset, the corresponding conversation record must be marked `Closed` so the next message starts a fresh conversation instead of appending to the old one.

This responsibility lives in **one place only**: the `ConversationRecorder` middleware in `server/middleware/recorder.go`. Every `DELETE /api/v1/agent/apps/{app}/users/{user}/sessions/{session}` is intercepted regardless of origin; after the delete succeeds the middleware calls `executor.CloseConversationSession(sessionID, agentID)`, which internally closes both the `admin` and `user` perspectives via `ConversationStore.CloseBySession`.

**Do not**:

- Re-implement `CloseBySession` inside individual client bots (Telegram `handleResetCommand`, Slack/Discord reset command handlers). The middleware already handles it transparently for every session delete, including those triggered by the admin `/conversations/{id}/reset-session` endpoint.
- Add `SetConversationStore` methods to `Client` structs. The conversation store stays behind the executor/middleware boundary — clients have no business touching it.
- Duplicate the `admin`+`user` perspective loop. It belongs in `Executor.CloseConversationSession`, a single function.

**Files**: `server/middleware/recorder.go` (interception), `server/clients/executor.go` (`CloseConversationSession`), `server/store/conversations.go` (`Closed` flag + `CloseBySession` + `FindBySession` skipping closed records).

**History**: Fix was introduced by commit `75cc2de` on 2026-04-10. A later audit attempted to re-add the close-on-reset logic inside each client bot; this was reverted in favour of the original middleware-based design.

---

## 22. ADK REST API construction via `adkrest.NewServer`

**Date**: 2026-04-18
**Status**: Implemented

`google.golang.org/adk` v1.0.0 removed `adkrest.NewHandler(launcher.Config, timeout)` in favour of `adkrest.NewServer(adkrest.ServerConfig{...})`, which takes each service directly instead of a `launcher.Config` wrapper. This matches the broader v1.0.0 refactor that decoupled `adkrest` from `cmd/launcher`.

`*agent.Service` still exposes `Handler() http.Handler`; middleware does not care whether the underlying implementation is `*adkrest.Server` or anything else satisfying the interface.

**Do not**: Reintroduce `launcher.Config` as an input to `adkrest`. The dependency is being removed upstream — keep each concrete service (session, memory, artifact, loader) as the source of truth.

**Files**: `server/agent/agent.go`.
