---
title: "Cron"
---

Cron clients run [commands](/magec/docs/commands/) on a schedule. Define a cron expression, pick a command, and Magec will execute it automatically — no external scheduler needed.

<div class="screenshots" style="margin-bottom: 2rem;">
{{< screenshot src="img/screenshots/admin-clients-cron.png" alt="Admin UI — Cron client" >}}
</div>

Set the **schedule** using a standard cron expression (e.g. `0 9 * * *` for every day at 9am) and select the **command** to run. The command defines the prompt and which agent handles it — the cron client just decides when it fires.

This is useful for things that need to happen regularly without anyone pressing a button: daily summaries, periodic health checks, scheduled reports, automated maintenance tasks.

## Schedule shorthands

Besides standard cron expressions, you can use these shorthands:

| Shorthand | Equivalent | Runs |
|-----------|-----------|------|
| `@yearly` | `0 0 1 1 *` | Once a year (Jan 1, midnight) |
| `@monthly` | `0 0 1 * *` | Once a month (1st, midnight) |
| `@weekly` | `0 0 * * 0` | Once a week (Sunday, midnight) |
| `@daily` | `0 0 * * *` | Once a day (midnight) |
| `@hourly` | `0 * * * *` | Once an hour |
