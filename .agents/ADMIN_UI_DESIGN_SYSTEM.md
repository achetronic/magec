# Magec Admin UI — Design System

> Living document. Update when patterns change.

## Principles

1. **Consistency over novelty** — reuse existing patterns, never invent new control types
2. **Quiet by default** — muted colors (`arena-500/600`), reveal on hover
3. **Color = meaning** — each entity has ONE assigned color, used everywhere
4. **Controls match their purpose** — segmented controls for view modes, icon buttons for actions, selects for filters

---

## Entity Colors

| Entity | Color | Icon |
|--------|-------|------|
| Backends | `purple` | `server` |
| Memory | `green` | `database` |
| MCP Servers | `atlantico` | `bolt` |
| Agents | `sol` | `users` |
| Flows | `rose` | `flow` |
| Commands | `indigo` | `command` |
| Clients | `lava` | `phone` |
| Conversations | `teal` | `chat` |

Tinted backgrounds always `{color}-500/10` or `{color}-500/15`, text `{color}-300` or `{color}-400`.

---

## Card Component (`Card.vue`)

**Props**: `active` (Boolean), `color` (String — entity color key)

Base classes: `bg-piedra-900 border border-piedra-700/50 rounded-xl p-4 transition-all duration-200`

**Hover behavior**: When `color` is set, hover tints the border to the entity color and adds a subtle glow shadow. Without `color`, hover lightens to `piedra-600/50`.

| Property | Value | Rationale |
|----------|-------|----------|
| Border opacity | `{color}-500/15` | Barely visible tint — quiet by default, color on interaction |
| Shadow | `0_0_15px_-3px_rgba({r},{g},{b},0.04)` | Imperceptible glow, just enough to lift the card |
| Fallback (no color) | `hover:border-piedra-600/50` | Neutral lightening for generic cards |

**DRY rule**: All entity list views use `<Card :color="entityColor">`. No view should duplicate hover border/shadow classes. Specialized cards (like `MemoryCard`) should wrap `Card` as their outer container rather than reimplementing the same `<div>` with duplicated styles.

**Color map** (8 keys): `purple`, `green`, `atlantico`, `sol`, `rose`, `indigo`, `lava`, `teal` — matches the Entity Colors table above.

---

## Control Taxonomy

### 1. Segmented Control (view mode switches)

For toggling between **views of the same data** (Messages/Raw, User/Admin perspective, Session/Long-term).

**Container**: `flex items-center gap-1 p-0.5 rounded-lg bg-piedra-800 [border border-piedra-700/50 if standalone]`
**Segments**: `px-3 py-1.5 text-xs font-medium rounded-md transition-colors cursor-pointer`
- Active: `bg-piedra-700 text-arena-100` (neutral) or `bg-{color}-500/20 text-{color}-300` (color-coded)
- Inactive: `text-arena-500 hover:text-arena-300`
- Disabled: `text-arena-600 cursor-not-allowed`

### 2. Icon Button (actions)

For **actions** that don't change view mode (refresh, delete, edit, settings).

`p-1.5 rounded-lg transition-colors group/btn`
- Default: icon `text-arena-500`, hover bg `hover:bg-piedra-800`
- Destructive: hover bg `hover:bg-lava-500/10`, icon `group-hover/btn:text-lava-400`
- Active state (toggle on): `bg-{color}-500/10`, icon `text-{color}-400`

### 3. Primary Action Button (CTA)

For **creating** things. Always rightmost in header.

`px-3 py-1.5 bg-sol-500 hover:bg-sol-600 text-piedra-950 text-xs font-medium rounded-lg transition-colors`

### 4. Filter Select

For **filtering lists**. Placed in a filter bar below header.

`bg-piedra-800 border border-piedra-700/50 text-arena-200 text-xs rounded-lg px-2.5 py-1.5 outline-none focus:border-piedra-600`

### 5. Filter Pills (toggleable tags)

For **multi-select filters** like agent tags.

`px-2.5 py-1 text-[11px] font-medium rounded-lg border transition-all cursor-pointer`
- Selected: `bg-{color}-500/15 text-{color}-300 border-{color}-500/30`
- Unselected: `bg-piedra-800 text-arena-500 border-piedra-700/40 hover:border-piedra-600`

### 6. Icon Button with Label

For **destructive actions that need clarity** (e.g. "clear all"). Icon + short text, same muted style as plain icon buttons.

