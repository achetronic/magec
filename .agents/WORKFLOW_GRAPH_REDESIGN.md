# Workflow Graph Redesign

Branch: `feat/workflow-graph-redesign`

Status: design locked, backend implementation in progress.

This document is the single source of truth for the rewrite of Magec's
"Agentic Flows" from a recursive tree onto the graph engine introduced in
adk-go v2.0.0. Keep it in lockstep with the code. If the model changes,
this file changes first.

## 1. Why

The current flow system is a tree in three coupled layers:

1. Data: `store.FlowStep` is a recursive node (`Steps []FlowStep`) whose
   `Type` is one of `agent`, `sequential`, `parallel`, `loop`.
2. Builder: `server/agent/flow.go` walks the tree and emits nested adk v1
   workflow agents (`sequentialagent`, `parallelagent`, `loopagent`).
3. Editor: `FlowCanvas.vue` plus `FlowBlock.vue` is a nested-container
   editor, not a graph editor. The structure is the nesting; there are no
   edges.

A tree forces three workarounds that a graph removes outright:

- Conditional loop exit is faked with a synthetic CEL evaluator agent
  (`flowexit`) injected as the loop's last child, which emits
  `Actions.Escalate=true` to make adk's `loopagent` stop. See
  `server/agent/flowexit/evaluator.go`.
- `exit_loop` is a tool that bubbles the same `Escalate` up to the nearest
  loop. Conditional, injected per subtree.
- `wrapAgent` exists only to dodge adk v1's single-parent constraint when a
  flow is reused as a step.

adk-go v2 ships a directed-graph workflow engine. A visual drag-and-drop
editor is a graph (boxes plus arrows), not a tree. Moving to the graph model
aligns all three layers with what the feature actually is, and lets us delete
the three workarounds instead of carrying more.

## 2. Decisions (locked with the user)

- D1: Pure graph (`nodes[]` plus `edges[]`), not an improved tree. All three
  layers are rewritten.
- D2: Break compatibility. The old `store.FlowStep` format is dropped.
  Saved flows are discarded. The breaking change is announced externally.
  No tree-to-graph importer.
- D3: CEL stays, but as a router node, not as an edge attribute. See section 4.
- D4: Human-in-the-loop (adk v2 pause/resume) is NOT in this redesign. It is
  a follow-up feature once the graph model has landed. The model leaves room
  for it (see section 8).
- D5: Backend first. Lock the data model and builder, get them compiling and
  tested, then point the visual editor at a stable shape.
- D6: Loops are raw edges (an edge back plus a router exit node), not a
  high-level `loop` sugar type. This is how adk-go exposes graphs.

## 3. The adk-go v2 contract (verified against source, tag v2.0.0)

Module path is `google.golang.org/adk/v2`, Go 1.25+.

The graph is a flat list of edges:

```go
type Edge struct {
    From  Node
    To    Node
    Route Route // nil means unconditional
}
```

Built with `workflow.NewEdgeBuilder()` (`.Add`, `.AddRoute`, `.AddFanOut`,
`.AddFanIn`, `.AddRoutes`, `.Build`) or the helpers `workflow.Chain(nodes...)`
and `workflow.Concat(...)`. `workflow.Start` is the entry sentinel node.

Node kinds we use:

- `AgentNode` via `workflow.NewAgentNode(a agent.Agent, cfg)`: wraps a normal
  agent. The wrapped agent must emit its final output as `Event.Output` to
  feed successors.
- `FunctionNode` via `workflow.NewFunctionNode[IN, OUT]` and the streaming
  `workflow.NewEmittingFunctionNode`: a Go function as a node. This is our
  router.
- `JoinNode` via `workflow.NewJoinNode`: a fan-in barrier. Fires once after
  every declared predecessor completes, receiving their outputs as a single
  `map[string]any` keyed by predecessor name. Conditional routing into a
  JoinNode is a configuration error.

Routing is decided by the source node, not by the edge. A node emits an event
with `ev.Routes = []string{label}`; the edge's `Route` (`StringRoute`,
`IntRoute`, `BoolRoute`, `MultiRoute[T]`, or `workflow.Default`) matches
against that label. `workflow.Default` matches when no concrete route does.

`NodeConfig` gives, per node and for free: `RetryConfig` (attempts, backoff,
jitter, `ShouldRetry`), `Timeout`, and `ParallelWorker` (run the node once per
item of a list input and aggregate).

The whole graph is exposed as a normal `agent.Agent` through:

```go
workflowagent.New(workflowagent.Config{
    Name, Description, SubAgents, Edges,
}) (agent.Agent, error)
```

This is what the Magec builder returns, replacing the nested
sequential/parallel/loop agents. The single-parent constraint disappears
because the graph is flat.

## 4. CEL as a router node (D3)

