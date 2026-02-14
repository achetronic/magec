---
title: "Cron"
---

Scheduled agent invocations using cron expressions. Each cron client references a command (reusable prompt + agent pair) and fires on schedule.

| Field | Description |
|-------|-------------|
| `schedule` | Cron expression (e.g. `0 9 * * *` for every day at 9am) |
| `commandId` | Which command to execute |