`flex items-center gap-1 p-1.5 hover:bg-piedra-800 rounded-lg transition-colors group/btn`
- Icon + text both `text-arena-500 group-hover/btn:text-arena-300`
- Text: `text-[10px] font-medium`
- Never use colored/red backgrounds for destructive actions in headers — red steals attention. Keep muted; the confirm dialog provides the safety net.

---

## Header Layouts

### Standard List (Agents, Flows, Memory, etc.)
```
[h2 title] ——————————————————————— [+ New CTA]
```

### Conversations List (no create, has refresh + auto-refresh + destructive)
```
[h2 title] ——— [segmented: Off | 5s | 30s]   [icon: ↻]  [icon+label: 🗑 All]
                └────── segmented ──────┘      └────── icon buttons ──────────┘
                                         gap-3
```
- Auto-refresh is a **segmented control** (view mode of the polling behavior), not an icon toggle.
- Manual refresh is an icon button. On auto-refresh tick, the icon briefly spins 180° and highlights (`text-arena-200 rotate-180`) for 400ms as visual feedback.
- "Clear all" uses icon+label pattern (`🗑 All`) — a bare trash icon is ambiguous ("delete what?").

### Detail View (back navigation)
```
[back ◁] [title / badges / meta] ——— [Off|5s|30s] [↻] | [User|Admin] [Messages|Raw] [✕ Session] [🗑]
                                      └─ refresh ──────┘   └─ view toggles + actions ──────────────┘
                                                     divider (w-px h-4 bg-piedra-700/50)
```

- Auto-refresh + manual refresh grouped on the left of controls, same pattern as list views.
- A thin vertical divider (`w-px h-4 bg-piedra-700/50`) separates refresh controls from view/action controls.
- Auto-refresh resets to `Off` when navigating to a different item.
- Timer cleaned up in `onBeforeUnmount`.

Segmented controls and icon buttons are visually distinct groups separated by `gap-3`.

---

## Spacing

| Context | Value | Why |
|---------|-------|-----|
| Page sections | `space-y-4` | Clear separation between major blocks |
| Card grid | `gap-3`, `grid-cols-1 sm:grid-cols-2` | Balanced density |
| Card padding | `p-4` | Enough room for 3–4 content lines |
| Card / header internal lines | `space-y-2` | Each line (title, badges, meta) needs breathing room — tighter spacing makes them compete |
| Control groups (same type) | `gap-1.5` | Buttons that belong together |
| Between control types | `gap-3` | Visual separator between segmented controls and icon buttons |
| List item rows | `py-3 px-3` | Hover highlight needs generous vertical padding to feel clickable, not cramped |
| Inline content blocks | `py-4` on container | Breathing room around a scrollable content area (message thread, log list, etc.) |

---

## Detail Header Anatomy

Two lines inside `space-y-1.5`. Title + info popover on line 1; categorical badges on line 2. Text metadata (time, IDs, counts) lives in a hover popover to keep the header compact.

```
[back ◁]  Software Factory  [summarized] [👁]      [Off|5s|30s] [↻] | [User|Admin] [Messages|Raw] [⬇PDF] [✕ Session] [🗑]
          [Direct] [Flow] [VoiceUI]
```

On `< lg` screens, controls drop below the title row (`flex-col lg:flex-row`).

| Line | Content | Style |
|------|---------|-------|
| 1. Title | Entity name + status badge + info popover icon | `text-sm font-semibold text-arena-200` |
| 2. Tags | Categorical badges | `Badge variant="muted"` with `!py-0`, `gap-1.5` |

### Info popover

An `eye` icon on the title line. On hover, shows a floating panel (`bg-piedra-900 border border-piedra-700/50 rounded-lg shadow-xl`) with key-value rows:
- Started (timestamp)
- User (userId)
- Session (full sessionId, monospace)
- Messages (count)

Icon style: `text-arena-600`, hover `text-arena-400`. Panel: `min-w-52`, `z-50`, positioned `top-full mt-1.5`.

### Detail header rules

- **Never place colored badges on the title line** (except status badges like `summarized`) — they fight for attention with the title.
- **Text metadata belongs in a popover**, not inline — it clutters the header and competes with controls for horizontal space.
- **Responsive breakpoint at `lg`** (1024px): controls drop below the title row. Above `lg`, everything fits in a single row.
- **Consistency with cards**: detail headers follow the same badge pattern as list cards (muted variant, `!py-0`).

---

## Conversation Card Anatomy

Four distinct lines, each with a single purpose. Internal spacing `space-y-2`.

