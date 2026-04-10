# Release Notes Template

## Format

```markdown
## vX.Y.Z

### 🐛 Bug fixes

- **Short title describing the fix** — What was broken, why it was broken, what was changed to fix it.

### ✨ Features

- **Short title describing the feature** — What it does and why it's useful.

### 🔧 Improvements

- **Short title describing the improvement** — What changed and what's better now.

### 💥 Breaking changes

- **Short title** — What changed, what breaks, and how to migrate.
```

## Rules

- **English only**, plain language — no jargon, no internal references
- **Only include sections that have entries** — omit empty sections entirely
- **One bullet per change** — if a fix touches multiple files, it's still one bullet
- **Bullet structure**: `**Title** — cause, effect, fix` (for bugs) or `**Title** — what it does` (for features)
- **No emojis inside bullets** — only on section headings
- **Version follows semver**: `MAJOR.MINOR.PATCH`
  - `PATCH` — bug fixes only
  - `MINOR` — new features, backwards compatible
  - `MAJOR` — breaking changes

## Section icons

| Section          | Icon |
| ---------------- | ---- |
| Bug fixes        | 🐛   |
| Features         | ✨   |
| Improvements     | 🔧   |
| Breaking changes | 💥   |

## Example

```markdown
## v0.15.1

### 🐛 Bug fixes

- **Discord: thread messages ignored when `allowedChannels` is set** — Messages inside a thread were silently dropped even when the thread was created from an allowed channel. Discord threads have their own ID, different from the parent channel, so they were failing the permission check. The bot now also checks the parent channel ID when deciding whether to allow a message.

- **Discord/Slack: `threadHistoryLimit` could not be updated from the Admin UI** — When saving a client from the edit form, the value was sent as a string instead of a number. The backend rejected the request with a deserialization error. The form now correctly sends integer and number fields as numbers.
```
