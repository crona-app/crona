---
hosted: true
title: Issues and Planning
description: Turn a backlog into a day you can see and act on.
order: 5.1
---

# Plan the workday

Crona uses issues as the bridge between intention and focused work. Capture the work first; decide when to do it without losing the original context.

## Capture an issue

Give an issue a concrete title, then add an estimate, notes, repository, and stream when they are useful. An issue should describe a move you can finish, not an entire project.

Good: `replace the docs install instructions`.

Less useful: `improve documentation`.

## Put work on the day

Set an issue's due date to put it on that date in Today. Assigning a date to a **Backlog** issue moves it to **Planned**. In the TUI, press `2` to open Work, select the issue, and press `d` to set the date.

In Today's Issues pane, the **Planned** list shows issues dated for the selected day, with open pinned issues at the top. The list can include issues in other statuses, such as **Ready** and **In Progress**. **Overdue** and **Upcoming** show issues with earlier and later dates.

## Use the issue lifecycle

Statuses tell you where an issue is in its journey:

`Backlog → Planned → Ready → In Progress → In Review → Done`

Use **Blocked** when something outside the issue is holding it up. Use **Abandoned** when you have deliberately decided not to continue.

Starting a focus session promotes a Planned or Ready issue to **In Progress**. The daemon validates transitions so different Crona clients do not silently disagree.

## Set the active context

Check out a repository, stream, or issue when you want commands and reports to default to it. The active context is shared by the TUI and CLI through the local daemon.

It is a pointer for the work you are doing now. It is not a replacement for planning dates or issue history.

## Next step

Select an issue and [start a focus session](focus-sessions.md). When the work is done, Crona will ask you to leave a short summary.
