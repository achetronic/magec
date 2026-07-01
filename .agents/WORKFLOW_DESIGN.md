# Workflow Design

Reference for Magec's Agentic Flows: the data model, how it maps onto the
adk-go v2 workflow engine, every node type, validation, runtime state, the
code-node sandbox, and the visual editor. Read this to reconstruct the system
or to work on it.

## Overview

A flow is a directed graph of typed nodes joined by edges. It is stored as a
`store.FlowDefinition`, validated at save time by `server/agent/flowgraph`,
and built into a single adk-go v2 workflow agent by `server/agent/flow.go`.
The built agent is registered alongside regular agents and addressable by the
flow ID.

Engine: `google.golang.org/adk/v2` (packages `workflow` and
`agent/workflowagent`). A flow becomes `workflowagent.New(Config{Name, Edges})`,
where `Edges` is a flat `[]workflow.Edge{From, To, Route}`.

## Design principles

The intent behind the shape of the system, which the concrete values in code
serve:

- **Two expression languages, by purpose.** CEL is used for small, safe,
  guaranteed-terminating snippets: router rule guards (return a bool) and
  expression-node transforms (return a value). Starlark (via starlet) is used
  for the code node, where an operator wants real logic (loops, functions,
  multi-step transforms). CEL is the default for a one-liner; the code node is
  the escape hatch for genuine programs.
- **Starlark is deliberately extended.** The code node loads starlet's module
  libraries (json, http, hashing, regex, and more) so a script can do useful
  work, not just arithmetic. Which libraries are available is an admin decision
  (see the code-node section), not a fixed set baked into the model.
- **Dedicated blocks over generic ones.** Expression, template and code are
  distinct node types even though all three transform data, because each states
  its intent: a CEL one-liner, a text template for assembling prompts, or a
  full script. The operator picks the least powerful block that fits.
- **Node labels favour the operator's mental model over adk's.** The editor
  labels a block by what it does for a flow author, not by the adk primitive
  underneath or the store `Type` (for example the `parallel` node is shown as
  "Foreach" because it runs an agent once per list item, and an embedded flow
  is a "Subflow" though it is an adk WorkflowNode). The store `Type` and the adk
  construct are an implementation detail; the label communicates purpose.
- **Node size signals intent.** Nodes with growable content (router rules and
  the text nodes: expression, template, code) open larger and stay resizable so
  long content is readable; selection-only nodes (agent, subflow, join) have a
  fixed compact size and no resize handle.
- **Connection lines are the Magec accent (sol/amber).** Edges and the active
  connection drag use the sol accent so the graph reads as one coherent Magec
  surface; entity colours are reserved for the node cards.
- **Admin governs capability, author governs logic.** What a code node may
  reach (libraries) and its resource ceilings are set by whoever deploys
  Magec; the flow author writes the graph and the scripts within those bounds.

## Data model (`server/store/types.go`)

```go
type FlowDefinition struct {
    ID          string
    Name        string
    Description string
    Entry       string     // FlowNode.ID where execution starts
    Nodes       []FlowNode
    Edges       []FlowEdge
    StartX, StartY float64  // editor position of the Start box (layout hint)
    A2A         *A2AConfig
}

type FlowNode struct {
    ID   string // unique in the flow; becomes the adk Node.Name() and event.Author
    Type string // agent | router | join | parallel | subflow | expression | template | code

    AgentID       string     // agent, parallel: the AgentDefinition to run
    ResponseAgent bool       // agent: include this node's output in the flow response
    Rules         []FlowRule // router: ordered CEL branches
    DefaultRoute  string     // router: label emitted when no rule matches
    MaxConcurrency int       // parallel: max items in flight, 0 = unlimited
    FlowID        string     // subflow: the FlowDefinition to embed
    Expression    string     // expression: CEL value over input + state
    Template      string     // template: text with {{ input }} / {{ state.key }}
    Script        string     // code: Starlark source, must assign `output`
    TimeoutMs     int        // code: per-node wall-clock ceiling, 0 = inherit Settings
    MaxOutputBytes int       // code: per-node output cap, 0 = inherit Settings
    OutputKey     string     // expression, template, code: also write output to state.<key>

    X, Y, W, H    float64    // editor layout hints, ignored by builder and validation
}

type FlowRule struct {
    When  string // CEL expression over `state` (map) and `iterations` (int); must return bool
    Route string // label emitted when When is true
}

type FlowEdge struct {
    From  string // FlowNode.ID
    To    string // FlowNode.ID
    Route string // only for edges leaving a router: the label to match; empty otherwise
}
```