```
┌─────────────────────────────────────────────────────────┐
│  [icon]   Software Factory                              │
│           "hola, quiero que cada uno de los agentes..."  │
│           [Direct] [Flow] [VoiceUI]                     │
│           5m ago                                        │
└─────────────────────────────────────────────────────────┘
```

| Line | Content | Style |
|------|---------|-------|
| 1. Title | Agent/flow name + optional `summarized` badge | `text-sm font-medium text-arena-100` |
| 2. Preview | First user message, quoted and italic | `text-[11px] text-arena-500 italic` — `"text here"` |
| 3. Tags | Categorical badges: source, flow type, client name | `Badge variant="muted"` with `!py-0`, `gap-1.5` |
| 4. Timestamp | Relative time only | `text-[10px] text-arena-600 tabular-nums` |

### Card design rules

- **Never mix badges with text metadata** on the same line — badges are categories, text is temporal/contextual.
- **Preview text** uses quotes and italic to convey "someone said this" without needing a label.
- **Source badge** shows the client type capitalized (`Direct`, `Voice UI`, `Telegram`, `Cron`, `Webhook`) — the icon on the left also encodes this visually but the badge adds textual clarity.
- **Flow badge** says `Flow` (the type), never the flow name — the name is already the card title.
- **Client name badge** shows the client's configured name (e.g. `VoiceUI`). Only present when the request was authenticated with a client token.
- **userId removed** from card — visible in detail view, too much info density for a list card.

---

## List Item / Row Layout

For any repeating row in a scrollable area (messages, logs, events, audit entries).

| Property | Value | Why |
|----------|-------|-----|
| Row padding | `py-3 px-3` | Hover highlight looks generous, not cramped |
| Negative margin | `-mx-3` | Row highlight bleeds to container edge |
| Row gap (between icon and content) | `gap-3` | Enough room for avatar/icon + text |
| Hover | `hover:bg-piedra-800/30 rounded-lg` | Subtle, no border shift |
| Inter-row spacing | `space-y-1` on container | Rows sit close but padding gives each one room |

### Row content hierarchy

- **Primary label** (author, event type): smallest readable size (`text-[10px]`), muted (`text-arena-500`). If an avatar/icon already identifies the entity, the label is secondary — keep it quiet.
- **Body content** (message text, log details): `text-[13px] text-arena-300 leading-[1.7]`. This is what the user came to read — it must dominate visually.
- **Hover-only metadata** (timestamps): `opacity-0 group-hover:opacity-100` — available on demand, invisible by default.

> **Rule**: If a visual indicator (avatar, icon, color) already communicates the entity, the text label must not compete with the body content. Drop font-weight, drop color coding, shrink size.

---

## Badge Variants

The `Badge` component has a `variant` prop. Use **`muted`** for all informational/categorical chips on cards. Reserve colored variants only for **status indicators**.

| Variant | Style | When to use |
|---------|-------|-------------|
| `muted` | `bg-piedra-800 text-arena-500` | Tags, categories, entity references, type labels — anything descriptive/informational |
| `green` | `bg-green-500/15 text-green-300` | Status only: `summarized`, `healthy`, `active` |
| `default` | `bg-piedra-800 text-arena-300` | Avoid — slightly brighter than `muted`, creates visual competition with titles |
| Colored (`sol`, `atlantico`, `rose`, etc.) | `bg-{color}-500/15 text-{color}-300` | **Never on cards** — colored badges compete with card titles and create noise. Only acceptable in isolated contexts (e.g., filter pills) |

### Rule: badges on cards are always `muted`

Colored badges draw the eye away from the card title and create visual hierarchy conflicts. On list cards, all badges (backend/model, STT, TTS, MCP count, agent references, type labels) use `variant="muted"`. The card's hover border already provides the entity color — badges don't need to repeat it.

**Exception**: status badges like `summarized` use `variant="green"` because they convey a state change, not a category.

---

## Typography Scale

| Role | Classes |
|------|---------|
| Page title | `text-sm font-semibold text-arena-200` |
| Card title | `text-sm font-medium text-arena-100` |
| Body | `text-xs text-arena-400` |
| Meta / hint | `text-[10px] text-arena-500` |
| Badge | `text-[10px] font-medium` |
| Section label | `text-[10px] font-medium text-arena-500 uppercase tracking-wider` |

---

