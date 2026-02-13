# ADK Go Built-in Tools Reference

Reference for all tools shipped with `google.golang.org/adk` (evaluated at v0.4.0, with `main`-only packages noted).

## Tools Overview

| Package | Tool Name | Type | Used in Magec | Description |
|---------|-----------|------|---------------|-------------|
| `exitlooptool` | `exit_loop` | Tool | Yes (loop escalate) | Sets `Escalate=true` to break out of a `loopagent` |
| `functiontool` | *(factory)* | Constructor | Yes (indirectly) | Creates tools from Go functions. Used by `adk-utils-go` for memory tools |
| `mcptoolset` | *(toolset)* | Toolset | Yes | Wraps MCP servers as ADK toolsets |
| `agenttool` | `{agent.Name()}` | Tool | Not yet | Calls another agent as a tool (creates runner + session internally) |
| `geminitool` | `google_search` + factory | Tool | No | Gemini-native tools (Google Search, Retrieval). Only works with Gemini models |
| `loadartifactstool` | `load_artifacts` | Tool | Not yet | Lets agents load session artifacts (files, images) on demand |
| `toolconfirmation` | *(protocol)* | Protocol | Not yet | Human-in-the-Loop confirmation via `ctx.RequestConfirmation()` |
| `loadmemorytool` | `load_memory` | Tool | Not yet (custom equivalent exists) | LLM-driven memory search. Official ADK version of `search_memory` |
| `preloadmemorytool` | `preload_memory` | Tool (auto) | Not yet (custom equivalent exists) | Auto-injects relevant memories into system prompt before each LLM call |

> **Note**: `loadmemorytool` and `preloadmemorytool` are only on `main` branch, not yet in any release tag.

## Detailed Descriptions

### `exitlooptool`

Sets `ctx.Actions().Escalate = true` and `SkipSummarization = true`. The `loopagent` checks this flag after each sub-agent iteration and exits when true.

- **When to use**: Only inside loops with escalate enabled
- **How it works**: Injected as a tool into agents that participate in a loop with `exitLoop: true`
- **Not a base tool**: Contextual to loop containers only

```go
import "google.golang.org/adk/tool/exitlooptool"

exitTool, _ := exitlooptool.New()
// Tool name: "exit_loop"
// Description: "Exits the loop. Call this function only when you are instructed to do so."
```

### `functiontool`

Factory for creating `tool.Tool` from any Go function. The function signature must be `func(tool.Context, ArgsStruct) (ResultType, error)`. ADK auto-generates the JSON schema from the args struct.

- **Used by**: `adk-utils-go` memory tools (`search_memory`, `save_to_memory`), `exitlooptool` itself
- **Not a tool itself**: It's a constructor

```go
import "google.golang.org/adk/tool/functiontool"

myTool, _ := functiontool.New(functiontool.Config{
    Name:        "my_tool",
    Description: "Does something useful",
}, myFunction)
```

### `mcptoolset`

Wraps an MCP (Model Context Protocol) server as an ADK `Toolset`. All tools exposed by the MCP server become available to the agent.

- **Already used in Magec**: `buildToolsets()` in `agent.go` creates these for each MCP server reference
- **Supports**: stdio (subprocess) and HTTP/SSE transports

```go
import "google.golang.org/adk/tool/mcptoolset"

ts, _ := mcptoolset.New(mcptoolset.Config{
    Transport: transport,
})
```

### `agenttool`

Wraps an agent as a callable tool. When the LLM calls this tool, it creates a new `runner.Runner`, a fresh in-memory session, and executes the target agent. Returns the agent's final text output.

- **Use case**: Agent-to-agent composition without flows. E.g., Magec could call Itahisa as a tool for a subtask
- **Key behavior**: Creates an isolated session (state is copied, but session is separate)
- **Config option**: `SkipSummarization` to avoid summarizing the sub-agent's output
- **Future potential**: Could be exposed in admin UI as "this agent can call these other agents as tools"

```go
import "google.golang.org/adk/tool/agenttool"

agentAsTool := agenttool.New(otherAgent, &agenttool.Config{
    SkipSummarization: false,
})
// Tool name = otherAgent.Name()
// LLM sees it as a function with a "request" string parameter
```

### `geminitool`