In adk v2 an edge does not evaluate anything. So CEL cannot live on the edge.
Instead a router node evaluates CEL against the shared flow state and emits a
route label. Outgoing edges carry `StringRoute(label)` and match it.

Router semantics: an ordered list of rules. The first rule whose CEL guard is
true emits its route label. If none match, the node emits the default label.
Each distinct label corresponds to exactly one outgoing edge.

```
router "decide" {
    rule { when: "state.score >= 0.8", route: "accept" }
    rule { when: "state.score <  0.3", route: "reject" }
    default: "revise"
}
edges:
    decide --route accept--> publish
    decide --route reject--> discard
    decide --route revise--> rewrite
```

This is strictly more powerful than today's `flowexit` hack, which could only
kill the enclosing loop. A router can send flow to any node: leave a loop,
re-enter it, branch elsewhere.

CEL reuse: the existing `flowexit.Compile` / `EvaluateExitWhen` logic and the
`flow:` state namespace (`flowstate.StateKeyPrefix`) are kept. The synthetic
`NewExitWhenAgent` and its `Escalate` emission are deleted. The router reads
flow state through `ctx.state` (adk v2 `NewFunctionNodeFromState` or a direct
`ctx.State()` read under the `flow:` prefix).

The `flowstate` set_state / get_state toolset stays. Agents inside a flow
still need a shared scratchpad; the router just reads the same namespace.

## 5. New data model (server/store)

The recursive `FlowStep` is gone. A flow is a set of named nodes plus a set of
directed edges.

```go
// FlowNodeType identifies the kind of graph node.
const (
    FlowNodeAgent  = "agent"  // wraps an AgentDefinition or a sub-flow
    FlowNodeRouter = "router" // evaluates CEL rules, emits a route label
    FlowNodeJoin   = "join"   // fan-in barrier
)

// FlowNode is one vertex of the flow graph.
// ID is unique within the flow and becomes the adk Node.Name().
type FlowNode struct {
    ID            string        `json:"id"`
    Type          string        `json:"type"`
    AgentID       string        `json:"agentId,omitempty"`       // type=agent
    ResponseAgent bool          `json:"responseAgent,omitempty"` // type=agent
    Rules         []FlowRule    `json:"rules,omitempty"`         // type=router
    DefaultRoute  string        `json:"defaultRoute,omitempty"`  // type=router
}

// FlowRule is one ordered branch of a router node.
type FlowRule struct {
    When  string `json:"when"`  // CEL expression over `state`
    Route string `json:"route"` // emitted label when When is true
}

// FlowEdge is a directed connection. Route empty means unconditional.
// Route is only meaningful when From is a router node; it matches the
// label the router emits.
type FlowEdge struct {
    From  string `json:"from"`  // FlowNode.ID
    To    string `json:"to"`    // FlowNode.ID
    Route string `json:"route,omitempty"`
}

// FlowDefinition is a multi-agent workflow stored as a directed graph.
type FlowDefinition struct {
    ID          string     `json:"id"`
    Name        string     `json:"name"`
    Description string     `json:"description,omitempty"`
    Entry       string     `json:"entry"` // FlowNode.ID; builder wires Start -> Entry
    Nodes       []FlowNode `json:"nodes"`
    Edges       []FlowEdge `json:"edges"`
    A2A         *A2AConfig `json:"a2a,omitempty"`
}
```

Notes:

- `Entry` is the single source of truth for where the graph starts: the
  builder synthesizes a `Start -> Entry` edge, so operators never wire the
  Start sentinel themselves and `FlowEdge` never references it. We keep it explicit
  rather than inferring the single source, because a graph can legitimately
  have several roots and we want the operator's intent recorded.
- Node IDs are the naming scheme. The old synthetic `flowID_path` pairing
  between `flow.go` and `FlowDefinition.ResponseAgentNames()` is gone. The
  response filter matches `event.Author` against the IDs of nodes flagged
  `ResponseAgent`. Far simpler, no lockstep convention.
- A sub-flow is still an `agent` node whose `AgentID` resolves to another
  flow. adk v2 `WorkflowNode` (nested workflow) is the clean substrate; until
  it is wired, a built sub-flow agent is referenced like any other agent.

## 6. Validation (replaces validateFlowStep)

The admin API validates the graph at save time:

1. Every `FlowEdge.From` and `.To` references an existing node ID. Edges must
   not reference the reserved START sentinel (the builder wires Start->Entry).
2. `Entry` references an existing node.
3. `agent` nodes have a non-empty `AgentID`.
4. `router` nodes have at least one rule; every `FlowRule.When` compiles as
   CEL (`flowexit.Compile`); every rule `Route` and the `DefaultRoute` has a
   matching outgoing edge; every outgoing edge of a router has a `Route` that
   some rule or the default emits.
5. No conditional edge targets a `join` node (adk constraint: a barrier waits
   for all declared predecessors, so a route-skipped predecessor would
   deadlock it).