## Decisions Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-02-15 | Segmented controls for view modes, icon buttons for actions | Prevents mixing control types that look similar but do different things |
| 2026-02-15 | Auto-refresh as segmented control (Off/5s/30s), not icon toggle | Two identical refresh icons side-by-side is confusing; segmented control clearly shows the current polling state |
| 2026-02-15 | Auto-refresh tick pulses the refresh icon (spin 180° + highlight 400ms) | Visual feedback that something happened, without being intrusive |
| 2026-02-15 | "Clear all" as icon+label (`🗑 All`), not red button or bare trash icon | Red steals attention in a muted UI; bare trash icon is ambiguous ("delete one? all?"); icon+label is clear and quiet |
| 2026-02-15 | Conversation messages: linear thread layout, no chat bubbles | Bubbles with colored borders are visually heavy; thread layout (left-aligned, no bg) is more readable for audit logs |
| 2026-02-15 | ~~Perspective toggle uses colored segments (teal/sol)~~ → neutral gray | Colored active states created visual noise; all segmented controls should look uniform |
| 2026-02-15 | `dual` badge removed from conversation list | Cryptic, no user value — perspective switching lives in detail view |
| 2026-02-15 | Never mix badges and text metadata on the same line | Badges are categories, text is temporal — mixing them creates visual noise with no hierarchy |
| 2026-02-15 | Preview text in italic with quotes | Conveys "someone said this" without a label; `text-arena-500` keeps it subordinate to the title |
| 2026-02-15 | Flow badge says "Flow" not the flow name | The card title already shows the name; repeating it in a badge is redundant |
| 2026-02-15 | Source badge always present, capitalized | Even with a client name badge, source type (Direct/Cron/Webhook) communicates *how* the conversation was triggered |
| 2026-02-15 | userId removed from conversation cards | Too much density for a list; available in detail view where there's room |
| 2026-02-15 | Card lines spaced with `space-y-2` | Prevents information from feeling crammed together |
| 2026-02-15 | Detail headers: 2-line layout (title+info / badges) with `space-y-1.5` | Meta moved to popover; two lines keep the header compact and leave room for controls |
| 2026-02-15 | Badges never on the title line (except status like `summarized`) | Colored/muted badges next to a title fight for attention; title must stand alone as the primary anchor |
| 2026-02-15 | Row labels (author, event type) subordinate to body content | When an avatar/icon already identifies the entity, the text label must be quiet: small, muted, no font-weight — body text is what the user came to read |
| 2026-02-15 | List rows use `py-3 px-3` padding | Hover highlight needs generous vertical padding; cramped rows feel unclickable and make the UI look dense |
| 2026-02-15 | Reset session uses `close` icon + "Session" label, not `refresh` icon | `refresh` means "reload"; destructive actions need an icon that conveys removal + a label for clarity — same pattern as "Clear All" |
| 2026-02-15 | Detail views get the same auto-refresh pattern as list views | Users need live-updating data in detail views too; reusing the exact same control (segmented Off/5s/30s + pulse icon) keeps the UI learnable |
| 2026-02-15 | All segmented controls use neutral gray active state (`bg-piedra-700 text-arena-100`) | Colored active states (teal/sol) created visual noise; segmented controls are about *what* you're viewing, not *what kind* — keep them uniform |
| 2026-02-15 | PDF export uses icon+label (`download` + "PDF") pattern | Clear intent without ambiguity; same pattern as "Clear All" and "✕ Session" — icon alone could mean multiple things |
| 2026-02-15 | Detail header meta (time, user, session, msg count) moved to hover popover | Inline meta text consumed too much horizontal space, forcing controls to overlap or wrap too early; popover keeps data accessible without cluttering the header |
| 2026-02-15 | Detail header responsive: `flex-col lg:flex-row` | Controls bar is wide; below `lg` it drops to its own line cleanly instead of overlapping the title |
| 2026-02-15 | Messages displayed newest-first, "load older" button at the bottom | Most recent activity is what the user cares about first; older messages loaded on demand downward |
| 2026-02-15 | Card.vue `color` prop with `colorMap` for entity-colored hover borders | DRY: all 7 list views pass their entity color to Card instead of duplicating `hover:border-{color}` classes |
| 2026-02-15 | Card hover intensity: border `/15`, shadow `0.04` | First version (`/30`, `0.08`) was too intense — halved both values for a subtle, barely-perceptible effect |
| 2026-02-15 | MemoryCard should wrap Card, not duplicate its styles | Pending refactor — MemoryCard currently has its own outer `<div>` with duplicated hover classes; should use `<Card color="green">` as container |
| 2026-02-15 | All card badges use `variant="muted"` — no colored badges on cards | Colored badges (`sol`, `atlantico`, `rose`) competed visually with card titles; `muted` keeps badges subordinate. Only exception: status badges like `summarized` (green) |