Invariants:

- A node ID is the adk `Node.Name()` and therefore the `event.Author`. Output
  filtering (webhook/cron) matches these IDs, so no synthetic naming exists.
- `Entry` names the start node. The builder wires the `workflow.Start` sentinel
  to it; no `FlowEdge` references the sentinel.
- `OutputKey` must be a valid CEL identifier (letters, digits, underscore; no
  hyphen) so `state.<key>` parses downstream.
- Layout fields (`X/Y/W/H`, `StartX/StartY`) are admin-UI only; the builder and
  validation ignore them.

## Node types and how each maps to adk

Every node is created in `server/agent/flow.go` `buildNode`. Nodes that emit a
value do so through `workflow.NewEmittingFunctionNode`, yielding a
`session.Event` with `Author = node.ID`, `Output = <value>`, and, when
`OutputKey` is set, `Actions.StateDelta[flow:<OutputKey>] = <value>`.

- **agent**: wraps the AgentDefinition (`AgentID`) built by `BuildAgentInstance`
  into `workflow.NewAgentNode`. The shared flow-state toolset is injected so
  the agent can `set_state`/`get_state`.
- **router** (`server/agent/router_node.go`): a function node. Reads flow state,
  evaluates `Rules` in order, emits the first matching `Route` (or
  `DefaultRoute`) as `ev.Routes`. Outgoing edges carry `workflow.StringRoute`
  labels that match. Passes `input` through unchanged. Increments a per-router
  activation counter and exposes it to CEL as `iterations`.
- **join**: `workflow.NewJoinNode`, a fan-in barrier that fires once after all
  declared predecessors complete. Routed edges into a join are a config error.
- **parallel**: wraps the `AgentID` agent node in `workflow.NewParallelWorker`,
  running it once per item of a list-typed input, `MaxConcurrency` in flight,
  aggregating per-item outputs into a list.
- **subflow**: builds the referenced flow (`FlowID`) into a
  `workflow.NewWorkflowNode` from that flow's own edges (`buildEdges` is shared
  by the top-level flow and every subflow).
- **expression** / **template** (`server/agent/transform_nodes.go`): transform
  nodes. Expression evaluates a CEL value; template renders placeholders. Both
  emit a value and honour `OutputKey`.
- **code** (`server/agent/code_node.go`): runs Starlark. See the sandbox
  section.

## Routing and loops

Edges do not evaluate anything. Routing is decided by the source node: a router
emits `ev.Routes = [label]` and each outgoing edge's `Route` matches against it
(`StringRoute`). An edge with an empty `Route` is unconditional and only valid
from a non-router node.

Sequencing is a chain of edges. Fan-out is several edges from one node,
reunified by a `join`. A loop is a back edge (an edge whose target is an
earlier node) gated by a router: one route continues the loop, another leaves
it. There is no loop container node.

The router exposes `iterations` (its own activation count this run) to CEL, so
an operator writes a loop exit as `state.done == true || iterations >= 5`.
`maxLoopIterations` (a builder constant) is a hard runaway guard that fails the
workflow if a single router activates past it.

## Shared flow state

Agents in a flow share a key-value scratchpad via the `set_state`/`get_state`
tools (`server/agent/tools/flowstate`), backed by `session.state` under the
`flow:` prefix, visible to every node in the same flow for the conversation.
Router and expression nodes read it through
`flowexit.ExtractFlowState(ctx.State())`, which strips the `flow:` prefix and
hides internal keys (the `flow:__iterations:` counters). `OutputKey` on
transform and code nodes writes into the same namespace.

## CEL (`server/agent/flowexit`)

Two environments, both from `github.com/google/cel-go`:

- Router guards: variables `state` (map) and `iterations` (int); must return
  `bool`. Compiled by `Compile`, evaluated by `Evaluate` (a runtime error or a
  non-bool result is treated as `false` and logged).
- Expression nodes: variables `input` (dyn) and `state` (map); returns any
  value. Compiled by `CompileValue`, evaluated by `EvaluateValue` (a runtime
  error fails the node). String and list extensions are enabled, so
  `input.split(",")` yields a list and `.map()`/`.filter()` are available.

## Code node sandbox (`server/agent/code_node.go`)

