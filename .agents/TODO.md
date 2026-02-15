# Magec - TODO

## High Priority

### Infinite Conversation (ContextGuard Middleware)

**Problem**: LLM context windows are finite. Long conversations silently degrade or fail when the context fills up. There's no mechanism to detect this and continue seamlessly.

**Solution**: A new `ContextGuard` middleware in the HTTP chain that monitors context usage and automatically rotates sessions when nearing the limit.

**Architecture**:

```
Cliente → RecorderUser → FlowFilter → RecorderAdmin → ContextGuard → ADK
```

**How it works**:

1. Before each `/run`, ContextGuard reads the current ADK session (GET), estimates token count of the history (~4 chars ≈ 1 token), and compares against the model's `context_window`.
2. **Below 80%** → transparent pass-through.
3. **At 80%+** → intercepts:
   - Sends a summarization prompt to ADK on the current session
   - Creates a new ADK session with the summary as initial context
   - Re-sends the original user message to the new session
   - Returns the response with hidden HTML metadata:
     ```html
     <!--MAGEC_SESSION_CONTINUED:{"oldSessionId":"abc","newSessionId":"xyz","summary":"..."}:MAGEC_SESSION_CONTINUED-->
     Normal AI response here...
     ```
4. ConversationStore links conversations via `ParentID` (field already exists, unused).

**Context window lookup**:
- New package `server/contextwindow/` with `models.json` embedded via `//go:embed`
- Source: [Charm Crush provider.json](https://github.com/charmbracelet/crush/blob/main/internal/agent/hyper/provider.json)
- Lookup by `modelName` → `context_window`. Fallback: 128k if model not found.

**What does NOT change**:
- ADK API — untouched
- `/run` and `/run_sse` response format — clients that don't understand the HTML tag ignore it (it's an HTML comment)
- Existing clients (Telegram, cron, webhook) — keep working, session rotation is server-side and transparent
- Voice-UI and admin-UI can optionally parse the tag in the future to show "continued from previous session"

**New files**:
- `server/contextwindow/models.json` — embedded model context window data
- `server/contextwindow/contextwindow.go` — lookup + token estimation
- `server/middleware/contextguard.go` — the middleware

**Files to modify**:
- `server/main.go` — insert ContextGuard in the middleware chain
- `server/clients/executor.go` — pass model info for context window lookup

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

---

### Migrate voice-ui to Vue 3

**Problem**: The voice-ui is vanilla JS with ES modules loaded from CDN (no build step). The admin-ui already uses Vue 3 + Vite + Tailwind v4 + Pinia. Having two different stacks increases maintenance burden and makes it harder to share components, styles, and patterns.

**Goal**: Rewrite voice-ui using the same stack as admin-ui (Vue 3 Composition API, Vite, Tailwind v4). Migrate the class-based architecture (`MagecApp`, `AudioCapture`, `AudioRecorder`, etc.) to Vue components with composables.

**Key considerations**:
- PWA support must be preserved (manifest, service worker, installability)
- Audio pipeline (AudioWorklet, WebSocket, wake word, VAD) needs careful handling — these are low-level Web APIs that don't map directly to Vue reactivity
- i18n system (`data-i18n` attributes) should migrate to a Vue-native approach
- The Centella/Magec orb visualizer (canvas-based) can be a standalone component
- Session management, settings persistence, and error handling should use Pinia stores

---

## Medium Priority

### Refactor MemoryCard to use Card component

**Problem**: `MemoryCard.vue` has its own `<div class="bg-piedra-900 border rounded-xl p-4 ...">` with duplicated hover color styles (`hover:border-green-500/15 hover:shadow-[...]`). These values must be kept in sync manually with `Card.vue`'s `colorMap`.

**Solution**: Make `MemoryCard` wrap `<Card color="green">` as its outer container and put all its custom content (radio button, health indicator, action buttons) inside the slot. This way hover styles are controlled in one place.

**Files to modify**:
- `admin-ui/src/views/memory/MemoryCard.vue` — replace outer `<div>` with `<Card color="green">`

---

### Add Voice Activity Detection During TTS

**Problem**: On mobile, microphone can pick up speaker output and trigger wake word detection while TTS is playing.

**Possible solutions**:
- Mute microphone during TTS playback
- Implement echo cancellation
- Increase wake word threshold temporarily during playback

---

### Admin API Authentication

**Problem**: The admin API on port 8081 has no authentication. Anyone with network access can modify agents, backends, and clients.

**Possible solutions**:
- Basic auth (simple, config-based)
- API key in header
- Session-based login

---

### Move `response_format` Out of Clients Into Server-Side Config

**Problem**: The TTS `response_format` (e.g. `opus`, `mp3`, `wav`) is currently hardcoded by each client (Telegram sends `"opus"`). The speech proxy passes it through to the backend TTS service without any server-side control.

