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

## 23. ADK REST API construction via `adkrest.NewServer`

**Date**: 2026-04-18
**Status**: Implemented

`google.golang.org/adk` v1.0.0 removed `adkrest.NewHandler(launcher.Config, timeout)` in favour of `adkrest.NewServer(adkrest.ServerConfig{...})`, which takes each service directly instead of a `launcher.Config` wrapper. This matches the broader v1.0.0 refactor that decoupled `adkrest` from `cmd/launcher`.

`*agent.Service` still exposes `Handler() http.Handler`; middleware does not care whether the underlying implementation is `*adkrest.Server` or anything else satisfying the interface.

**Do not**: Reintroduce `launcher.Config` as an input to `adkrest`. The dependency is being removed upstream — keep each concrete service (session, memory, artifact, loader) as the source of truth.

**Files**: `server/agent/agent.go`.

---

## 24. User attachments are always persisted as session artifacts instead of inlined

**Date**: 2026-04-19
**Status**: Implemented

Previously, when a user attached files to a chat message (Telegram photos/documents, Slack file uploads, Discord attachments), the client embedded them inline in the `/run_sse` request body as `inlineData` parts. Large files — a 20MB PDF, a multi-megabyte screenshot — burn context tokens even when the model does not need to read them, and sometimes overflow the backend's request size limit. A size threshold split the logic, but it proved insufficient as even small PDFs could exhaust the context window.

**Decision**: All files are now persisted via the ADK `artifact.Service` with a deterministic per-message filename (`{messageID}_{originalName}`). The prompt is annotated with a `MAGEC_ATTACHED_ARTIFACTS` HTML-comment block listing each file's name, MIME type and human size, with instructions telling the LLM to call `load_artifact` on demand.

The artifact toolset (`save_artifact`/`load_artifact`/`list_artifacts`) is already universal (decision #17), so no tool wiring changes were needed. The artifact service is plumbed into each client through `SetArtifactService(artifact.Service)` and sourced from `agentRouterHandler.ArtifactService()`, which refreshes on every store rebuild.

**Helper shape (DRY real, not over-engineered)**:

Rather than a single `PrepareAttachments(...)` function with many parameters and a fat return struct, the helpers are broken into one-job primitives in `server/clients/msgutil/attachments.go`:

- `StoreAsArtifact(ctx, svc, appName, userID, sessionID, filename, mimeType, data) (descriptor, error)` — persists via `artifact.Service.Save` and returns a single descriptor line.
- `AttachedArtifactsBlock(lines []string) string` — wraps lines in the `MAGEC_ATTACHED_ARTIFACTS` HTML-comment block; returns `""` on empty input so callers can concatenate unconditionally.

Each client keeps its own short loop iterating over platform-specific attachment types (`msg.Photo/Document/...`, `ev.Message.Files`, `m.Attachments`). That asymmetry is inherent to each platform — hiding it behind a generic "prepare everything" call would leak platform details back into the helper API.

**Fallback**: when `artifacts` is `nil` (no agents configured yet, or the feature is explicitly disabled), the files are dropped with a warning. This ensures safety.

**Graceful degradation**: if `StoreAsArtifact` fails (disk error, service down), the client logs a warning and drops the file — the text part of the user message is still delivered.

**Rationale**:

- Token budget: binary files or text blobs are almost always checkpoint data the model consults occasionally, not content it needs in every turn.
- Backend compatibility: OpenAI, Anthropic and Gemini all reject requests above their respective body caps. Artifact offloading keeps the path uniform regardless of provider.
- Reuses decision #17 universal artifact toolset — no new tools or wiring.

**Do not**:

- Reintroduce an inline path or size thresholds. All attachments should flow through the artifact system to protect the context window.
- Replace the single-purpose helpers with a fat `PrepareAttachments` returning a multi-field struct. The current shape is DRY where it matters (the ADK wire format, the prompt block, the save call) without coupling the three platform-specific loops.

**Files**: `server/clients/msgutil/attachments.go` (helpers + tests), `server/agent/agent.go` (`ArtifactService()` getter), `server/main.go` (router plumbing, `clientManager.artifactService()`), `server/clients/telegram/bot.go`, `server/clients/slack/bot.go`, `server/clients/discord/bot.go`.

---

## 25. `load_artifact` injects content via RequestProcessor, not as JSON base64

**Date**: 2026-04-30
**Status**: Implemented

The previous `load_artifact` tool returned binary artifact contents as base64 strings inside a `LoadResult` JSON struct. That value travelled through ADK's standard FunctionResponse path, meaning the LLM saw a wall of base64 text it could not interpret. For PDFs and images this defeated the entire artifact-offloading design (decision #24): the user attached a PDF, the model called `load_artifact` to read it, and instead of getting a multimodal attachment it got megabytes of opaque text that exploded the context window.

**Decision**: `load_artifact` is no longer a plain `functiontool`. It implements ADK's `tool.Tool` interface manually plus the duck-typed `toolinternal.RequestProcessor` interface (matched at runtime via `t.(RequestProcessor)` in ADK's flow):