Provides Gemini-native tools (Google Search, Retrieval, etc.). These are server-side tools that execute within Google's infrastructure — no local code runs.

- **Not applicable for Magec**: Only works with Gemini models via Google's API
- **Includes**: `GoogleSearch` struct and a `New()` factory for any `genai.Tool`

```go
import "google.golang.org/adk/tool/geminitool"

// Built-in Google Search
search := geminitool.GoogleSearch{}

// Custom Gemini tool (e.g., Retrieval)
retrieval := geminitool.New("data_retrieval", &genai.Tool{
    Retrieval: &genai.Retrieval{...},
})
```

### `loadartifactstool`

Gives agents the ability to load artifacts (files, images, documents) stored in the session's artifact service. On startup, it injects instructions listing available artifacts and handles `load_artifacts` function calls by fetching the artifact content.

- **Depends on**: An `ArtifactService` being configured (in-memory or persistent)
- **Future use**: When implementing Telegram file support (from TODO.md), this tool would let agents access uploaded files
- **How it works**: Lists artifacts → injects instruction → LLM calls `load_artifacts(artifact_names)` → tool loads and appends content to context

```go
import "google.golang.org/adk/tool/loadartifactstool"

artifactsTool := loadartifactstool.New()
// Tool name: "load_artifacts"
```

### `toolconfirmation` (Human-in-the-Loop)

Not a tool to inject — it's a protocol built into the `tool.Context` interface. Any tool can request human confirmation before executing sensitive actions.

- **How it works**:
  1. Inside any tool's `Run()`, call `ctx.RequestConfirmation(hint, payload)`
  2. ADK emits a `FunctionCall` event with name `adk_request_confirmation`
  3. Client/UI listens for this event, shows confirmation prompt to user
  4. Client sends back a `FunctionResponse` with `{confirmed: true/false}`
  5. Tool reads `ctx.ToolConfirmation().Confirmed` to proceed or abort

- **What Magec needs**: A notification area in the admin UI that listens for `adk_request_confirmation` events and lets operators confirm/deny

```go
// Inside any tool's Run function:
func myTool(ctx tool.Context, args MyArgs) (map[string]any, error) {
    confirmation := ctx.ToolConfirmation()
    if confirmation == nil {
        ctx.RequestConfirmation("About to delete all records", args)
        return nil, nil // Execution pauses until confirmation
    }
    if !confirmation.Confirmed {
        return map[string]any{"status": "cancelled by user"}, nil
    }
    // Proceed with the action...
}
```

- **Event format**: `FunctionCallName = "adk_request_confirmation"` with args containing `toolConfirmation` (hint) and `originalFunctionCall` (the actual tool call)
- **Helper**: `toolconfirmation.OriginalCallFrom(functionCall)` extracts the original tool call from the confirmation wrapper

### `loadmemorytool` (main branch only)

LLM-callable tool that searches long-term memory by query. The LLM decides when to call it (e.g., when the user asks about past conversations). Returns matching memory entries.

- **Tool name**: `load_memory`
- **How it works**: LLM calls `load_memory(query)` → tool calls `ctx.SearchMemory()` → returns `{memories: [...]}`
- **Auto-injects instruction**: "You have memory. You can use it to answer questions. If any questions need you to look up the memory, you should call load_memory function with a query."
- **Magec equivalent**: `search_memory` from `adk-utils-go/tools/memory`. Same concept, different implementation
- **Key difference from current Magec approach**: The current `search_memory` + `save_to_memory` are custom tools from `adk-utils-go`. This is the official ADK version but only handles search (no save). Also uses `ctx.SearchMemory()` which goes through the ADK `memory.Service` interface directly, whereas the custom tools in `adk-utils-go` talk to Postgres directly

```go
import "google.golang.org/adk/tool/loadmemorytool"

memTool := loadmemorytool.New()
// Tool name: "load_memory"
// LLM calls it when it needs to recall information
```

### `preloadmemorytool` (main branch only)

Automatic memory injection tool — NOT called by the LLM. Runs transparently via `ProcessRequest` before every LLM call. Takes the user's current query, searches memory, and injects matching results into the system instructions as `<PAST_CONVERSATIONS>` context.

