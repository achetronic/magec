# Magec - TODO

## Recently Completed

### This branch (`feature/admin-mcp-server`, 2026-05-21)

- **Embedded admin MCP server** — new opt-in HTTP server on port 8082 (`server.mcp.enabled: true`) exposes the full admin API as ~54 MCP tools (`magec_list_backends`, `magec_create_agent`, etc.) over Streamable HTTP. Auth reuses `server.adminPassword` as a bearer token via the new `middleware.BearerAuth`. Tools call `*store.Store` directly — no HTTP roundtrip to the admin port. Skills upload/download and backup/restore stay on admin REST (binary streams don't map cleanly to tools). Decision #30. Files: `server/api/mcp/`, `server/main.go`, `server/config/config.go`, `server/middleware/middleware.go`, `website/content/docs/admin-mcp-server.md`, `website/hugo.toml`.

### Earlier branch (`feature/lazy-load-skills`, 2026-05-09)

- **Skills as on-disk packages** — skills now live at `data/skills/{slug}/SKILL.md` with optional `references/`, `assets/`, `scripts/` sub-trees. Store keeps only `{id, slug}`; everything else (frontmatter, instructions, resources) is read live from disk. Inline injection into system prompts is gone. Decision #29.
- **Per-agent `skilltoolset`** — adopted `google.golang.org/adk/tool/skilltoolset` (v1.2.0). Each agent gets its own `skilltoolset` rooted at `data/skills/` but filtered through `agent/tools/skills.AgentFS`, an `fs.FS` wrapper that whitelists only the slugs linked to the agent. Agents with no linked skills get no toolset at all.
- **Single upload endpoint (`POST /skills/upload`)** — accepts a SKILL.md or a `.zip`/`.tar.gz` package; `?replace=true` overwrites an existing skill that owns the same slug while preserving its store ID so agent links stay valid. Replaces the old `/skills`, `/skills/{id}`, `/skills/{id}/references`, `/skills/{id}/package` admin endpoints.
- **Read-only Skill viewer (admin UI)** — `SkillViewDialog.vue` opens on card click, renders frontmatter (license, compatibility, allowed-tools, custom metadata), instructions as Markdown, and lists every resource grouped by `references/assets/scripts`. Includes a Download button that streams the on-disk package as `tar.gz` for backups or magec-to-magec copies.
- **Upload-only Skill modal** — `SkillDialog.vue` no longer has Manual/Package toggle, Name/Description/Instructions fields, or a per-file uploader. One dropzone, one button.
- **Legacy stores → breaking change, no migrator** — pre-decision-#29 stores are flagged at load time by `store.detectBrokenSkills`. Each broken skill (entry with fields outside `{id, slug}`) gets one WARN-level log on startup with id/reason/action and is filtered out of every `*Store` accessor (`ListSkills`, `GetSkill`, `GetSkillBySlug`, `ListRawSkills`, `GetRawSkill`). The admin UI doesn't see them, the agent build path doesn't link them, and the upload handler can re-create the same slug as a clean skill. The legacy entry stays in `store.json` until the operator removes it by hand and re-uploads. No auto-migration, no on-disk repair: an earlier compat layer rotted silently and produced corrupted SKILL.md files.
- **Tolerant frontmatter parsing** — admin GET (`hydrateSkill`) uses a permissive YAML decode so non-canonical frontmatter keys (`version:`, `author:`, `tags:`) don't blank the viewer; runtime uses `agent/tools/skills.TolerantSource` so the agent toolset stays alive when ADK's strict `KnownFields(true)` parser would have rejected the SKILL.md outright.
- **Magec-flavoured markdown renderer** — `frontend/admin-ui/src/lib/markdown.js` + `.magec-markdown` block in `style.css`. Custom `marked` renderer + tiny regex syntax highlighter for YAML/JSON/bash/Go/JS/TS/Python/Markdown. No external highlight library; visual choices documented inline.

### Earlier branch (`feature/flow-state`, 2026-05-04)

- **Flow-shared state** — agents inside any flow get `set_state(key, value)` and `get_state(key)` tools backed by `session.state` under the `flow:` prefix. Standalone agents do not get them. Implemented in `server/agent/tools/flowstate/`. Decision #28.
- **LLM-driven loop exit** — loop steps gained an `ExitLoop` flag. When set, every agent in the loop's subtree (any depth, propagated through nested sequentials/parallels) receives ADK's native `exit_loop` tool. The current iteration completes before the loop terminates (matches ADK's loopagent semantics). Decision #28.
- **Expression-driven loop exit (`ExitWhen`)** — loop steps gained a CEL expression evaluated against the shared flow state after every iteration. New `server/agent/flowexit/` package with `Compile` (called both by admin validation and at flow build time) and `NewExitWhenAgent` (synthetic evaluator agent appended as the last child of the loopagent, emits `Escalate` when the expression is true). New direct dep: `github.com/google/cel-go`. Decision #28.
- **`BuildAgentInstance` refactor** — `buildSingleAgent` was renamed and exported, signature converted to a `BuildAgentInstanceParams` struct. `flow.go` now builds a fresh ADK instance per agent appearance via `BuildAgentInstance`, allowing scope-dependent toolsets/instructions without polluting the standalone catalogue. `wrapAgent` is kept only for flow-as-step composition.
- **Loop config dialog (admin UI)** — replaced the `prompt()`-based max iterations input with `LoopConfigDialog.vue` exposing the three exit strategies (max only / agent decides / expression). `FlowBlock` badge now shows the chosen strategy alongside the iteration count.

### Earlier branch (`feature/always-artifacts`, 2026-04-30)

- **Always-artifact attachments** — removed the inline-vs-artifact size threshold; `msgutil.ShouldInline` and `msgutil.InlinePart` are gone. Every user upload (Telegram/Slack/Discord) flows through `msgutil.StoreAsArtifact` + `AttachedArtifactsBlock`. Decision #24 rewritten.
- **`load_artifact` via RequestProcessor** — replaced the old `functiontool` that returned base64 in JSON (which corrupted the model's view and burned context) with a manual `tool.Tool` implementing `RequestProcessor`. It mutates the existing `FunctionResponse` Content rather than appending a new one, preserving the "1 session event = 1 req.Contents entry" invariant ContextGuard relies on. Decision #25.
- **`adk-utils-go` v0.15.2** — bumped from v0.13.0 to pick up `lowercaseTypes` in the Anthropic provider, which normalises `Type: "STRING"` (Gemini-style) to `"string"` (JSON Schema draft 2020-12) for tool input schemas built with `genai.Schema`.
- **`export_artifact` tool + Settings.TemporaryDir** — new tool decodes an artifact and writes the raw bytes to disk under a single, centrally-resolved directory (`Store.ResolveTemporaryDir()` → `Settings.TemporaryDir` or `os.TempDir()`). Lets the model hand artifacts off to any filesystem-aware tool. Admin UI gained a Runtime section under Settings to configure the path. Decision #26.
- **`get_artifact_url` tool + ephemeral URL endpoint** — new `GET /api/v1/ephemeral/artifacts/{token}` serves an artifact's raw bytes when called with a valid HMAC-SHA256 signed token. Companion tool `get_artifact_url` mints those URLs (1 h TTL, signed with `server.encryptionKey`, public hostname from `getA2APublicURL`). Lets sidecars and remote consumers fetch artifacts over HTTP without sharing volumes. New `server/ephemeral` package with sign/verify primitives. Decision #27.

### Older branch (`feature/todo-audit-cleanup`, 2026-04-18)

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
