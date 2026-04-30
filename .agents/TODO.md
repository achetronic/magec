# Magec - TODO

## Recently Completed

### This branch (`feature/todo-audit-cleanup`, 2026-04-18)

- **Telegram: thread-aware error messages** — the 4 remaining error-path `SendMessage` calls in `telegram/bot.go` now pass `MessageThreadID`, keeping error replies in the origin topic.
- **Slack multimodal (inlineData)** — new `extractInlineDataFromFiles` in `slack/bot.go` processes non-audio files (`image/*`, `application/pdf`, `text/*`, etc.) from DMs, encodes them as base64 `inlineData` parts. `callAgentSSE` and `processMessage` now accept a `[]map[string]interface{}` parts slice. 5MB/file limit.
- **Discord multimodal (inlineData)** — new `extractInlineDataFromAttachments` in `discord/bot.go` processes non-audio message attachments the same way. Voice messages still handled separately by `handleVoice`.
- **MemoryCard border** — removed the active-state override so the card follows the normal Card hover behaviour (grey at rest, green tint on hover). Active state is indicated solely by the radio button at the top-left.

### Earlier audit (2026-04-18)

- **Client Config — DefaultAgent + ThreadHistoryLimit** — `defaultAgent` persisted on `!agent <id>` with fallback chain, `threadHistoryLimit` replaces hardcoded 50. Shared schema helpers in `server/clients/provider.go`.
- **Large Message Handling (Telegram/Slack/Discord)** — `server/clients/msgutil/` package with `ValidateInputLength` + `SplitMessage`.
- **Artifact Management Toolset** — `server/agent/tools/artifacts/toolset.go`. Wired via `base_toolset.go`. Filesystem-backed. Clients auto-deliver new artifacts as file attachments.
- **Voice Provider Registry (Multi-Backend TTS/STT)** — `server/voice/` with OpenAI + Gemini providers. Typed `TTSConfig`/`STTConfig` structs per provider. `GET /voice/types` endpoint.
- **Skill Card View Formatter** — `lib/frontmatter.js`, canonical skills render structured cards.
- **Skill Package Upload (ZIP/tar.gz)** — `POST /skills/{id}/package`.
- **Secret Deletion + Env Var Cleanup** — `DeleteSecret` calls `os.Unsetenv`, `UpdateSecret` unsets old key on rename, both re-expand store data via `reExpandDataLocked`.
- **Telegram Multimodal Files (inlineData)** — `telegram/bot.go` downloads `Photo`/`Document`/`Video`/`Audio`/`Animation`/`VideoNote`/`Sticker`, encodes base64, sends as `inlineData` parts. 5MB limit enforced.
- **Voice UI Multiline Text Input** — `PanelHistory.vue` uses `<textarea>` with Shift+Enter for newline, Enter to send, auto-resize up to 150px.
- **Voice UI File Attachments** — `PanelHistory.vue` supports `image/*`, `application/pdf`, `text/*` via `<input type="file">`, base64 encoded, preview thumbnails, removable.
- **Tool Execution Visibility** — `msgutil/sse.go` has `FormatToolCall*` / `FormatToolResult*` for Telegram (expandable blockquote), Discord and Slack (compact summary). Voice UI has collapsible tool blocks in `ChatMessage.vue`. `!showtools` toggle per client.
- **Admin UI: Strip Metadata in ConversationDetail** — `stripMetadata` applied in `renderMarkdown` and `handleExportPDF`.
- **MemoryCard uses Card component** — `MemoryCard.vue` wraps `<Card color="green">` correctly.
- **Composable Flows (flow-as-step) — partial** — `server/agent/agent.go:sortFlowsTopologically` provides topological sort with cycle detection. Still pending: admin API validation, `InheritResponseAgents` field, UI selector showing flows as step targets.
- **Telegram Groups + Threads — in-thread replies** — @mention filter in groups/supergroups implemented. `MessageThreadID` threaded in all outbound calls (typing, tool counter, text response, error paths, artifact delivery).

---

## High Priority

### Multimodal File Support — AppMention in Slack (channel files)

**Status**: DMs done in this branch. Channel mentions pending.

The Slack `extractInlineDataFromFiles` helper only runs in the DM path (`handleMessage`) because `AppMentionEvent` does not expose a `Files` field in `slack-go v0.17.3`. When a user shares a file in a channel and mentions the bot, we don't yet extract it.

**Approach options**:
- Upgrade to a Slack events library version that surfaces files on `AppMentionEvent`, or fetch the parent message via `conversations.history` with the `ts` from the mention.
- Alternatively, subscribe to `file_shared` events and correlate with the mention.

**Files**: `server/clients/slack/bot.go`.

---

### Improve Drag-and-Drop UX in Visual Flow Editor

The visual flow editor's drag-and-drop already has placeholder, ghost, drop highlight, dead zones and vuedraggable reorder. Remaining polish items are subjective — define specific UX targets before working on this.

**Modify**: `frontend/admin-ui/src/views/flows/` (flow editor components)

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

### Composable Flows (flow-as-step) — UI + Admin Validation

**Status**: Topological sort + cycle detection at build time already implemented (`agent/agent.go:sortFlowsTopologically`). Still pending:

- Admin API validation (reject cycles at save time)
- `FlowStep.InheritResponseAgents *bool` (default true). When false, step's sub-flow responseAgents excluded from parent flow's `ResponseAgentIDs()`
- Admin UI: flow step selector must show both agents and flows (distinguished by `type`)

**Modify**: `server/store/types.go`, `server/api/admin/flows.go`, `frontend/admin-ui/src/views/flows/`.

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

### Evaluate Flow Subagent Invocation Model

Should clients target sub-agents within flows? Should flows support conditional routing? Should execution include per-step metadata?

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
