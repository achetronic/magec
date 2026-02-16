---
title: "Cron"
---

Cron clients run [commands](/magec/docs/commands/) on a schedule. Define when and what, and Magec handles the rest — no external scheduler, no extra infrastructure. The command fires at the specified time, the agent processes it, and the result is logged.

This is how you automate tasks that need to happen regularly without anyone pressing a button.

## How it works

1. You create a [command](/magec/docs/commands/) — a reusable prompt paired with an agent
2. You create a cron client and select that command
3. You set the schedule (when it should fire)
4. Magec's built-in scheduler triggers the command at the specified times

The command defines **what** happens (the prompt and which agent processes it). The cron client defines **when** it happens. This separation means you can reuse the same command across multiple cron jobs with different schedules, or change the schedule without touching the command.

## Configuration

| Field | Description |
|-------|-------------|
| `name` | Display name — helps you identify this job in the Admin UI |
| `schedule` | Cron expression or shorthand (see below) |
| `commandId` | Which command to run |
| `allowedAgents` | Which agents/flows this cron job can access |

## Schedule format

Magec uses standard 5-field cron expressions:

```
┌───────────── minute (0-59)
│ ┌───────────── hour (0-23)
│ │ ┌───────────── day of month (1-31)
│ │ │ ┌───────────── month (1-12)
│ │ │ │ ┌───────────── day of week (0-6, Sunday = 0)
│ │ │ │ │
* * * * *
```

### Examples

| Expression | Runs |
|-----------|------|
| `0 9 * * *` | Every day at 9:00 AM |
| `0 9 * * 1-5` | Every weekday at 9:00 AM |
| `*/15 * * * *` | Every 15 minutes |
| `0 0 * * *` | Every day at midnight |
| `0 8,12,18 * * *` | At 8 AM, noon, and 6 PM daily |
| `0 0 1 * *` | First day of every month at midnight |
| `30 14 * * 5` | Every Friday at 2:30 PM |

### Shorthands

For common schedules, you can use these instead of writing the full expression:

| Shorthand | Equivalent | Runs |
|-----------|-----------|------|
| `@yearly` | `0 0 1 1 *` | Once a year (January 1, midnight) |
| `@monthly` | `0 0 1 * *` | Once a month (1st, midnight) |
| `@weekly` | `0 0 * * 0` | Once a week (Sunday, midnight) |
| `@daily` | `0 0 * * *` | Once a day (midnight) |
| `@hourly` | `0 * * * *` | Once an hour (on the hour) |

## Use cases

### Daily reports

*"Every morning at 8 AM, generate a summary of yesterday's metrics."*

Command prompt: *"Query the database for yesterday's key metrics. Summarize total users, revenue, and notable events. Format as a brief executive summary."*

Schedule: `0 8 * * *`

### Periodic health checks

*"Every hour, check all monitored services and report anything unusual."*

Command prompt: *"Check the status of all services. Report any that are down, degraded, or showing unusual patterns. If everything is fine, just say so."*

Schedule: `@hourly`

### Weekly digest

*"Every Monday morning, compile a digest of the past week."*

Command prompt: *"Compile a digest of the past week's activity. Include completed tasks, open issues, and upcoming deadlines."*

Schedule: `0 9 * * 1`

### Automated maintenance

*"Every night at 2 AM, clean up temporary files and optimize the database."*

Command prompt: *"Run the maintenance routine: check disk space, clean temporary files older than 7 days, and report the current system health."*

Schedule: `0 2 * * *`

{{< callout type="info" >}}
Cron clients use Magec's built-in scheduler — no need for system crontab, external schedulers, or additional containers. The scheduler polls every 30 seconds and automatically picks up changes when you modify a cron client's schedule.
{{< /callout >}}
