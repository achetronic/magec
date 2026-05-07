# Documentation writing style

How Magec's docs are written. Read this before touching `website/content/docs/`. The goal is docs that read like a human wrote them, not like an AI vomited them.

## Audience and tone

Write for a competent reader who doesn't yet know the feature. They've already done getting-started and built a basic flow or agent. They're now learning something new.

Tone is direct and matter-of-fact. Second person ("you set this", "you'll see"), present tense. No marketing voice, no "powerful and flexible", no "robust and scalable". Show what the feature does and how to use it; let the reader decide if it's powerful.

## Length and density

Short over long. If a section is over 200 words, ask whether half of it is filler. Repetition for emphasis is filler. Restating the same idea in different words is filler. A second example that doesn't add a new dimension is filler.

A docs page should feel scannable. Headings every few paragraphs, code blocks where they help, lists where the content is genuinely a list. A reader who skims should still come away with the right mental model.

## Anti-AI tics

These are the patterns that make text read as AI-generated. Avoid them.

### Em-dash as a connector

Don't use ` — ` to glue two clauses inside a sentence. It's the single biggest tell.

- Bad: `A flow runs each step in order — that's enough for many cases — but not for these:`
- Good: `A flow runs each step in order. That's enough for many cases, but not for these:`

Em-dashes are fine when they're parenthetical (`the cap (always active) is a safety net`) or when truly setting off an aside, but the moment you find yourself writing two of them in three paragraphs, replace them with full stops or commas.

### "X — Y" construct

Bad: `Iterative refinement — Loops let agents revise their work.`
Good: `Iterative refinement. Loops let agents revise their work.`

When listing items, end each bullet with a full stop, not a dash continuing into a definition.

### "Two/three things:" + bullet list of nouns

When you see yourself about to write `Two consequences:` followed by a bullet list, stop and think whether the items are really independent (then keep the list) or just a single thought you're padding (then write a sentence).

Bad:

```
The behaviour you can rely on:
- The second message of a conversation sees state written by the first.
- Two different conversations on the same flow do not share state.
- State never crosses between flows.
```

Good if the items are genuinely independent things to remember (often, they are). Keep them as a bullet list. The mistake is using the construct when the content is one paragraph.

### Restating the obvious

Bad: `The expression must return a boolean. If it doesn't return a boolean, it's rejected.`
Good: `The expression must return a boolean.`

Bad: `Click the button. The dialog opens. Now you can configure the loop.`
Good: `Click the button to open the loop config dialog.`

### Excessive enumeration

Don't enumerate every operator, every helper, every supported value when a sentence will do. Refer to the canonical source instead.

Bad: ``Operators: `==`, `!=`, `<`, `>`, `<=`, `>=`, `&&`, `||`, `!`. Helpers: `has()`, `size()`, `in`, string `.contains()`. The expression must return a boolean.``

Good: `Expressions follow the [CEL](https://github.com/google/cel-spec) syntax. As long as your expression returns a boolean, you can write it however you want.`

The reader who needs the operator list goes to the linked spec. Everyone else doesn't.

### "Important consequences:" / "Note that:" preambles

Bad: `Important consequences: A is X. B is Y.`
Good: `A is X. B is Y.`

If a sentence is important, just say it. The framing word adds nothing.

### Smart quotes and typography flourishes

Plain ASCII. Straight quotes. No "smart" curly quotes, no en-dashes for ranges where a regular hyphen does, no bullet characters as text.

## When to use what

| Format | Use when |
|--------|----------|
| Paragraph | You're explaining a concept, giving context, or reasoning about behaviour |
| Bullet list | You have 2+ independent items of similar weight (no internal logical order) |
| Numbered list | The order matters (steps, sequential events) |
| Table | You're comparing 3+ things along the same dimensions, or mapping inputs to outputs |
| Code block | Anything the reader might paste, copy, or run verbatim (prompts, config, expressions, commands) |
| Callout | Side notes that the main flow shouldn't carry but the reader needs to know |

If a list has only two items, consider writing it as a sentence with "and" between them.

## Code and examples

Examples should be copy-pasteable and complete. If you write a system prompt, write the whole prompt. If you write a config, write enough that the reader can plug it in.

Show realistic content. Avoid placeholder names like `foo`, `bar`, `myVar`. Use names that suggest the role: `Generator`, `Critic`, `state.approved`, `research_results`.

Annotate trees and pseudo-output blocks with comments where it helps:

```
Loop
└── Sequential
    ├── Generator
    └── Critic   ← marked as response agent
```

## Headings and structure