- `Run(ctx, args)` returns only metadata (`{success, name, message}`). It does NOT load the file. This prevents binary content from leaking into the FunctionResponse.
- `ProcessRequest(ctx, req)` runs on the next LLM turn. It detects the prior `load_artifact` FunctionResponse, calls `ctx.Artifacts().Load(name)`, and appends a fresh user-role `*genai.Content` with the original `*genai.Part` (text or `InlineData`) into `req.Contents`.

The provider adapters (`adk-utils-go/genai/{openai,anthropic,gemini}`) then translate that Part into their native multimodal format (FileParam / Base64PDFSource / inlineData). The model reads the file as a real attachment, never as base64 text.

`save_artifact` and `list_artifacts` remain plain function tools — they only deal with metadata, no binary payload to deliver.

**Why not use ADK's official `loadartifactstool`**: the official tool batches multiple artifacts per call and prepends a global "you have artifacts X, Y, Z, you must load them before answering" instruction every turn via `utils.AppendInstructions`. Magec already manages attachment hints itself through the `MAGEC_ATTACHED_ARTIFACTS` block in the user prompt (decision #24), and surfaces saved artifacts through client diffing. The official tool would duplicate that logic and pollute the system instructions on every request even when no artifacts exist. Our single-file `load_artifact` mirrors the same RequestProcessor mechanism but stays scoped to one explicit call at a time.

**Do not**:

- Reintroduce a JSON path that returns artifact bytes as base64 to the model. The whole point of artifact offloading is to keep binaries out of context until the model decides to read them, and even then to deliver them through the multimodal channel.
- Wrap binary payloads in custom JSON structs from any future tool. If a tool needs to deliver binary or multimodal content to the LLM, it must follow this same pattern: lightweight metadata in `Run`, real `*genai.Part` injection in `ProcessRequest`.
- Depend on ADK's `internal/toolinternal` package directly. The `RequestProcessor` interface is matched structurally at runtime; declaring the method with the right signature is enough.

**Files**: `server/agent/tools/artifacts/toolset.go` (loadArtifactTool), `server/agent/agent.go` (`artifactInstruction` rewording), `server/clients/msgutil/attachments.go` (`AttachedArtifactsBlock` rewording).

---

## 26. `export_artifact` bridges artifacts and the local filesystem; `Store.ResolveTemporaryDir` is the single source of truth

**Date**: 2026-05-01
**Status**: Implemented

Artifacts live in their own world: `data/artifacts/{appName}/{userID}/{sessionID}/{fileName}/{version}.json`, with binary payloads encoded as base64 inside JSON. That layout is invisible to any external tool that reads from the filesystem — a parser, a shell utility, an MCP server pointed at a workdir. There was no clean way for the model to hand an artifact off to such a tool.

**Decision**: a new function tool, `export_artifact`, decodes an artifact through `ctx.Artifacts().Load(name)`, writes the raw bytes (or the UTF-8 text for text artifacts) to a fresh file under a single, centrally-configured directory, and returns the absolute path. From that point on, any other tool reading from disk can pick the file up. The model only learns about an absolute path; it never proposes the destination directory.

**Single resolution point**: the directory is owned by `Store.ResolveTemporaryDir()`. That method is the **only** place where the fallback `Settings.TemporaryDir → os.TempDir()` is performed. Any future tool or subsystem that needs a transient on-disk location must call this method; nobody else may read `Settings.TemporaryDir` directly nor recompute the fallback elsewhere. `Settings.TemporaryDir` is exposed in the Admin UI Settings → Runtime section so the operator can pin the directory to whatever path their filesystem-aware tools (MCP servers, shells, scripts) are allowed to read from.

