---
title: "Skills"
---

Skills are reusable knowledge packs that you attach to agents. Think of them as "expertise modules" — a skill bundles instructions, references, assets and scripts that teach an agent how to handle a specific domain or task.

Without skills, all of an agent's knowledge has to live in its system prompt. That works fine for simple agents, but as the prompt grows with product catalogs, coding standards, response templates and domain rules, it becomes impossible to manage. Skills break that knowledge into independent, reusable pieces that the model loads **only when it needs them**.

## Lazy loading by design

A skill is **never** inlined into the system prompt. The agent sees a short summary of every skill linked to it (its name and description) and decides on its own when to call `load_skill` to read the full instructions, or `load_skill_resource` to pull in a specific file from the skill's bundled resources. Tokens are spent only on the skill the model actually uses.

This is implemented through ADK's [skilltoolset](https://google.github.io/adk-docs/tools/built-in-tools/#skill-toolset). Each agent gets three tools when at least one skill is linked to it:

| Tool | What it does |
|------|--------------|
| `list_skills` | Returns the names and descriptions of every skill linked to this agent. |
| `load_skill` | Reads a specific skill's full instructions on demand. |
| `load_skill_resource` | Reads one file from a skill's bundled resources. |

Agents with no linked skills get **no** skill tools at all — there is nothing to load.

## On-disk format

A skill is a directory under `data/skills/{slug}/` with the following layout:

```
data/skills/{slug}/
├── SKILL.md            (required, with YAML frontmatter)
├── references/         (optional, sub-directories OK)
├── assets/             (optional)
└── scripts/            (optional)
```

`{slug}` is the skill's unique on-disk identifier and **must** match the `name` field in the SKILL.md frontmatter (Magec validates this on upload).

### `SKILL.md` example

```markdown
---
name: returns-policy
description: How to handle return requests in our store.
license: MIT
allowed-tools:
  - search_orders
  - cancel_order
---

You are responsible for return requests.

- Returns are accepted within 30 days of purchase.
- Electronics have a 15-day window.
- Opened software cannot be returned.
- Always be empathetic; offer alternatives when a return is not possible.

When the customer mentions an order number, call `search_orders`. If the
return is approved, call `cancel_order` with the result.
```

The `name` and `description` fields are required. Everything else is optional but supported (see the [Agent Skills specification](https://agentskills.io/specification#frontmatter) for the full list).

### Resources, assets, scripts

The three resource sub-directories serve different purposes:

- `references/` — documentation, examples, schemas, anything the model might need to read while following the skill.
- `assets/` — templates, prompts, configuration files used by the skill.
- `scripts/` — executable scripts the skill expects to run via tools.

Magec only enforces the directory names; the contents are up to you. Files can live at any depth — `references/api/v2/schema.json` is fine.

## Uploading a skill

In the Admin UI, go to **Skills** and click **+ Upload Skill**. The dropzone accepts:

- A standalone `SKILL.md` (with valid frontmatter) — useful for quick text-only skills.
- A `.zip` archive containing `SKILL.md` at the root plus optional resource sub-directories.
- A `.tar.gz` / `.tgz` archive with the same layout.

If the upload's frontmatter `name` matches an existing skill, the request fails with a clear conflict message. Tick **Replace** in the modal to overwrite the existing skill — its store ID is preserved so any agent linked to it stays linked.

The same single upload endpoint handles both creation and updates. There is no manual create form, no "edit" mode, and no per-file upload after the fact: skills are uploaded as packages, full stop. To change one file, edit the package on your machine and re-upload with **Replace** enabled.

## Reading a skill

Click any skill card to open its viewer. The viewer is read-only and shows:

- The frontmatter (license, compatibility, allowed-tools, custom metadata).
- The instructions body, rendered as Markdown.
- Every resource grouped by `references / assets / scripts`, with relative paths and sizes.
- A **Download package** button that streams the on-disk directory as a `.tar.gz`. Useful for backups, magec-to-magec transfers, or letting the operator review what is actually on disk.

## Linking skills to agents

After uploading a skill, enable it on each agent that should use it:

1. Open the agent in the Admin UI.
2. Expand the **Skills** section.
3. Toggle on the skills you want this agent to be able to call.
4. Save.

The agent now sees the toggled skills in `list_skills` and can `load_skill` them on demand. Other agents are unaffected — every agent has its own skill whitelist.

## Hot-reload

Re-uploading a skill (or deleting one) takes effect on the next agent message. No restart needed.

## Why packages?

Skills become valuable when:

- **You reuse the same knowledge across multiple agents.** A product-catalog skill can be shared by a support agent, a sales agent and a FAQ bot. Update the package once, all agents see the change.
- **Your agent's knowledge changes frequently.** Replacing a skill package is easier than editing a sprawling system prompt.
- **You compose agents from building blocks.** Toggle skills on and off per agent without touching system prompts.
- **You want a portable, versionable artefact.** A `.tar.gz` is something you can store in git, attach to a release, share with a colleague, or copy between magec instances.

The system prompt defines *who* the agent is. Skills define *what* it can learn on demand.

{{< callout type="info" >}}
Skills are global. Upload once, link to as many agents as you want.
{{< /callout >}}
