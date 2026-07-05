# Magec - TODO

## Recently Completed

### This branch (`feat/workflow-graph-redesign`)

- **Flows as a directed graph on adk-go v2** — 8 node types (agent, router,
  join, parallel/Foreach, subflow, expression, template, code), CEL guards
  with `input` + `state` + `iterations`, fixed `otherwise` fallback route
  (decision #34), Starlark code node with admin-governed
  libraries and limits, graph validation, visual editor (fan-out, renamable
  node ID chips, NodeHelp popovers, StarlarkEditor, double-Escape fullscreen,
  blurred preview when minimized).
  Decisions #30, #28. Reference: `WORKFLOW_DESIGN.md`.
- **Run auditing end to end** — `runrecorder` plugin (raw ordered events +
  user input via OnUserMessageCallback), `server/runs` SQLite store
  (modernc.org/sqlite, retention sweeps, manual delete/clear), `RunAudit`
  middleware (attribution + SSE error capture), admin API with on-read
  activation projection (node-type snapshots per run, internal `__` nodes
  hidden), Runs UI (list with kind toggle + timeline detail with
  RunHeader/ActivationCard/StatusPill). Verified live on a 14-node torture
  flow. Decisions #31, #33.
- **Client metadata prefilter** — synthetic `__meta__` node strips the
  MAGEC_META block into `state.magec_meta`; `__` ID prefix reserved.
  Decision #32.
- **adk-utils-go v0.22.0** — Redis session service parity fixes (StateDelta
  applied to the live session, temp: key semantics, timestamps, List/Get
  validation); consumed from the remote module, vendor dropped.
- **Website flows docs rewritten for the graph model** — single concept-first
  page (Flow Control folded in), hardened through adversarial audit rounds,
  screenshots captured live from the admin UI.

---

## High Priority

### Run recorder: interrupted status on shutdown

A server restart mid-run produces a run record with status `completed` even
though the run was decapitated (observed live: a flow whose agents had called
tools but never emitted text was flushed as completed during the redeploy).
The AfterRunCallback defer fires during shutdown and flushes with the default
status. Fix candidates: flush as `interrupted` from Recorder.Close() for every
live accumulator, and/or only mark `completed` when a terminal run event was
observed. Tests: simulate Close() with a live accumulator and assert the
persisted status.

**Modify**: `server/agent/runrecorder/recorder.go` (+ test).

---

### Runs UI: event presentation polish

Remaining cosmetic issues observed on real runs:

- **Raw event summaries** — collapsed raw events show generic JSON; summarise
  the interesting shape instead (`functionCall: search_memory`,
  `functionResponse`, text preview).
- **MAGEC_META leaks into previews of pre-prefilter runs** — runs recorded
  before the `__meta__` prefilter existed still show the raw metadata block
  in input/output previews. Presentation-level filtering.

**Modify**: `server/api/admin/runs.go` (projection), `frontend/admin-ui/src/views/runs/`.

---

### Runs timeline: branch-aware projection

With fan-out, events of concurrent branches interleave in the single event
queue, which fragments the consecutive-author grouping (author A, B, A yields
three activations where two nodes ran) and makes branches hard to follow.
Agreed first steps, deliberately short of a git-graph redesign:

- Group activations by (author, branch) instead of author alone.
- Within a fan-out segment (between the forking node and its join), order
  activations branch by branch while keeping global chronology across
  segments.

**Modify**: `server/api/admin/runs.go` (projection + tests).

---

### MAGEC_META phase 2: metadata as state, context for agents

Phase 1 (the `__meta__` prefilter, decision #32) ships on this branch. Phase 2,
for its own branch:

- Client bots emit the metadata as a StateDelta instead of an inline comment
  block appended to the user text.
- Agent instructions gain a context block built from the metadata.
- Group chats prefix messages with the human author's name (`[Alby]:`) so
  multi-user conversations are attributable.

**Modify**: `server/clients/{telegram,discord,slack}/bot.go`, `server/clients/executor.go`, instruction builder in `server/agent/agent.go`.

---

### Run auditing phase 2: Conversations as a projection over runs

Rebuild the Conversations audit as a projection over recorded runs
(conversation = runs of one session; user perspective = on-read filter by
ResponseAgentNames). Retires the body-buffering ConversationRecorder
middleware and the persisted dual perspective. Clean break, no migrator.
Decision #31.

---

---

### Multimodal File Support — AppMention in Slack (channel files)

**Status**: DMs done in this branch. Channel mentions pending.

The Slack `extractInlineDataFromFiles` helper only runs in the DM path (`handleMessage`) because `AppMentionEvent` does not expose a `Files` field in `slack-go v0.17.3`. When a user shares a file in a channel and mentions the bot, we don't yet extract it.

**Approach options**:
- Upgrade to a Slack events library version that surfaces files on `AppMentionEvent`, or fetch the parent message via `conversations.history` with the `ts` from the mention.
- Alternatively, subscribe to `file_shared` events and correlate with the mention.

**Files**: `server/clients/slack/bot.go`.

---

### Workflow graph: ToolNode (deferred)

The flow redesign (`feat/workflow-graph-redesign`) added agent, router, join,
parallel and subflow nodes. adk-go v2 also offers `workflow.NewToolNode`, which
runs an MCP tool **directly as a step** (deterministic, no LLM turn). It is
deferred because Magec has **no static tool catalogue**: tools are discovered at
runtime when an MCP server connects, the store only persists the server
(endpoint/command), and `NewToolNode` requires a runnable `tool.Tool` with
resolved input AND output schemas. So ToolNode needs a prerequisite feature:

1. **MCP tool discovery** — an admin endpoint that connects to an MCP server
   and lists its tools with their JSON schemas (name, description, input/output).
2. **Editor tool picker** — a `tool` node whose reference is `{mcpServerId, toolName}`,
   resolved + validated against discovery (not free-typed).
3. **Builder** — resolve the chosen tool from the agent/flow's `mcptoolset` and
   wrap it with `workflow.NewToolNode`.

Until discovery exists, a tool node would be a free-typed `{server, name}` with
no validation and bad UX — explicitly not worth doing half-way.

**Modify**: `server/store/types.go`, `server/agent/flow.go`, `server/api/admin/mcps.go` (+ new discovery handler), `frontend/admin-ui/src/views/flows/`.

---

### Workflow graph: remaining debt

- **HITL pause/resume** — adk v2's `NewRequestInputEvent` + `NodeConfig.RerunOnResume`
  let a node pause for human input mid-flow (distinct from per-tool confirmation
  below). Phase 2 of the redesign.
- **String extensions in router guards** — the router CEL env lacks
  `ext.Strings`/`ext.Lists` (expression nodes have them); add if a guard ever
  needs `split`/`replace`.

---

### Human-in-the-Loop Tool Confirmation

**Problem**: MCP tools can perform sensitive actions (delete data, send emails, execute code). No way to require human approval before execution.

**Solution**: Use ADK v0.5.0's native `RequireConfirmationProvider` on `MCPToolset.Config`. This is a dynamic per-tool callback.

**Design decisions**:

- **Confirmation list lives on the agent, not on the MCP server**.
- **Agent config**: new field `toolConfirmation: ["delete_record", "send_email", "execute_*"]`.
- **Provider in `buildToolsets()`**: when creating each `MCPToolset`, pass a `RequireConfirmationProvider` that checks the agent's `toolConfirmation` list.
- **Admin UI — per-MCP tool selection**: the agent form shows tools from each connected MCP server (fetched via `client.ListTools()`).

**"Always Allow" (client-side, not in ADK)**: implement a shared `alwaysAllow map[string]bool` behind the provider, session-scoped.

**Chat UI confirmation dialog** (all clients): render `adk_request_confirmation` events with Approve/Reject/Always Allow buttons.

**Client changes**: Telegram (inline keyboard), Slack (block actions), Discord (buttons), Voice UI (confirmation card), Executor (auto-approve or skip).

See `.agents/ADK_TOOLS.md` for protocol details.

**Modify**: `server/agent/agent.go`, `server/store/types.go`, `server/clients/{telegram,slack,discord}/bot.go`, `server/clients/executor.go`, `frontend/voice-ui/`, `frontend/admin-ui/`.

---

### TTS Real-Time Streaming Playback

**Problem**: Current TTS waits for all audio chunks before playback (`response.blob()` in `OpenAITTS.js:67`). Noticeable delay.

**Solution**: Incremental playback using Web Audio API — decode and schedule each chunk as it arrives.

**Modify**: `frontend/voice-ui/src/lib/audio/OpenAITTS.js`.

---

## Medium Priority

### Use `artifact.Service.GetArtifactVersion` for richer metadata

**Status**: Available since ADK v1.1.0, not used.

Today both `list_artifacts` and the ephemeral URL endpoint go through `Load(version)`, which decodes the full blob just to read metadata (mimetype, size, etc.). ADK now exposes `GetArtifactVersion` which returns metadata only.

**Wins**:

- `list_artifacts` could return `{name, version, mime, size, createdAt}` per item without touching bytes.
- The ephemeral URL handler (`server/main.go:newEphemeralArtifactHandler`) could decide content-type and content-disposition without decoding the whole file twice.

**Files**: `server/agent/tools/artifacts/toolset.go`, `server/main.go`.

---

### Evaluate `adka2a.RunnerProvider` for A2A handler

**Status**: Available since ADK v1.2.0, not used.

ADK v1.2.0 added `RunnerProvider` to `adka2a.ExecutorConfig` so callers can supply their own runner factory instead of letting `adka2a` build one from `RunnerConfig`. Could let us share runners across A2A invocations or hook into custom plugin chains without the executor wrapping things its own way.

Worth evaluating whether it materially improves the rebuild path in `server/a2a/handler.go` or it's a marginal win we can skip. Spike, decide, write a short note in DECISIONS.md either way.

**Files**: `server/a2a/handler.go`.

---

### A2A FilePart → inline/artifact policy

**Problem**: A2A messages carry three `Part` types: `TextPart`, `FilePart` (binary + MIME), and `DataPart` (structured JSON). The A2A handler in `server/a2a/handler.go` currently only extracts the `TextPart` and ignores files. Agent cards advertise `DefaultInputModes: ["text/plain"]`.

**Solution**: Reuse the existing `msgutil` helpers (decision #24). When an incoming A2A message contains `FilePart`s:

- Files are saved via `msgutil.StoreAsArtifact(...)` using the router's `ArtifactService()`, descriptor lines wrapped with `msgutil.AttachedArtifactsBlock(...)`.
- Declare additional MIME types in the agent card's `DefaultInputModes` (`image/*`, `application/pdf`, `text/*`, `audio/*`).

**Files**: `server/a2a/handler.go`, agent-card builder inside the same file, wire `ArtifactService()` from `agentRouterHandler` into the handler on rebuild.

---

### Telegram Channel Posts (`update.ChannelPost`)

Messages posted in Telegram channels (not groups) are silently ignored. Supporting them requires deciding on:

- Authorization model — `msg.From` is optional on channel posts, so `isAllowed(userID, chatID)` needs a channel-aware path (likely based on `chatID` only).
- Trigger — always process, or only with a keyword/command/@mention.

**Modify**: `server/clients/telegram/bot.go` — register `handler.HandleChannelPost(...)`, decide on auth + trigger.

---

### Filter Tool Messages from `fetchThreadContext`

**Background**: `fetchThreadContext` is necessary — ADK only sees messages routed through the bot.

**Problem**: When `!showtools` is active, tool call/result messages posted by the bot into the thread are picked up by `fetchThreadContext` and injected as noise. The LLM already has that tool activity in its ADK session.

**Solution**: In `fetchThreadContext`, skip messages from the bot whose text starts with the tool-output prefixes (`🔧`, `✅`, `⚡`, `⚙️`). Check `msg.Author.ID == botID` (Discord) or `msg.BotID != ""` (Slack).

**Files**: `server/clients/slack/bot.go`, `server/clients/discord/bot.go`.

---

### Voice Activity Detection During TTS

On mobile, microphone picks up speaker output and triggers wake word during TTS playback. Options: mute mic during TTS, echo cancellation, or increase threshold temporarily.

**Files**: `frontend/voice-ui/src/lib/audio/AudioCapture.js`, `frontend/voice-ui/src/lib/audio/OpenAITTS.js`.

---

### Move `response_format` Out of Clients

TTS `response_format` (hardcoded `"opus"`) in Telegram/Slack/Discord clients (`bot.go` TTS calls). Could be per-agent in `TTSRef`, per-client in config, or documented as client contract. **Decision**: TBD. Related to the voice provider registry.

---

### Remote A2A Agents as Tools (orchestration mode)

**Problem**: A user may have multiple A2A agents deployed across their network. They want a local orchestrator agent that can call remote agents when it decides.

**Solution**: Use ADK's `agent/remoteagent` + `tool/agenttool` to wrap each remote A2A agent as a tool.

**Implementation**:

1. New entity `RemoteAgent` in the store: `{id, name, agentCardURL, credentials}`
2. In `buildToolsets()`: for each remote agent configured on the agent, create `remoteagent.NewA2A()` + `agenttool.New()`
3. Agent config: new field `remoteAgents []string`
4. Admin UI: section for managing remote agents + agent form multi-select

**Modify**: `server/agent/agent.go`, `server/store/types.go`, `server/api/admin/`, `frontend/admin-ui/`.

---

### Remote A2A Agents as Sub-agents (transfer mode)

**Problem**: In some cases, a remote A2A agent needs to interact directly with the user without the orchestrator in the middle.

**Solution**: Pass remote as `SubAgents` in `llmagent.Config`. Orchestrator's LLM can "transfer" conversation to the remote.

**Characteristics**: Remote gets full context. No orchestrator tokens during transfer. One transfer at a time.

**Implementation**: Same `RemoteAgent` entity as tool mode. Per-remote flag for tool-vs-subagent mode.

---

### Evaluate Subagent-as-Tool Pattern

ADK supports agents as tools — orchestrator decides at runtime which specialists to call.

---

### ContextGuard Summary Tier Migration (app/user scope)

**Problem**: ContextGuard summaries are session-scoped. When a user switches client or starts a new thread, the agent loses all prior conversation context.

**Blocked by**: All users are currently `default_user` (hardcoded in `telegram/bot.go:491`, `discord/bot.go:255`, `slack/bot.go`). Moving to `app:` tier would share summaries across all clients — wrong.

**Solution (requires real user identity)**:

1. Per-client user identity: `discord_123456`, `slack_U0ABC`, `telegram_98765`
2. Move ContextGuard state keys to `user:` tier (`session.KeyPrefixUser` prefix)

**Modify**: `adk-utils-go/plugin/contextguard/contextguard.go`, all client bots, `server/clients/executor.go`.

---

## Low Priority

### Remove `migrateTTSConfig` Store Migration

Introduced in v0.18.0. By v0.20.0 all installations will have migrated. Remove from `server/store/store.go`.

---

### More TTS Voices Configuration UI

Voice selection is server-side only. Could add UI for preview and selection.

### Offline Mode

Cache TTS, service worker, local transcription model.

### Multi-Language Wake Words

Different models per language, auto-switch based on i18n selection.