- **Tool name**: `preload_memory` (but LLM never sees it)
- **How it works**: On each LLM request → extracts user query → calls `ctx.SearchMemory()` → formats results with timestamps and authors → appends to system instructions
- **Key insight**: This is what Magec currently does manually via the `memoryInstruction` constant and the "CRITICAL: At the START of every conversation, you MUST call search_memory" prompt hack. `preloadmemorytool` does it automatically and cleanly — no need to instruct the LLM to search at startup
- **Complementary to `loadmemorytool`**: Use `preloadmemorytool` for automatic context (always have relevant memories), and `loadmemorytool` for on-demand deep searches when the LLM decides it needs more info

```go
import "google.golang.org/adk/tool/preloadmemorytool"

preload := preloadmemorytool.New()
// Not a function the LLM calls — it runs automatically via ProcessRequest
// Injects: "The following content is from your previous conversations..."
```

## FilterToolset (v0.4.0)

Utility to dynamically filter which tools from a toolset are exposed to the LLM based on runtime context.

```go
import "google.golang.org/adk/tool"

// Filter by tool names
filtered := tool.FilterToolset(myToolset, tool.StringPredicate([]string{"allowed_tool_1", "allowed_tool_2"}))

// Custom predicate based on context
filtered := tool.FilterToolset(myToolset, func(ctx agent.ReadonlyContext, t tool.Tool) bool {
    // Decide based on session state, user, etc.
    return someCondition
})
```

## Relevance for Magec

### Use now
- `exitlooptool` — For loop escalate (contextual injection, not base)
- `functiontool` — Already used indirectly
- `mcptoolset` — Already used

### Use soon
- `toolconfirmation` — Human-in-the-Loop for flows. Needs admin UI notification area
- `agenttool` — Agent-to-agent calls without flows. Configurable via admin UI

### Use later
- `loadartifactstool` — When implementing Telegram file/artifact support
- `loadmemorytool` + `preloadmemorytool` — When released in a tag, consider replacing `adk-utils-go` memory tools with these official ones. `preloadmemorytool` would eliminate the current prompt hack for startup memory search

### Skip
- `geminitool` — Gemini-only, not applicable with OpenAI/Anthropic backends

### Memory tools: `adk-utils-go` vs ADK official

Magec currently uses custom memory tools from `adk-utils-go/tools/memory` (v0.1.7 in use, v0.2.0 available):

| Tool | `adk-utils-go` v0.1.7 | `adk-utils-go` v0.2.0 | ADK official (`main`) |
|------|----------------------|----------------------|----------------------|
| **Search** | `search_memory` | `search_memory` (with entry IDs) | `loadmemorytool` (`load_memory`) |
| **Save** | `save_to_memory` | `save_to_memory` | — (no equivalent) |
| **Update** | — | `update_memory` (by entry ID) | — (no equivalent) |
| **Delete** | — | `delete_memory` (by entry ID) | — (no equivalent) |
| **Auto-preload** | — | — | `preloadmemorytool` (auto-injects memories) |

#### `adk-utils-go` v0.2.0 additions

The v0.2.0 introduces an `ExtendedMemoryService` interface (implemented by Postgres provider). When the memory service supports it, two extra tools are automatically registered:

- **`update_memory(id, content)`** — Updates an existing memory entry's content by ID. The ID comes from `search_memory` results (which now include `id` fields in v0.2.0)
- **`delete_memory(id)`** — Permanently deletes a memory entry by ID

These are enabled by default but can be disabled via `DisableExtendedTools: true` in the config. They only appear when the underlying `MemoryService` implements `memorytypes.ExtendedMemoryService` (Postgres does, in-memory does not).

#### Migration plan

The `memoryInstruction` in `agent.go` forces the LLM to call `search_memory` at the start of every conversation. `preloadmemorytool` would replace that pattern entirely by auto-injecting memories before the LLM even runs.

When ADK's memory tools reach a release tag:
1. Replace `search_memory` with `loadmemorytool` (on-demand search)
2. Add `preloadmemorytool` (automatic context injection)
3. Keep `save_to_memory`, `update_memory`, and `delete_memory` from `adk-utils-go` (ADK has no save/update/delete equivalents)
4. Remove the `memoryInstruction` constant and its "CRITICAL: At the START..." hack

**Short-term action**: Upgrade Magec from `adk-utils-go` v0.1.7 to v0.2.0 to get `update_memory` and `delete_memory` for Postgres-backed agents