Runs user Starlark via `github.com/1set/starlet`. The script receives `input`
(the upstream node's output) and `state` (flow state map) as globals and must
assign a top-level `output`; that value becomes the node output (nil if unset).

Capability is an admin decision, not the flow author's:

- Every starlet built-in module ships enabled. The admin disables modules in
  `Settings.Flows.DisabledLibraries`. The enabled loader list is prebuilt once
  in `agent.New` (`GetAllBuiltinModuleNames` minus the disabled set) and shared
  across executions; a fresh `Machine` is built per execution (nodes run
  concurrently, e.g. inside a parallel worker).

Limits guard Magec's own availability and have two levels:

- `Settings.Flows.ExecutionTimeoutMs` and `.MaxOutputBytes` are global hard
  ceilings. A node's `TimeoutMs`/`MaxOutputBytes` may only ask for less:
  `effective = min(node, ceiling)`, with 0 meaning "no limit / inherit"
  (`effectiveLimit`). When a timeout is active the machine also gets a Starlark
  step budget, so a tight infinite loop is cut by whichever fires first
  (step budget or context deadline). Output is measured by JSON-marshalling the
  result; over-cap output fails the node before it is emitted or written to
  state.

## Validation (`server/agent/flowgraph/validate.go`)

`Validate(*store.FlowDefinition)` runs on save (admin API `flows.go`):

- Unique, safe node IDs (`[a-zA-Z_][a-zA-Z0-9_-]*`); `START` is reserved.
- Per type: agent/parallel need `AgentID`; router needs at least one rule, a
  default route, and every rule `When` must compile as CEL; subflow needs
  `FlowID`; expression needs `Expression` (must compile); template needs
  `Template`; code needs `Script`. `OutputKey`, when present, must match the
  state-key pattern. `TimeoutMs`/`MaxOutputBytes` must be non-negative.
- `Entry` exists. Every edge endpoint exists; edges never reference `START`.
- Router coherence: each label a router can emit has exactly one outgoing edge,
  and every routed edge's label is one the router can emit; non-router nodes
  have no routed outgoing edges.
- No routed edge targets a join.
- Every node is reachable from `Entry`; at least one terminal node exists.
  Conditional cycles (loops through a router) are allowed; a fully
  unconditional cycle is rejected.

## Topological ordering

`sortFlowsTopologically` (`server/agent/agent.go`) orders flow builds so a
subflow is built before the flow that embeds it, and detects cycles across
flows. `FlowDefinition.AgentIDs()` reports every referenced entity (agent and
parallel `AgentID`, subflow `FlowID`) for this ordering.

## Settings (`Settings.Flows`)

```go
type FlowsSettings struct {
    DisabledLibraries  []string // starlet modules turned off; empty = all enabled
    ExecutionTimeoutMs int      // global code-node timeout ceiling, 0 = no limit
    MaxOutputBytes     int      // global code-node output ceiling, 0 = no limit
}
```

The initial settings for a fresh store define sane defaults (see the store's
initial-settings construction in `store.go`). Exposed in the admin UI Settings
"Flows" section.

## Visual editor (`frontend/admin-ui/src/views/flows`)

- `FlowCanvas.vue`: the canvas. Pan/zoom, an SVG bezier edge layer, drag from a
  node's output port to another node to connect, a grouped add-node toolbar
  (execution / flow control / data), a full-screen toggle, and a draggable
  Start box whose port sets the entry.
- `FlowNode.vue`: one node card per type, colour-coded per its entity colour
  (see `ENTITY_COLORS.md`), with the per-type body (agent picker, router rules,
  code editor with limit overrides, etc.). Nodes with growable content are
  resizable and persist `w/h`; selection-only nodes are fixed-size. Labels are
  operator-facing (the `parallel` type shows as "Foreach"). Entry is marked
  with a dot.
- `FlowDialog.vue`: wraps the canvas in the flow create/edit modal, binds the
  `{entry, nodes, edges}` model, serialises nodes/edges on save. The modal is
  persistent (Escape does not close it) so an unsaved graph is not lost.

## Key files

- `server/store/types.go`: `FlowDefinition`, `FlowNode`, `FlowEdge`,
  `FlowRule`, `FlowsSettings`, node-type constants, `Settings.Flows`, and the
  `AgentIDs`/`ResponseAgentIDs`/`ResponseAgentNames`/`FirstAgentID` helpers.
- `server/agent/flow.go`: `BuildFlowAgent`, `buildEdges`, `buildNode`.
- `server/agent/router_node.go`, `transform_nodes.go`, `code_node.go`: node
  builders.
- `server/agent/flowexit/`: CEL compile/evaluate and flow-state extraction.
- `server/agent/flowgraph/validate.go`: graph validation.
- `server/api/admin/flows.go`, `server/api/admin/settings.go`: admin API.
- `frontend/admin-ui/src/views/flows/`, `.../settings/FlowsSection.vue`: editor
  and settings UI.