**Options to evaluate**:
1. Add `response_format` to `TTSRef` in the store (per-agent) — simple, but all clients of an agent get the same format
2. Add `response_format` to each client type's config — per-client control, proxy resolves it
3. Keep it client-side but document it as the expected contract

**Decision**: TBD

---


### Human-in-the-Loop Tool Confirmation

**Problem**: There's no way for an agent to pause and ask a human for approval before executing a sensitive action (e.g., deleting data, sending money, modifying config).

**ADK support**: v0.4.0 provides `toolconfirmation` — any tool can call `ctx.RequestConfirmation(hint, payload)` to pause execution and emit an `adk_request_confirmation` event.

**Current blocker**: All clients call `/api/v1/agent/run` synchronously. If a tool requests confirmation mid-execution, it's a deadlock.

**Required architecture changes**:

1. **Switch clients to SSE streaming** — Use `/run/sse` instead of `/run`
2. **Admin UI notification area** — Listens for confirmation events, shows Approve/Reject buttons
3. **Telegram support** — Inline keyboard with approve/reject buttons
4. **Pending confirmations store** (alternative to full SSE)

**Implementation order**:
1. Admin UI notification area + SSE connection
2. Pending confirmations API
3. Telegram inline keyboard integration

See `.agents/ADK_TOOLS.md` for full details on `toolconfirmation`.

---

### ~~Mover APIs a `api/admin/` y `api/user/`~~ ✅

**Completado**: `server/admin/` → `server/api/admin/`, `server/userapi/` → `server/api/user/` (package renombrado de `userapi` a `user`). Makefile swagger targets actualizados.

---

### Evaluate Flow Subagent Invocation Model

**Context**: Flows are registered as ADK agents and invoked via the same `/api/v1/agent/run` endpoint as regular agents. This means a flow *is* an agent from the caller's perspective — it just orchestrates sub-agents internally (sequential, parallel, loop).

**Current behavior**: Flow root step registered as `flow.ID` (UUID). Sub-steps named `{flowID}_{depth}`. Agent leaf nodes reuse the pre-built ADK agent directly (by store ID lookup).

**Questions to evaluate**:
1. Should clients be able to target individual sub-agents within a flow directly? Currently they can only invoke the flow root.
2. Should flows support conditional routing (e.g., route to different agents based on input classification)? ADK doesn't have a native "router" workflow agent — would need a custom agent.
3. Should flow execution results include per-step metadata (which agent ran, latency, token usage)? Currently only the final output is returned.
4. Should flows be composable — i.e., a flow step can reference another flow, not just an agent? Currently `FlowStepAgent` only looks up agents in `agentMap`.

**No action required now** — this is a design evaluation item for when more complex multi-agent workflows are needed.

---

### Evaluate Subagent-as-Tool Pattern

**Context**: ADK supports registering agents as tools that another agent can invoke on-demand during its reasoning loop (rather than through a fixed sequential/parallel flow). This enables an orchestrator agent to decide *at runtime* which specialists to call, how many times, and in what order — based on the user's input.

**Why it's interesting**:
- More flexible than static flows: the orchestrator agent reasons about *when* to call each subagent, rather than following a hardcoded graph.
- Natural fit for open-ended questions that might need 1, 2, or all specialists depending on complexity.
- Subagents become tools alongside MCP tools — the LLM picks what to call.

**Open UX/GUI questions**:
1. How to represent this in the admin UI? A flow has a clear visual tree. An agent-with-subagent-tools is conceptually different — the orchestrator *chooses* which tools to call. The flow editor doesn't map to this model.
2. Should this be a new agent config section ("Sub-agents" alongside MCP Servers)? Or a special flow step type?
3. How to surface tool-call decisions to the user? The orchestrator might call a subagent 0 or N times — harder to predict/debug than a fixed flow.
4. Does `responseAgent` still make sense here? The orchestrator itself is the response agent by definition, since it decides what to return.

**No action required now** — design evaluation for when the current sequential/parallel flow model feels too rigid.

---

## Low Priority

### Unify `models/` and `pretrained/` directories

**Problem**: ONNX models are split across two top-level directories — `models/` (wake word models + `wakewords.yaml`) and `pretrained/` (mel-spectrogram, VAD, embeddings). The distinction is unclear and adds confusion.

**Goal**: Merge both into a single `models/` directory with subdirectories by purpose (e.g. `models/wakeword/`, `models/vad/`, `models/embeddings/`).

**Files to update**:
- `server/voice/detector.go` — wake word model paths
- `server/voice/vad.go` — VAD model path
- `models/wakewords.yaml` — model path references
- `scripts/download-model.go` — download targets
- `docker/build/Dockerfile` — COPY paths
- `config.example.yaml` — if any model paths are referenced

---

### ~~Unificar `client/` y `clients/`~~ ✅

