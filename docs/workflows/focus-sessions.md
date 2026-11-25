---
hosted: true
title: Focus Sessions
description: Work on one issue, handle interruptions, and leave a useful session record.
order: 5.2
---

# Focus on one thing

A focus session connects a timed work block to an issue. You do not start with a blank timer; you start with a reason for the work.

## Start a session

Select an issue in Planned, Ready, or In Progress and press `f` in the TUI to open **Start Focus Session**. Before the timer begins, Crona shows the issue’s estimate, worked time, duration, and expected end time when available.

Backlog issues need an explicit status change before focus can start. In the TUI, choose **Mark Ready and start** when prompted; from the CLI, change the issue to Ready (or Planned) first. The same rule applies to manually logged sessions.

Choose the profile that fits:

- **Stopwatch** runs until you end the session.
- **Pomodoro** gives the session work and break segments.
- **Timer** gives you one no-break countdown.

The daemon owns the timer, so it keeps running if you move between the TUI, CLI, and Companion.

## Handle an interruption

If something pulls you away, stash the session instead of abandoning the context. A stash keeps the issue, elapsed segments, and timer state available for later.

When you return, resume the stash or start a fresh session. Both choices keep the earlier work visible.

## Finish with a summary

Stopping a session opens a summary prompt. Write what changed, what you learned, or what remains:

```text
Reworked the docs flow and moved installation details into the core guide.
```

Completed sessions are saved in local history. You can review or amend them later when a duration or description needs correcting.

## If the timer reaches its limit

For a Pomodoro session, Crona moves through its configured break flow. For a no-break timer, you can commit the session or add more time. Extending a Timer adds duration; it does not create Pomodoro cycles.

## Next step

Use [Exports and Reports](exports-and-reports.md) when you want to turn the day or a set of sessions into something you can read elsewhere.
