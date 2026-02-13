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

### ~~Unificar `client/` y `clients/`~~ ✅

**Completado**: `server/client/` absorbido en `server/clients/`. Cada subtipo tiene `spec.go` (schema) junto a su runtime (`handler.go`, `bot.go`, `scheduler.go`). Registry y Provider interface en `clients/provider.go` + `clients/registry.go`.

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
- [x] Sidebar navigation (3 groups: Infraestructura, Agentes, Conexiones)
- [x] TopBar with section context + stats badges + refresh
- [x] Flow editor: canvas with nested blocks, drag-and-drop, pan/zoom
- [x] Clients can select flows in allowedAgents (UI shows agents + flows)
- [x] `responseAgent` flag on FlowStep: filter flow event stream by marked agents
- [x] Flow editor: broadcast icon toggle for responseAgent on agent nodes
- [x] Pre-seed session state with all agent outputKeys (prevents ADK template failures)
- [x] extractResponseText rewritten: filter by author, concat with `\n---\n`, backwards-compat
