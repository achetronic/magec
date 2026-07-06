---
title: "Context Guard"
badge: "Experimental"
---

{{< callout type="warning" >}}
Context Guard is **experimental**. It works and has been tested with real conversations, but there are edge cases not fully covered yet. Read [Limitations](#limitations) for details.
{{< /callout >}}

AI models can only handle a certain amount of text at once. This is called the **context window** (think of it as the model's short-term memory). The longer a conversation gets, the more of that memory it uses up. When it fills up, the conversation breaks.

Context Guard keeps that from happening. It watches the conversation size and, when things start getting tight, compresses the older messages into a short summary. The recent messages stay exactly as they are. The agent doesn't lose track of anything; the conversation just takes up less space.

**Without Context Guard**, long conversations eventually fail. **With it**, they can go on as long as you need.

## How it works

Before every message is sent to the model, Context Guard checks the conversation:

1. **Is it getting too long?** It checks either the token count or the number of messages, depending on which strategy you chose.
2. **No.** Nothing happens. Everything goes through normally.
3. **Yes.** It splits the conversation in two: the **old stuff** and the **recent stuff**.
4. It asks the agent's own model to write a summary of the old stuff.
5. It swaps out all those old messages for the summary.
6. The model now sees: `[summary of earlier conversation] + [recent messages in full]`.

The summary is saved between messages, so it doesn't disappear. Each time Context Guard runs again, it folds the previous summary into the new one. Nothing is forgotten; it just gets more compact over time.

If anything goes wrong during summarization (model error, timeout, empty response), Context Guard steps aside and lets the original conversation through. It never blocks anything.

## Strategies

Two options. You pick one per agent.

### Token threshold

**Recommended for most agents.** It estimates how many tokens the conversation is using and only compresses when it's running out of room.

How much room it keeps:

- **Models with big context windows (over 200k tokens):** Keeps 20,000 tokens free.
- **Smaller models:** Keeps 20% of the window free.

When the conversation eats into that reserved space, Context Guard kicks in. It keeps the most recent 20% of the conversation intact and summarizes everything before that.

This is the best choice when the agent uses tools, does complex work, or when you want to get the most out of the model before compressing anything.

### Sliding window

Simpler. It counts messages. When there are more than `maxTurns` messages, everything except the last `maxTurns` gets summarized.

It doesn't look at token counts at all, just the number of messages.

Good for chatbots, Q&A agents, or any case where old messages stop being useful quickly.

{{< callout type="info" >}}
**Heads up if the agent uses tools.** Tool calls add hidden messages: the model's request to call the tool and the tool's response each count as separate messages. One question where the agent calls a tool creates 4 messages, not 2. With `maxTurns: 20`, that's only about 5 tool-using exchanges before summarization fires. If the agent uses tools a lot, set a higher value (40–80) or just use token threshold instead.
{{< /callout >}}

## Setup

Open an agent in the Admin UI, expand the **LLM** section, and turn on **Context Guard**.

| Setting | What it does |
|---------|-------------|
| Enabled | Turns Context Guard on or off |
| Strategy | `Token threshold` (default) or `Sliding window` |
| Max turns | Sliding window only: how many messages to keep. Default: `20` |

If it's off, nothing changes for the agent. Zero overhead.

### Which strategy to pick

| Agent type | Strategy | Why |
|---|---|---|
| Agent with tools | Token threshold | Uses the full context window, only compresses when needed |
| Simple chatbot | Sliding window, `maxTurns: 30` | Keeps things light, old messages rarely matter |
| Long multi-step tasks | Token threshold | Needs all the context it can get to track progress |

## What goes into the summary

Context Guard asks the agent's own model to write a summary with four sections:

- **Current State.** What's being worked on, what's done, what's next
- **Key Information.** Names, dates, numbers, URLs, preferences, specifics
- **Context & Decisions.** What was decided and why, what was tried and dropped
- **Exact Next Steps.** What to do next, specifically, not just "keep going"

The idea: someone reading only this summary should be able to continue the conversation without asking "what were we talking about?"

Summaries are longer for models with big context windows and shorter for smaller models. There's both a soft limit (in the prompt) and a hard limit (on the API) to make sure the summary itself doesn't take up too much space.

If the model returns nothing (rare), Context Guard falls back to grabbing the first 200 characters of each message. Not pretty, but it keeps things moving.

## Limitations

### Tool responses disappear from summaries

When old messages get summarized, tool calls show up as just `[tool X returned a result]`. The actual data the tool returned (files, search results, API responses) is not included in the summary prompt.

The agent's conclusions based on that data will still be in the conversation, but the raw data itself is gone. This is how Google's [ADK Python](https://github.com/google/adk-python) handles it too.

### Huge tool responses can cause trouble

If a tool dumps a massive response into the conversation (100k+ tokens), that whole blob is sitting in the message history. Context Guard will catch it on the *next* message and compress, but the current message already has the giant response in it.

With the token threshold strategy this usually isn't a problem: the safety buffer absorbs it. But if a single tool response is bigger than the whole buffer, the model might reject the request before Context Guard can do anything about it.

**How to avoid this:** Don't let your tools return unlimited data. Limit results, paginate, or return a reference instead of the full content. [Claude Code](https://github.com/charmbracelet/crush) caps every tool: bash output at 30k characters, file reads at 5MB, search at 100 results. Follow that pattern.

### No retry after a failure

If the conversation is already too big and the model says "too many tokens," the request fails. Context Guard works *before* the call to keep this from happening, but it can't fix things *after* the fact.

Automatic retry (compress and try again) may come in a future version, but it depends on a feature in ADK that hasn't been released yet.

### Token counting is approximate

Tokens are estimated as **4 characters ≈ 1 token**. Same method Google's ADK uses. Works well for English, but can be off for CJK languages (underestimates) or code (overestimates).

### Sliding window ignores message size

Sliding window counts messages, not their size. A message with 50,000 tokens counts the same as one with 10. If your conversations have wildly different message sizes, use token threshold instead.

## How it knows the model's limits

Context Guard needs each model's context window size to know when to compress. It loads this from a model catalog that refreshes every 6 hours.

If a model isn't in the catalog, it assumes 128,000 tokens (safe enough for most current models).