6. Node IDs are unique and match a safe identifier pattern (they become adk
   node names and session-state key fragments).
7. The graph is connected from `Entry` (no orphan nodes) and has at least one
   terminal node.

Cycles are allowed: a loop is a back edge. The runtime guards against runaway
loops; see open question O1.

## 7. Builder rewrite (server/agent/flow.go)

`BuildFlowAgent` stops recursing a tree and instead:

1. Builds one adk node per `FlowNode`:
   - `agent`: `workflow.NewAgentNode(builtAgent, cfg)` where `builtAgent`
     comes from `BuildAgentInstance` (unchanged), named after `FlowNode.ID`.
   - `router`: a `workflow.NewEmittingFunctionNode` that loads flow state,
     evaluates the CEL rules in order, and emits `ev.Routes = []string{label}`.
   - `join`: `workflow.NewJoinNode(id)`.
2. Translates `FlowEdge` to `workflow.Edge`, mapping the entry to
   `workflow.Start` and attaching `workflow.StringRoute(route)` where set.
3. Returns `workflowagent.New(workflowagent.Config{Name, Description, Edges})`.

Deleted with this rewrite:

- `server/agent/flowexit/` synthetic agent and its `Escalate` path (the CEL
  compile/eval helpers move into the router node).
- `exit_loop` tool injection.
- `wrapAgent` and the single-parent dance.
- `buildStep`, `buildChildren`, the `flowID_path` naming, and
  `FlowDefinition.ResponseAgentNames()`.

Kept: `BuildAgentInstance`, the `flowstate` toolset, per-node retry/timeout
(now first class via `NodeConfig`).

## 8. Phasing

Cutover order, backend first (D5):

1. Data model: new `store` types plus graph validation plus tests.
   Compiles independently of adk v2 (pure data).
2. adk v2 bump: `go.mod` to `google.golang.org/adk/v2`, Go 1.25+. This is a
   module-wide migration (the `session.NewEvent(ctx, ...)` signature change
   and the `ToolContext`/`CallbackContext` unification touch code outside
   flows). Treated as its own commit so the blast radius is visible.
3. Builder rewrite against the v2 workflow package. Delete the workarounds.
4. Admin API and serialization wired to the new shape; old endpoints updated.
5. Visual editor rewrite: node-and-edge canvas with real connectors.

HITL (D4) lands after step 5 as a new node capability: a router or function
node that emits `workflow.NewRequestInputEvent` and returns
`ErrNodeInterrupted`, with `NodeConfig.RerunOnResume` controlling re-entry
versus handoff. RunState already persists in `session.State`, so the data
model needs no change for it. This is why the model is left HITL-ready.

## 9. Open questions (to resolve during implementation)

- O1 (loop cap): RESOLVED. A raw back edge has no `MaxIterations`, and adk v2
  does not cap loops: its scheduler resets a node's lifecycle on every
  loop-back (`scheduler.go`: "loop-back routing starts a fresh lifecycle")
  and `NodeState.Attempt` counts only retry-on-failure, not loop turns. adk
  v2 only rejects *unconditional* cycles (`validateCycles`, `ErrUnconditionalCycle`);
  conditional cycles are legal. Since every loop in our model passes through a
  router (only routers emit routed edges), every loop is conditional and thus
  legal — we mirror that check in our own `Validate` so an all-unconditional
  cycle fails early with a clear message.

  The loop turn count lives in flow state (the `flow:` namespace, which
  survives loop-backs unlike the node lifecycle) under the per-router key
  `flow:__iterations:<routerID>`. The router increments it before evaluating
  and exposes it to the CEL environment as the integer variable `iterations`,
  alongside the existing `state` map. The operator writes the loop exit as any
  other condition, e.g. `state.done == true || iterations >= 5`. This keeps
  D6 (raw edges, no loop sugar) and does not touch the validated data model;
  it only extends the CEL env in `flowexit` (touched anyway at cutover).

  Hard runaway guard: a global constant `maxLoopIterations` (e.g. 1000) caps
  any single router's activation count. Exceeding it terminates the workflow
  with an error (fail-safe, never silently continue). Not operator-configurable
  for now; a future per-router `MaxIterations` field can be added without
  breaking the model. The variable name `iterations` is chosen to match the
  vocabulary of the old `loopagent` `MaxIterations`, so migrating operators
  reuse the same mental model.
- O2 (flow state to ctx.state): confirm whether the router reads through the
  existing `flow:` session-state namespace or adk v2's typed
  `NewFunctionNodeFromState`. Affects only the builder, not the model.
- O3 (sub-flow as node): wire adk v2 `WorkflowNode` for nested flows versus
  referencing a pre-built sub-flow agent. Affects the builder.
- O4 (voice config): RESOLVED. `FirstAgentID` now returns the AgentID of the
  first agent node reached from `Entry` breadth-first, falling back to the
  first agent node in declaration order, then "". Implemented on the graph in
  store.
