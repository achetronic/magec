# Entity Color Map

Each entity type in the admin UI has an assigned color for visual consistency across cards, badges, icons, chips, and any future UI element.

## Assignments

| Entity | Color | Tailwind Token | Hex (reference 400) |
|--------|-------|----------------|---------------------|
| **Agents** | Sol (amber) | `sol` | `#fbbf24` |
| **Backends** | Purple | `purple` | `#c084fc` |
| **MCP Servers** | Atlantico (cyan) | `atlantico` | `#38bcd8` |
| **Memory** | Green | `green` | `#4ade80` |
| **Clients** | Lava (red) | `lava` | `#f87171` |
| **Commands** | Indigo | `indigo` | `#818cf8` |
| **Flows** | Rose | `rose` | `#fb7185` |
| **Conversations** | Teal | `teal` | `#2dd4bf` |

## Client Type Badges

Within the Clients section, badges differentiate client types:

| Client Type | Badge Color | Description |
|-------------|-------------|-------------|
| `direct`, `telegram` | Lava (red) | Interactive clients |
| `cron`, `webhook` | Teal | Automation clients |

## Usage Pattern

For any UI element tied to an entity, use its assigned color:

```
bg:       bg-{color}-500/15
text:     text-{color}-400  (icons, labels)
          text-{color}-300  (badge text, highlights)
border:   border-{color}-500/30
Badge:    <Badge variant="{color}">
```

## Custom Palette Colors

`sol`, `atlantico`, `lava`, `piedra`, and `arena` are custom colors defined in `admin-ui/src/style.css` via `@theme`. The rest (`purple`, `green`, `teal`, `rose`, `indigo`) come from Tailwind's default palette and require no extra config.

## Badge Variants

The `Badge.vue` component supports a `variant` prop matching each entity color. All seven entity colors are registered as variants.

## Notes

- **Sol** is also used as the global accent (active tabs, primary actions). This is intentional — agents are the hero entity.
- **Piedra** (grays) and **arena** (warm grays) are reserved for neutral/structural UI — not assigned to any entity.
- When showing cross-references (e.g. "used by agents" on a backend card), use the **referenced** entity's color, not the host card's color.
- **Triggers are eliminated** — cron and webhook are now client types. No separate entity color needed.
