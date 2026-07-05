---
title: "Skills vs. Specialized Agents"
---

A common question when building on Magec: should you create **one agent with many skills**, or **many specialized agents**? Both approaches work. The right choice depends on what you're building.

This page helps you decide.

## One agent, many skills

You have a single agent and attach multiple skills to it: product knowledge, return policy, shipping rules, escalation procedures. The agent handles everything.

**When this works best:**

- The user talks to one entity and expects it to know everything
- Topics overlap and the agent needs to combine knowledge in a single response
- You want a simple setup with one conversation thread
- Example: a **customer support bot** that answers questions about orders, returns, products, and shipping in one chat

**What it looks like:**

```
Agent: Support Bot
  ├── Skill: Product Catalog
  ├── Skill: Return Policy
  ├── Skill: Shipping Rules
  └── Skill: Escalation Procedures
```

The user writes *"I want to return the headphones I bought last week, and also ask about shipping times for the replacement."* The agent handles both topics in one response because it has all the skills loaded.

**Trade-offs:**

- The agent's context grows with each skill (instructions + reference files all get injected). Very large contexts can reduce response quality or hit token limits.
- The agent must decide which skill's knowledge applies to each question. It usually gets this right, but ambiguous cases can happen.
- All skills share the same LLM and system prompt personality.

## Many specialized agents

You create separate agents, each focused on one domain. Each agent has its own system prompt, skills, tools, and potentially its own LLM model. You can connect them through a [flow](/docs/flows/) or let users switch between them.

**When this works best:**

- Each domain requires different tools (MCP servers), models, or personalities
- You want to optimize cost by using cheaper models for simple tasks
- Responses need to be highly specialized and precise
- You're building a pipeline where each step is handled by an expert
- Example: a **multi-agent pipeline** where a researcher gathers data, an analyst processes it, and a writer formats the report

**What it looks like:**

```
Agent: Product Expert     → Skill: Product Catalog, MCP: inventory DB
Agent: Returns Specialist → Skill: Return Policy, MCP: order system
Agent: Shipping Advisor   → Skill: Shipping Rules, MCP: tracking API
```

Each agent has a focused system prompt, the right tools for its job, and potentially a different model. The returns agent connects to the order management system. The shipping agent connects to the tracking API. They don't need each other's tools or knowledge.

**Trade-offs:**

- More entities to manage in the Admin UI
- Users need to switch agents (via `/agent` in Telegram, agent switcher in Voice UI) or you need a flow to route them
- Context doesn't carry between agents unless you use flows with `outputKey`

## Decision guide

| Factor | One agent + skills | Many specialized agents |
|--------|-------------------|----------------------|
| **User experience** | Single conversation, one entity knows all | User switches agents or a flow routes automatically |
| **Context size** | Grows with each skill added | Each agent has a focused, smaller context |
| **Tools (MCP)** | All tools loaded on one agent | Each agent gets only the tools it needs |
| **LLM cost** | One model handles everything | Mix expensive models for hard tasks, cheap ones for simple tasks |
| **Maintenance** | Skills are modular and reusable | Agents are independent but more numerous |
| **Best for** | Support bots, general assistants, FAQ | Pipelines, expert systems, multi-step workflows |

## The hybrid approach

In practice, the best setups combine both patterns. Use skills to make individual agents knowledgeable, and use multiple agents when the domains are truly different.

**Example: a business platform**

```
Agent: Customer Support
  ├── Skill: Product Catalog
  ├── Skill: Return Policy
  ├── Skill: FAQ
  └── MCP: Order System

Agent: Sales Assistant
  ├── Skill: Product Catalog  ← same skill, reused
  ├── Skill: Pricing Rules
  └── MCP: CRM

Agent: DevOps Monitor
  ├── Skill: Runbook Procedures
  ├── MCP: Kubernetes
  └── MCP: PagerDuty
```

The Product Catalog skill is shared between Customer Support and Sales. Update it once and both agents benefit. But Customer Support and DevOps Monitor are completely different domains with different tools, so they're separate agents.

## Flows: the best of both worlds

[Agentic Flows](/docs/flows/) let you chain specialized agents into pipelines. Each agent does what it's best at, and the output feeds into the next step.

```
Flow: Weekly Business Report
  1. Data Agent    → queries databases, produces raw numbers
  2. Analyst Agent → interprets trends, flags anomalies
  3. Writer Agent  → formats everything into a readable report
```

Each agent in the flow has its own skills, tools, and model. The Data Agent might use a cheap, fast model with a database MCP. The Analyst might use a reasoning-heavy model. The Writer might use a creative model with a "report template" skill.

The combination pays off: each agent is an expert at its step, and skills provide the domain knowledge it needs.

{{< callout type="info" >}}
Start simple. One agent with a few skills covers most use cases. Split into multiple agents when you notice that one agent is trying to do too many different things, needs different tools for different tasks, or when response quality drops because the context is too large.
{{< /callout >}}
