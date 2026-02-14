---
title: "Agentic Flows"
---

Flows chain multiple agents into multi-step workflows. Build them visually with the drag-and-drop editor in the Admin UI, or define them as JSON.

The idea is simple: start small. A flow can be as easy as two agents running one after the other — one writes, the other reviews. From there, you can grow into complex pipelines with parallel branches, loops, and dozens of agents collaborating. The editor is the same; the complexity is up to you.

## Step types

| Type | Behavior |
|------|----------|
| `agent` | Run a single agent. Leaf node — its output becomes input for the next step. |
| `sequential` | Run child steps one after another, in order. The main building block for pipelines. |
| `parallel` | Run child steps simultaneously. Results are merged and passed forward. |
| `loop` | Repeat child steps until an agent calls the `exit_loop` tool, or `maxIterations` is reached. |

Steps can be nested freely — a sequential step can contain parallel branches, a parallel branch can contain loops, a loop can contain more sequences, etc.

## From simple to complex

A basic flow might be a sequential pipeline with 3 agents: a researcher, a writer, and an editor. Each one takes the output of the previous one and builds on it.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-flow-simple.png" alt="Admin UI — Simple 3-agent flow" >}}
</div>

But the same editor lets you build much larger workflows. The Software Factory example below chains 13 agents through a full SDLC: product manager → architect → developers → QA → documentation → deployment — with parallel branches and loops for iterative refinement.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-flows.png" alt="Admin UI — Software Factory flow (13 agents)" >}}
</div>

## Built-in example flows

- **Research Pipeline** — Parallel research + critique → fact-checking → synthesis. 4 agents working together.
- **Debate Arena** — Two debaters argue with a moderator controlling turns via loop. 3 agents in a structured debate.
- **Software Factory** — 13-agent SDLC pipeline: product manager → architect → developers → QA → docs → deployment.

## How data flows

Each step receives the accumulated output of all previous steps as context. The final step's output is returned to the user. In parallel steps, all branches receive the same input and their outputs are concatenated.