**Decoupling**: the tool does not reach into the store. `agent.New` accepts a `tempDirProvider func() string` parameter, the caller (`main.go` `agentRouterHandler.rebuild`) wires `dataStore.ResolveTemporaryDir` into it, and the closure flows down to the artifact toolset constructor. Each `export_artifact` call invokes the provider, so changes to `Settings.TemporaryDir` take effect on the next call without rebuilding the toolset (the agent rebuild on store change still happens for other reasons, but isn't required to refresh this path).

**Filename collisions**: `os.CreateTemp(dir, stem+"-*"+ext)` injects a random suffix between the original stem and extension (e.g. `report-1234567890.pdf`). Two parallel exports of the same artifact never collide and downstream tools can still infer the format from the extension.

**Description policy**: the tool description deliberately makes no mention of any specific MCP server or filesystem integration. It says only "writes the artifact's bytes to a file on disk and returns the absolute path so other tools can read it". Whether the operator has wired a filesystem MCP, a shell tool, or none at all is orthogonal to the tool's contract.

**Do not**:

- Do not duplicate the `TemporaryDir → os.TempDir()` fallback anywhere outside `Store.ResolveTemporaryDir`. Any caller that reads `Settings.TemporaryDir` directly is a bug.
- Do not let the model pass a destination path to `export_artifact`. The argument set is intentionally `{name}` only; the tool decides the directory and filename. Allowing a path opens up traversal and surprises ("where did my file end up?").
- Do not couple the artifact toolset to the Store. The `tempDirProvider` closure pattern keeps the tool ignorant of any storage concern.
- Do not reference specific external integrations (MCP filesystem, shell, etc.) in the tool's `Description`. The tool must read as a generic capability so that deployments without those integrations still see a coherent contract.
- Do not add cleanup logic for the temp directory inside Magec. When `TemporaryDir` is unset, the OS handles `os.TempDir()` cleanup. When it's set to a custom path, the operator owns retention.

**Files**: `server/store/types.go` (`Settings.TemporaryDir`), `server/store/store.go` (`ResolveTemporaryDir`), `server/agent/tools/artifacts/toolset.go` (`exportArtifact` + `tempDirProvider` plumbing), `server/agent/base_toolset.go`, `server/agent/agent.go` (`tempDirProvider` parameter, `artifactInstruction` updated), `server/main.go` (provider wired from `dataStore.ResolveTemporaryDir`), `frontend/admin-ui/src/views/settings/RuntimeSection.vue`, `frontend/admin-ui/src/views/settings/SettingsView.vue`.

---

## 27. Ephemeral signed URLs for artifacts (`/api/v1/ephemeral/artifacts/{token}`)

**Date**: 2026-05-01
**Status**: Implemented

`export_artifact` (decision #26) bridges the artifact world and the local filesystem when the consumer runs in the same container. That assumption breaks the moment a consumer lives in a sidecar (filesystem MCP server, shell tool runner, anything that fetches files over HTTP). Sharing volumes between containers is environment-specific and unfriendly to k8s deployments without RWX storage; we wanted a transport that works whenever there is just network connectivity between the consumer and Magec.

**Decision**: a new endpoint `GET /api/v1/ephemeral/artifacts/{token}` serves an artifact's raw bytes when called with a valid token, with no other authentication. The token is a JWT-shaped, HMAC-SHA256-signed envelope that carries the full descriptor of the artifact (`appName`, `userID`, `sessionID`, `name`) plus an absolute Unix expiration. The companion tool `get_artifact_url` mints those URLs from the model's session context.

**Namespace**: `/api/v1/ephemeral/...`. Reserved for any future resource that becomes accessible through a signed, short-lived URL (e.g. signed conversation exports, temporary skill-package downloads). This namespace name describes the **observable property** (the URL caducates) rather than the mechanism (HMAC signing), so it stays meaningful if we add other expiration drivers later (single-use, quota, revocation).

**Token shape**: `base64url(payload).base64url(hmac_sha256(payload, secret))`. JWT-like but stripped of algorithm negotiation; HMAC-SHA256 is the only supported algorithm and is not encoded inside the token. This keeps the verifier trivial and removes the alg-confusion class of vulnerabilities. Implementation lives in `server/ephemeral/`.

**TTL**: 1 hour, hardcoded (`ephemeralArtifactURLTTL`). One hour comfortably absorbs slow downstream consumers (HTTP retries, multi-step model reasoning, cold MCP startup) without leaving signed URLs alive long enough to matter if they leak. Not configurable until a concrete deployment asks for it (KISS).

**Signing secret**: `server.encryptionKey` from `config.yaml`. Reused on purpose: it is already a long-lived, persisted secret with operator semantics ("encrypt secrets at rest"), and the ephemeral-URL HMAC uses it the same way (sign payloads at rest in the URL). Tokens minted before a magec restart keep working as long as `encryptionKey` is unchanged. When `encryptionKey` is empty, the endpoint returns 503 and the `get_artifact_url` tool is **not registered** at all, instead of being advertised and always failing.

**Endpoint placement**: user API (port 8080), top-level under `/api/v1/`. Bypasses `ClientAuth` for the same reason `/api/v1/webhooks/*` does: the URL itself is the credential. Listed alongside webhooks in the bypass list of `middleware/middleware.go`.

**Decoupling**: the `artifact_toolset` does not read configuration. `agent.New` accepts an `artifactURLBuilder toolsartifacts.ArtifactURLBuilder` parameter. `main.go` builds the closure from `cfg.Server.EncryptionKey` + `getA2APublicURL(cfg)` and stores it on `agentRouterHandler` so it survives store rebuilds. The toolset only ever sees a closure that takes an `ArtifactURLRequest` (app/user/session/name/mime) and returns `{URL, ExpiresAt}`. `app/user/session` are sourced from `tool.Context` so the model cannot mint URLs for foreign sessions.

**Endpoint behaviour**: serves binary artifacts as raw bytes with `Content-Type` taken from `inlineData.MIMEType` (defaulting to `application/octet-stream`). Text artifacts come back as `text/plain; charset=utf-8`. Both responses set `Content-Disposition: attachment; filename="..."` so `curl -O` and `wget` pick a sensible local name. Response statuses: 200 success, 401 invalid/expired token, 404 missing/malformed token or artifact-not-found, 503 secret unconfigured, 410 if the artifact exists but has empty content.

**Public URL**: reuses `getA2APublicURL(cfg)`. Operators that already configured `server.publicURL` for A2A get ephemeral URLs for free. The same hostname must be reachable from any consumer that calls `get_artifact_url`'s URL — keeping a single public URL instead of two avoids the "internal vs external URL" rabbit hole until somebody actually needs it.

**Do not**:

- Do not reintroduce algorithm negotiation in the token (no JWT `alg` field). HMAC-SHA256 is fixed.
- Do not let the model pass `appName`, `userID` or `sessionID` to the tool. They come from `tool.Context` exclusively. Allowing them as arguments would let an agent mint URLs for foreign sessions if a wrong client token were reused elsewhere.
- Do not silently fall back to a generated-in-memory secret when `encryptionKey` is unset. Restarting magec would invalidate every minted URL silently; refusing to mint URLs is louder and clearer.
- Do not reuse `Settings.TemporaryDir` for ephemeral URLs. The two features serve different topologies (local fs vs HTTP); collapsing them obscures the choice the model has to make.
- Do not paste the bearer token of a client into an ephemeral URL pipeline as a workaround. The whole point of `/api/v1/ephemeral/` is that the URL is self-authenticating; mixing client tokens reintroduces the auth coupling we wanted to avoid.

**Files**: `server/ephemeral/ephemeral.go` + tests (Sign/Verify primitives), `server/main.go` (`newEphemeralArtifactHandler`, `newArtifactURLBuilder`, `ephemeralArtifactURLTTL`, mux registration, `agentRouterHandler.artifactURLBuilder` field), `server/middleware/middleware.go` (ClientAuth bypass), `server/agent/tools/artifacts/toolset.go` (`get_artifact_url` tool, `ArtifactURLBuilder` injection), `server/agent/agent.go` (parameter wiring + `artifactInstruction` updated), `server/agent/base_toolset.go`, `server/api/user/handlers.go` (`EphemeralArtifact` swagger stub).

---

## 28. Flow-shared state and loop exit control (issue #36)

**Date**: 2026-05-04
**Status**: Implemented

Agents inside a flow used to be "fire and forget" — they ran in the order the flow declared them, with no way for an agent to read what an earlier sibling produced (other than the system-prompt-only `{{agent.output:key}}` substitution, decision #19), and no way to influence the surrounding loop. This decision adds three capabilities, all scoped strictly to agents that participate in a flow.

### Shared state via the `flow:` prefix on session.state

`set_state(key, value)` and `get_state(key)` are wired into every agent that runs inside any flow (sequential, parallel, loop, nested). Both tools route through `tool.Context.State()`, which writes to the current event's `StateDelta` and through to `session.Service`. Keys are stored under the `flow:` prefix internally; the LLM sees plain keys.

Why session.state and not a parallel "Properties" namespace: ADK already propagates `session.state` synchronously between sub-agents in workflow agents (`runner.AppendEvent` applies `Actions.StateDelta` after every event, before the next sub-agent reads). Adding a parallel store would duplicate the mechanism. The prefix isolates flow-shared keys from ContextGuard's summary keys and from `outputKey` writes.

Why scope-restricted: standalone agents (those reachable directly by ID, not through a flow) get no state tools. They have no peers to share with; the tools would be noise in their catalogue.

### LLM-driven loop exit (`exit_loop`)

When a loop step has `exitLoop: true`, every agent in the loop's subtree (any depth, propagated through nested sequentials/parallels) receives ADK's native `exit_loop` tool from `google.golang.org/adk/tool/exitlooptool`. Calling it sets `event.Actions.Escalate=true`, which the surrounding `loopagent` already honours. The current iteration completes — ADK's `sequentialagent` and `parallelagent` ignore Escalate, only `loopagent` reacts — and the loop terminates before the next iteration.

### State-driven loop exit (`exitWhen`)

When a loop step has a non-empty `exitWhen` string, the flow builder appends a synthetic evaluator agent as the last child of the `loopagent`. After every iteration the evaluator reads the `flow:` subset of session state, evaluates the operator-supplied CEL expression, and emits an `Escalate` event when the expression returns true. Standard ADK loop semantics handle the actual termination.

CEL was picked over a hand-rolled DSL because it is the de-facto Google evaluation language (Kubernetes `ValidatingAdmissionPolicies`, IAM, Envoy), explicitly designed for compile-once / evaluate-many pipelines, thread-safe, type-aware enough to reject expressions that don't return bool, and expressive without inviting Turing-complete computation. Cost: 6 direct dependencies (Google + ANTLR), ~3MB binary.

The `state` variable exposed to CEL is a `map<string, dyn>` populated from session state; each session-state key under the `flow:` prefix appears as `state.<key>` in the expression. CEL runtime errors (missing key, type mismatch) are treated as `false` rather than aborting the conversation; `MaxIterations` remains the hard safety cap.

### Mutual exclusion in admin API and UI

`exitLoop` and `exitWhen` are mutually exclusive on the same loop step. The runtime would tolerate both, but two ways to express the same intent confuse operators. The admin API rejects flows that set both with a 400; the UI radio-button enforces one strategy at a time.

### Iteration-boundary semantics

Both exit mechanisms fire at the **end of an iteration**, not mid-iteration. If `exit_loop` is called by an agent in the middle of a sequential, the rest of the sequential still runs to completion before the loop terminates. This matches ADK's loopagent behaviour and keeps each iteration a coherent unit. Operators who need stop-immediate behaviour can restructure the flow so the decider is the last agent in the sequence.

### Per-appearance agent instances inside flows

To inject scope-dependent tools (state always, `exit_loop` conditionally) without polluting the standalone agent catalogue, `flow.go` builds a fresh ADK agent instance per appearance via `BuildAgentInstance` rather than reusing pre-built agents from a shared map. The standalone catalogue still holds one instance per `AgentDefinition` for direct invocation. Flow-as-step composition (a flow that references another flow as one of its leaves) keeps the previous behaviour — the referenced flow agent is reused via `wrapAgent`.

### Do not

- Do not add the state/exit_loop tools to standalone agents. They should appear only when the agent runs as part of a flow.
- Do not introduce a parallel "Properties" namespace separate from session.state. The prefix is enough; duplicating ADK's storage layer is a maintenance trap.
- Do not allow `exitLoop` / `exitWhen` on non-loop steps. Reject in the admin API.
- Do not let a CEL runtime error abort the conversation. Treat as `false` and log warn; `MaxIterations` is the safety net.
- Do not extend the CEL environment with extra variables beyond `state` without revisiting decision #28. Adding more variables broadens the attack/confusion surface.
- Do not let the model pick the `flow:` prefix or write to other tier prefixes (`app:`, `user:`, `temp:`). The toolset enforces the prefix; the key validator rejects colons.
- Do not implement mid-iteration exit (cancel siblings of the caller). It contradicts ADK's loop semantics and opens a cancellation rabbit hole on parallel branches.

**Files**:

- `server/agent/tools/flowstate/toolset.go` + tests — `set_state`/`get_state`.
- `server/agent/flowexit/compiler.go`, `evaluator.go` + tests — CEL compilation + synthetic evaluator agent.
- `server/agent/agent.go` — `BuildAgentInstance` (renamed from `buildSingleAgent`, now exported and accepting extras), `flowStateInstruction` and `exitLoopInstruction` constants, wiring in `New`.
- `server/agent/flow.go` — `FlowBuildDeps` carrying tool singletons, per-appearance instance build, `insideLoopWithExitLoop` propagation, evaluator agent appended to loops with `exitWhen`.
- `server/store/types.go` — `FlowStep.ExitLoop`, `FlowStep.ExitWhen`.
- `server/api/admin/flows.go` + tests — validation: exclusion, CEL compile, non-loop rejection.
- `frontend/admin-ui/src/views/flows/LoopConfigDialog.vue` — dialog with three exit strategies.
- `frontend/admin-ui/src/views/flows/FlowBlock.vue` — badge with strategy hint, dialog wiring, type-cycle cleans loop-only fields.
- `frontend/admin-ui/src/views/flows/FlowDialog.vue` — help text covering state and loop exit.
- `website/content/docs/flows.md` — public docs: state tools, loop exit strategies, iteration-boundary semantics.

---

## 29. Skills as on-disk packages backed by ADK skilltoolset

**Date**: 2026-05-09
**Status**: Implemented

Skills used to be stored as JSON inside `data/store.json` (full `Instructions` body, plus a `References[]` index pointing at flat files under `data/skills/{id}/`). The agent builder concatenated every linked skill — instructions plus inlined reference files — into the agent's system prompt at build time. Two problems with that:

- **Context bloat**: every linked skill burnt prompt tokens on every turn, even when the model would not have used the skill.
- **Format drift**: Magec's storage shape was a Magec-specific dialect, while the rest of the ecosystem (ADK, Agent Skills spec, third-party skill packages) standardises on a `SKILL.md` file with YAML frontmatter and `references/`/`assets/`/`scripts/` sub-directories.

**Decision**: align entirely with the upstream Agent Skills layout. Skills are real on-disk packages under `data/skills/{slug}/` and are exposed to agents through ADK's `tool/skilltoolset` rather than concatenated into the system prompt.

### On-disk layout

```
data/skills/{slug}/
├── SKILL.md            (YAML frontmatter + Markdown body)
├── references/         (optional, sub-tree of any depth)
├── assets/             (optional)
└── scripts/            (optional)
```

`{slug}` matches the SKILL.md frontmatter `name`; ADK's `FileSystemSource` enforces that invariant, so the Magec admin layer must keep them in sync.

### Store shape

```go
type Skill struct {
    ID   string  // immutable UUID, used by AgentDefinition.Skills[]
    Slug string  // on-disk directory == frontmatter.name
}
```

That's it — **no name, description, instructions or references in the store**. Everything else is read live from disk on every admin GET. This eliminates the "did I update the store or the file?" class of bug at the cost of a tiny per-request read of `SKILL.md`.

### Admin API: one upload endpoint, no manual edit

| Method | Path | Hace |
|---|---|---|
| `GET`    | `/skills`             | List `{id, slug, name, description}` per skill, with `name`/`description` read from each SKILL.md frontmatter. |
| `GET`    | `/skills/{id}`        | Full `SkillView`: frontmatter, instructions body, resources list (kind + relative path + size). |
| `POST`   | `/skills/upload`      | Create or replace a skill from a SKILL.md or a `.zip`/`.tar.gz` package. `?replace=true` overwrites an existing skill that owns the same slug, preserving its store ID so agent links stay valid. Conflict without `replace` returns 409 with `{existingId, slug}`. |
| `GET`    | `/skills/{id}/download` | Streams the on-disk directory as `tar.gz`. Re-uploading the produced archive via the same endpoint reconstructs the skill verbatim. |
| `DELETE` | `/skills/{id}`        | Removes the store record and the directory. |

There is intentionally no manual create/edit endpoint and no individual file upload to a live skill. The contract is **all-or-nothing**: upload a valid SKILL.md (with frontmatter ADK accepts) or a packaged archive, and that becomes the skill's content. Operators who want to tweak one file edit their package locally and re-upload with `replace=true`. That keeps the admin surface tiny and the on-disk truth uncontested.

### Per-agent scope

ADK's `skilltoolset` operates over a `skill.Source`. Magec uses ADK's `skill.NewFileSystemSource` directly — no custom parser, no custom format adapter — and wraps it in two thin layers. The outer one is `agent/tools/skills.AgentFS`, an `fs.FS` that filters `os.DirFS(data/skills)` down to the slug whitelist of the linked skills (so an agent only sees what the operator gave it). The inner one is `agent/tools/skills.TolerantSource`, which proxies `skill.Source` and falls back to a permissive frontmatter parse when ADK's strict `KnownFields(true)` validator rejects a SKILL.md because of non-canonical keys (`version:`, `author:`, `tags:`). Both layers are thin: Magec never owns the SKILL.md format, ADK does. Every `BuildAgentInstance` call constructs a fresh `skilltoolset` for its agent if `Skills[]` is non-empty; otherwise the toolset is omitted entirely so the agent does not see `list_skills`/`load_skill` at all.

This mirrors the universe of decisions #17 (artifacts toolset), #28 (flow-state toolset), etc.: **scope-restricted toolsets injected at build time**, no runtime negotiation.

### Breaking change — no migrator, no auto-repair

Pre-decision-#29 stores carried `Instructions` / `References[]` /
`Name` / `Description` directly on each skill entry. The new shape
keeps only `{ID, Slug}`. We deliberately do NOT ship a migrator
that rewrites the legacy entries:

- **Auto-migration is the wrong default for skill content.** The
  operator wrote those instructions; rebuilding them from
  truncated/legacy fields gets us a SKILL.md the operator did not
  author. Better that they re-upload the canonical package
  themselves.
- **A persistent compat layer rots silently.** The first iteration
  did include one; it left half-written SKILL.md files with
  stacked frontmatters and resources mis-routed under
  `references/{references,assets,scripts}/...` because every
  edge case it tried to handle was a new edge case it broke. Once
  removed, every store stays clean by construction.

**Detection at startup**: `store.detectBrokenSkills` walks the raw
`store.json` bytes and flags every skill entry that carries fields
outside `{id, slug}`. Each flagged ID is logged once at WARN level
on load with a clear instruction:

```
WARN skill in legacy format and will be ignored — re-upload through the admin UI
   id=abc-123
   reason="legacy fields present: instructions, name, references"
   action="remove the entry from data/store.json and re-upload the skill via Skills → Upload Skill"
```

**Runtime behaviour**: every Skill accessor on `*Store`
(`ListSkills`, `GetSkill`, `GetSkillBySlug`, `ListRawSkills`,
`GetRawSkill`) silently filters broken IDs out. The admin UI
therefore doesn't see them, the agent build path doesn't link
them, and the upload handler can re-create the same slug as a
clean skill (the broken entry no longer claims it). The legacy
entry stays in `store.json` until the operator removes it by
hand — we never overwrite the file behind the operator's back.

**Operator playbook**:

1. Read the WARN log to find the affected skill IDs.
2. Edit `data/store.json` and delete each flagged entry from the
   `"skills": [...]` array.
3. Open the admin UI → Skills → Upload Skill, drop a SKILL.md or
   `.zip`/`.tar.gz` package built against the
   [Agent Skills specification](https://agentskills.io/specification).
4. Re-link the skill on every agent that previously used it
   (the link is by ID, so a fresh upload gets a fresh ID).

This playbook is the entire migration path. There is no fallback,
no automatic conversion, no "try harder" repair pass.

### Do not

- Reintroduce `Name`, `Description`, `Instructions` or `References` to the in-store `Skill` struct. The on-disk SKILL.md is the authoritative source; anything else duplicates state and invites drift.
- Inline skill content into the system prompt at agent build time (the old `--- Skill: ... ---` blocks). The skilltoolset is the contract — the LLM decides when to call `load_skill`.
- Add ad-hoc endpoints for individual file uploads or single-field edits. Upload-only via `/skills/upload` keeps the admin API and the operator's mental model aligned.
- Bypass the per-agent `AgentFS` whitelist. The whole agent → skill scoping invariant lives there; if a future feature needs to expand scope (e.g. expose every skill to a "super-agent"), it should pass an explicit allow-all list, not skip the wrapper.
- Use a custom `skill.Source` adapter just to bridge a different on-disk layout. ADK's `FileSystemSource` plus our `AgentFS` is the entire bridge — keep it that way.
- Reintroduce a migrator or auto-repair pass for legacy stores. Decided against during the first cut: a compat layer rotted silently and produced corrupted on-disk packages. The breaking-change path (detect + log + filter) is the contract, full stop.

**Files**:

- `server/store/types.go` — `Skill{ID, Slug}` (legacy fields removed).
- `server/store/store.go` — Skill CRUD, `SkillsDir()`, `SkillDir(slug)`, `detectBrokenSkills`, broken-skill filter on every accessor.
- `server/store/skills_broken_test.go` — detection + accessor-filter tests.
- `server/agent/tools/skills/agentfs.go` + tests — per-agent fs.FS whitelist wrapper.
- `server/agent/tools/skills/package.go` + tests — `ParsePackage`, `WritePackage`, `PackageAsTarGz`, slug helpers.
- `server/agent/tools/skills/tolerant.go` + tests — permissive `skill.Source` wrapper that survives non-canonical frontmatter keys (`version:`, `author:`, …).
- `server/agent/agent.go` — `buildSkillToolset`, `BuildAgentInstanceParams.SkillSlugs`/`SkillsDir`, removal of inline skill injection in `buildInstruction`.
- `server/agent/flow.go` — propagates `SkillSlugs`/`SkillsDir` through `FlowBuildDeps`.
- `server/api/admin/skills.go` — upload-only handlers, hydrated GET shape with permissive frontmatter parser.
- `server/api/admin/handler.go` — new route table.
- `frontend/admin-ui/src/views/skills/SkillDialog.vue` — upload-only modal with replace toggle.
- `frontend/admin-ui/src/views/skills/SkillViewDialog.vue` — read-only viewer (frontmatter, instructions, resources, download).
- `frontend/admin-ui/src/views/skills/SkillsList.vue` — opens viewer on click, upload from header button.
- `frontend/admin-ui/src/lib/api/skills.js` — upload/get/list/delete/download.
- `frontend/admin-ui/src/lib/markdown.js` + style.css `.magec-markdown` block — Magec-flavoured markdown renderer for skill instructions.


## 30. Flows are a directed graph on adk-go v2

Flows are a directed graph of typed nodes joined by edges, built on the
adk-go v2 workflow engine (`google.golang.org/adk/v2`). A `FlowDefinition`
holds `Entry`, `Nodes` and `Edges`; the builder emits a single
`workflowagent` and synthesizes the `Start -> Entry` edge. A node's ID is its
adk `Node.Name()` and the `event.Author`, so output filtering matches node IDs
directly with no synthetic naming scheme.

Node types: `agent` (runs an AgentDefinition), `router` (ordered CEL rules over
flow state emit a route label matched by outgoing `StringRoute` edges),
`join` (fan-in barrier), `parallel` (runs one agent once per list item via
`ParallelWorker`), `subflow` (embeds another flow as a `WorkflowNode` built
from its edges), `expression` (CEL value over `input`+`state`), `template`
(placeholder text), and `code` (Starlark). Loops are back edges gated by a
router; there is no loop container.

### Why a graph

A visual flow editor is a graph (boxes and arrows), so the data model, the
builder and the editor share one shape. Sequencing is an edge, fan-out is
edges plus a `join`, a loop is a back edge gated by a `router`; none of these
need a container node or an escalate/exit-loop mechanism. Flows persist only in
graph form (`Entry`/`Nodes`/`Edges`); there is no importer for other shapes.

### Code node capabilities and limits

The `code` node runs user Starlark via `github.com/1set/starlet`. The admin
who deploys governs capability, not the flow author: every starlet library
ships enabled, and the admin disables some in `Settings.Flows.DisabledLibraries`.
A fresh starlet Machine is built per execution from a loader list prebuilt in
`agent.New`. Execution limits (wall-clock timeout, output-size cap) have a
global ceiling in `Settings.Flows` and an optional per-node override, with
effective = min(node, ceiling) and 0 = unlimited; a runaway script is cut by a
Starlark step budget and the context deadline. This keeps the sandbox decision
where it belongs: the operator knows whether the deployment is an isolated
distroless container or an exposed binary.

### Do not

- Reintroduce `sequential`/`parallel`/`loop` container node types. Sequencing
  is edges, fan-out is edges plus a `join`, loops are a back edge plus a
  router; container nodes would reintroduce nesting.
- Hardcode a library allowlist for the code node. Capability is an admin
  runtime decision (Settings), not a compile-time policy.
- Drop the code-node execution limits as non-configurable. They guard Magec's
  own availability (a runaway loop or huge output), independent of the
  network/disk capability question, and the admin can disable them knowingly.
- Evaluate anything in an edge. Edges only match a label; the source router
  node decides routing by emitting `ev.Routes`.

**Files**:

- `server/store/types.go`: `FlowDefinition{Entry,Nodes,Edges,StartX,StartY}`, `FlowNode`, `FlowEdge`, `FlowRule`, node-type constants, `FlowsSettings` on `Settings`.
- `server/agent/flow.go`: `buildEdges`/`buildNode`, `workflowagent.New`, subflow via `WorkflowNode`, parallel via `ParallelWorker`.
- `server/agent/router_node.go`: CEL router with the `iterations` counter and `maxLoopIterations` guard.
- `server/agent/transform_nodes.go`: expression and template nodes.
- `server/agent/code_node.go`: Starlark code node, `effectiveLimit`, per-execution machine.
- `server/agent/flowexit/`: CEL compile/evaluate for router guards and expression values.
- `server/agent/flowgraph/validate.go` + tests: graph validation.
- `server/api/admin/flows.go`: `flowgraph.Validate` on save.
- `frontend/admin-ui/src/views/flows/`: `FlowCanvas.vue`, `FlowNode.vue` (graph editor, resizable nodes, full-screen, draggable Start).
- `frontend/admin-ui/src/views/settings/FlowsSection.vue`: library pills and execution limits.


## 31. Run auditing: raw events recorded by a runner plugin, projected on read

Every runner invocation (agent or flow, from any entry point) is recorded by
`server/agent/runrecorder`, an adk v2 plugin registered in
`runner.PluginConfig.Plugins` alongside contextguard. The recorder is a pure
observer: it never mutates events, never returns an error from a callback, and
swallows sink failures, so auditing can never break a run.

What is persisted per run: the ordered raw `session.Event` list (the adk
workflow scheduler yields events through a single queue, so per-run order is
total) plus run metadata (app, session, user, client, source, timestamps,
status, error). No distilled format is stored; the admin API derives the
per-node activation timeline at read time (`projectActivations` in
`server/api/admin/runs.go`). Views can therefore evolve or be fixed without
data migration, and `?raw=true` exposes the untouched events for replay.

Storage is SQLite at `data/runs.db` through `server/runs`, using
`modernc.org/sqlite` (pure Go). CGO surface stays at its current minimum
(onnxruntime only); this was an explicit constraint. SQLite is scoped to runs
only; `store.json` is untouched. Retention is SQL (`Sweep` by age and by
newest-N per app, swept hourly). Sessions live in Redis and are ephemeral, so
this database is the durable audit copy.

Two facts about the adk plugin API shaped the design (verified empirically):
run-fatal node errors are not events and never reach the plugin (they surface
only as the run iterator's error, consumed inside adkrest), and
`AfterRunCallback` carries no outcome. The recorder therefore exposes
`MarkRunError`, called by the `RunAudit` middleware, which also annotates
client attribution (`Annotate`) since the plugin cannot see HTTP. The
middleware scans the SSE response incrementally for `event: error` frames; it
does not buffer the stream.

`event.Author` for plain function/join workflow nodes is the workflow agent's
name, not the node name; Magec's own nodes set `Author` to the node ID. The
recorder persists both `Author` and `NodeInfo.Path`, and the projection groups
by consecutive runs of Author (falling back to NodePath) so loop reactivations
appear as separate activations.

OpenTelemetry was considered and excluded: adk v2 already emits `invoke_node`
spans natively for users with a collector, and persisting spans locally would
reimplement a tracing backend. The recorder is a product feature (embedded,
queryable, own retention), not infra telemetry.

Planned phase 2: rebuild the Conversations audit as a projection over recorded
runs and retire the stream-buffering conversation middleware and the persisted
dual perspective.