Use sentence case for headings (`How data flows between agents`), not title case (`How Data Flows Between Agents`). Magec's docs are consistent on this.

A page typically has:

1. A 1-2 paragraph intro that says what the page covers and links to prerequisites.
2. The "why this exists" section, with concrete cases the feature addresses.
3. The main content, organised by user task.
4. A summary or troubleshooting section at the end if relevant.

Don't write a "Conclusion" section. Stop when you've said what you need to say.

## Cross-linking

Link to other docs pages with relative URLs (`/docs/agents/`, `/docs/flows/`). Don't repeat content from those pages, link to it. If a concept is already covered elsewhere, a single sentence with a link is enough.

External links: use Markdown link syntax with descriptive text (`[CEL playground](https://playcel.undistro.io/)`), not raw URLs.

## Screenshots

Use the `{{< screenshot >}}` shortcode. Place files under `website/themes/magec/static/img/screenshots/` with a descriptive name (`admin-flow-loop-mode-dialog.png`), not the filename your screenshot tool gave you (no `Screenshot 2026-XX-XX.png`).

A docs page can include placeholder screenshot shortcodes pointing to files that don't exist yet — Hugo will render the alt text. Track the missing screenshots in a TODO so they get captured before release.

## Callouts

Two types:

```
{{< callout type="info" >}}
Side information the reader benefits from but isn't critical.
{{< /callout >}}

{{< callout type="warning" >}}
Something that will trip the reader if they're not careful.
{{< /callout >}}
```

`type="warning"` is the default. Use it sparingly. If half the page is callouts, the page is structured wrong.

## Tables

Tables are for comparison. If a table has only one column of "values" and the rest is description, it's probably a list disguised as a table.

Keep cells short. If a cell becomes a paragraph, that content belongs in the body and the table should reference it.

## Voice and verbs

- Active voice over passive: "Magec rejects malformed expressions" not "Malformed expressions are rejected by Magec".
- Concrete verbs over abstract: "the loop stops" not "termination occurs".
- Direct over hedged: "use this when X" not "you might want to consider using this when X could be the case".

## Things to never do

- Use the phrase "leverage" for "use".
- Write "It's important to note that..." (delete those words, the rest of the sentence is fine on its own).
- Write "In a nutshell" or "At its core". The reader is here to learn the thing, not the meta-summary.
- End a section with a recap of the same section.
- Open a section with "In this section, we will...".
- Use "we" or "let's" except when literally walking through code together (rare in reference docs).
- Add emojis to body text (✅/❌ in tables for the "yes/no" axis is fine and consistent with the project).

## Examples of before/after

### Example 1

**Before** (AI tics):

> Flows are powerful, flexible workflows that allow you to compose multiple agents — each agent specialises in a particular task — into a coordinated multi-step pipeline. By leveraging flows, you can achieve quality through specialisation, parallel processing, and iterative refinement. It's important to note that flows are configured visually in the Admin UI.

**After**:

> A flow chains multiple agents into a multi-step workflow. Instead of one agent handling everything, you split the work: one agent researches, another writes, another reviews. Each agent focuses on what it does best, and the flow coordinates them. You build flows visually in the Admin UI.

### Example 2

**Before** (over-enumerated):

> The CEL expression supports the following operators: `==`, `!=`, `<`, `>`, `<=`, `>=`, `&&`, `||`, `!`. It also includes these built-in helpers: `has()`, `size()`, the `in` keyword for membership tests, and `.contains()` for string operations. Note that all expressions must return a boolean value to be valid.

**After**:

> Expressions follow the [CEL](https://github.com/google/cel-spec) syntax. As long as your expression returns a boolean, you can write it however you want.

### Example 3

**Before** (em-dash overload):

> The shared state — accessible to all agents in the flow — persists for the whole conversation — meaning that subsequent messages can read what previous turns wrote.

**After**:

> Shared state is visible to every agent in the flow and persists for the whole conversation. The next message in the same conversation still sees what the previous turn wrote.

## Reviewing before merge

Before publishing a docs page, run through this:

1. Read the page top to bottom out loud. Anywhere you stumble, rewrite.
2. Count em-dashes. If there are more than two on the page and they're not parenthetical, reduce them.
3. Look for any sentence with "—" that could be a full stop. Replace it.
4. Find every list and ask: are these really independent items, or is this one thought I padded out?
5. Find every "important note that" / "in this section" / "as we mentioned earlier" and delete the phrase.
6. Skim the page like a busy reader. Can you get the gist from headings + first sentences only? If not, restructure.