**Completado**: `server/client/` absorbido en `server/clients/`. Cada subtipo tiene `spec.go` (schema) junto a su runtime (`handler.go`, `bot.go`, `scheduler.go`). Registry y Provider interface en `clients/provider.go` + `clients/registry.go`.

---

---

### Gestión de credenciales desde la Admin UI

**Problema**: Las credenciales (API keys de backends, bot tokens de Telegram, etc.) están en claro en `data/store.json`. Ahora se soporta `${VAR}` via `os.ExpandEnv()` al cargar el store, pero no hay forma de gestionar esas variables desde la UI.

**Objetivo**: Diseñar un mecanismo para que la Admin UI permita definir/editar credenciales sin exponerlas en el JSON. Opciones a evaluar:
- `.env` file que el servidor cargue al arrancar + UI para editarlo
- Secrets store separado (fichero cifrado, vault externo)
- Campos `x-format: "password"` en los schemas ya marcan qué campos son sensibles — aprovechar eso para enmascarar/persistir de forma diferente

---

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

### Credential Management for Connection Strings

**Problem**: Connection strings contain credentials in plain text (`redis://:password@...`, `postgres://user:pass@...`). Currently stored directly in `data/store.json` and visible in the admin UI.

**Possible approaches**:
- Environment variable expansion in connection strings (`redis://:${REDIS_PASS}@localhost:6379/0`)
- Separate secrets store (encrypted at rest)
- Reference external secret managers (Vault, K8s secrets)
- At minimum: mask passwords in API responses, only show `****` in UI

### Speaker Identification

**Goal**: Identify who is speaking to personalize responses or restrict commands.

**Recommended solution**: [WeSpeaker](https://github.com/wenet-e2e/wespeaker)

### Telegram File/Artifact Support

**Goal**: Allow users to send files via Telegram for the AI to process.

**Depends on**: ADK artifacts implementation in `adk-utils-go`

### Database Persistence for Store

**Problem**: `data/store.json` is a single JSON file. Works fine for small setups but won't scale for multi-user or HA deployments.

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
- [x] Store-based agent creation: `agent.New()` accepts store types directly
- [x] Config split: YAML for infra only, all resources via admin API/store
- [x] Multi-agent support (server): `NewMultiLoader` routes by `appName`
- [x] Multi-agent support (voice-ui): `setAgent(agentId)` propagated to all components
- [x] Hot-reload agents on store changes: `OnChange()` channel + 500ms debounce
- [x] Voice endpoint redesign: `/api/v1/voice/{agentId}/speech` and `/transcription`
- [x] Rename with cascade: All resource types support renaming via PUT
- [x] Admin UI framework migration (Vue 3 + Vite + Tailwind v4 + Pinia)
- [x] Admin UI polish: Toast, Skeletons, Transitions, Empty States, Search Palette, Responsive Sidebar, Keyboard Shortcuts
- [x] Commands entity: reusable prompts with name, description, prompt
- [x] Triggers→Clients consolidation: cron and webhook are now client types
- [x] Client type registry: JSON Schema replaces FieldSpec (with oneOf, x-entity, x-format, x-placeholder)
- [x] device→direct client type rename
- [x] Migration chain: CronJobs→Triggers→Clients (automatic on store load)
- [x] Cron scheduler + Webhook HTTP handler (server/trigger/ package)
- [x] Cron shorthand support (@daily, @hourly, @weekly, @monthly, @yearly, @annually, @midnight)
- [x] Flow appName uses flow UUID instead of flow name (consistent with agent addressing)
- [x] Regenerated admin swagger docs (removed stale FieldSpec references)
- [x] Memory providers migrated to JSON Schema (ConfigSchema replaces ConfigFields/FieldSpec)
- [x] Webhook endpoint: POST /api/v1/webhooks/{id} with Bearer token auth
- [x] Webhook Swagger docs (userapi)
- [x] OutputKey on AgentDefinition (ADK output key for flows)
- [x] Entity color system (7 entities, documented in ENTITY_COLORS.md)
- [x] Card.vue `color` prop: hover border tint + glow shadow per entity color (DRY, all 7 list views use it)
- [x] Card hover intensity reduction: border `/15` opacity, shadow `0.04` — subtle, not distracting
- [x] Sidebar navigation (3 groups: Infraestructura, Agentes, Conexiones)
- [x] TopBar with section context + stats badges + refresh
- [x] Flow editor: canvas with nested blocks, drag-and-drop, pan/zoom
- [x] Clients can select flows in allowedAgents (UI shows agents + flows)
- [x] `responseAgent` flag on FlowStep: filter flow event stream by marked agents
- [x] Flow editor: broadcast icon toggle for responseAgent on agent nodes
- [x] Pre-seed session state with all agent outputKeys (prevents ADK template failures)
- [x] extractResponseText rewritten: filter by author, concat with `\n---\n`, backwards-compat
