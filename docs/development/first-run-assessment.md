# First-run assessment follow-up

Triaged 2026-11-25 from `.handoff/2026-11-20-p1-first-run-focus.md` and the
current source on `fix/first-run-focus-flow`. This is an assessment backlog,
not approval to implement every proposed change. The original handoff is
git-ignored and its unchecked implementation checklist is stale.

## First pass: getting to focused work

Implemented and committed:

- Typed daemon focus-status errors and recovery targets (`3d1ea35e`).
- Readable CLI errors without transport-code prefixes (`cc312fd3`).
- Explicit TUI status recovery before starting focus (`8ebc34bc`).
- Actionable Today empty states (`3fb38c98`).
- Preservation of the recovery target in dialog state (`0362d3aa`).

The branch also includes the separate expired-Pomodoro extension fix
(`54efb76b`). The focus-workflow documentation and recovery-label cleanup
complete the pending source changes. Automated checks do not replace the
remaining isolated first-run walkthrough: onboarding, issue creation, Backlog
recovery, timer start, CLI rejection wording, and filtered empty states.

## Follow-up order

The priorities below describe impact on a new user's ability to finish a task.
Items inherited from the handoff need fresh reproduction before implementation.

| Order | Assessment | Priority | Evidence and next action | Acceptance criteria |
| --- | --- | --- | --- | --- |
| 1 | Error visibility and recovery | High | The original assessment flags errors relegated to a footer. Reproduce a failed create/save, a daemon disconnect, and a rejected action with and without a dialog open. Compare inline errors with a persistent global notice before choosing a presentation. | Users can see the failure, retain entered data, and identify a retry or correction. Errors remain readable at narrow widths. Protocol codes remain diagnostic data, not user-facing prefixes. |
| 2 | Help overlay and key-hint accuracy | High | Compare the help overlay and displayed hints with the active key registry, including empty, filtered, protected, and timer-running states. Record concrete missing or misleading hints before editing. | Every advertised key works in its displayed context. The first-run path to create, plan, start, cancel, and get help is discoverable. Unavailable actions are not advertised as usable. |
| 3 | Save and submit hints in forms | Medium | Walk through issue creation, timer setup, check-in, and representative settings forms using the keyboard. Check the smallest supported terminal size as well as a normal window. | The submit key and cancel action are visible without guessing, focus order is predictable, and validation preserves input. Hints match the actual form handlers. |
| 4 | Vocabulary consistency | Medium | Inventory visible uses of task/issue, Planned/Ready/Backlog, focus/session/timer, and Review/Momentum. Distinguish intentional domain differences from conflicting labels. | A new user can follow the same concept from Today to Issues to timer setup and help. Labels agree with documented lifecycle behavior; this is a copy pass, not a protocol rename. |
| 5 | Export wizard length | Medium | The category, report choice, scope, preset, and detail stages are in `tui/internal/tui/dialogs/controller/controller_issues_export.go` and `controller_export_openers.go`. Count interactions for daily PDF, weekly Markdown, clipboard, and scoped exports. Identify redundant steps before proposing shortcuts. | Common exports have a clear short path; advanced scope/format choices remain available. Back/cancel preserves choices, and completion identifies the output location or clipboard result. |
| 6 | Column ratios and empty-space use | Low unless content becomes inaccessible | Capture empty and populated screens at narrow, medium, and wide widths. Reproduce clipping or navigation loss separately from aesthetic imbalance. | Primary work stays readable, selection details remain associated with their source, and resize does not hide required actions. Promote any functional clipping bug to High. |
| 7 | Momentum wording | Low | Read the empty, first-event, and populated states in `tui/internal/tui/views/momentumview/`. Identify terms that cannot be understood from the screen or help. | Labels explain what is counted and over which period. Empty states explain how data appears without implying that missing history is a failure. |

## Previously reported pluralization

The session summary in `tui/internal/tui/views/sessions/log.go` already selects
`segment` for one and `segments` otherwise. Do not reopen the old “1 segments”
finding without locating another affected surface. Spot-check zero, one, and
multiple segments in the next walkthrough; treat any remaining occurrence as
a small independent copy fix.

## Assessment procedure

Use an isolated runtime with telemetry disabled; never clear the user's runtime.
For each reproduced issue, record the starting state, exact keys, expected and
actual result, terminal dimensions, and a capture when layout matters. Keep
confirmed defects separate from design preferences. Implement each coherent
pass separately, with focused checks and a live keyboard walkthrough.

## Task 1 outcome (2026-11-25)

Implemented persistent, dismissible shell error notices, inline correlated save
errors, retained form inputs, and duplicate-submit protection. Background failures
no longer replace a form's validation. Late results preserve the currently open
dialog. RPC codes remain in diagnostics; connection failures explain recovery and
lost replies warn that a change may already have been saved.

Verification: unit tests, lint, focused model/layout/dialog-runtime/helper tests,
and narrow/wide rendered frames. An isolated TUI/daemon walkthrough completed
onboarding, entered an issue and new repo/stream, then temporarily moved only the
test socket. The save failed inline with the values retained. Escape dismissed
the separate background notice without closing the form. After restoring the
socket, Ctrl+S saved the same issue; Today showed one Backlog item. Both test
processes were stopped. No live user data was changed.

The walkthrough also exposed an existing separate issue: at 60x20, onboarding
content extends beyond the viewport. Track this as a High-priority first-run
layout finding; reproduce all onboarding pages at the minimum supported size
before redesigning that screen. It is outside the error-notice change.

## Task 2 outcome (2026-11-25)

At 60x20, an isolated fresh run starts on Welcome. Press Right three times to
reach Workspace, Privacy, and Review. All four pages should fit within the
frame. Before the fix, every page clipped its shared heading and progress row.
Privacy also clipped its choices and lost the key legend. Review lost its
completion choices and key legend. The regression test renders all four
pages through the shell at 60x20. The onboarding renderer now wraps copy to the
available width and uses the compact logo and spacing in short terminals.

On Today with no visible issues, `?` previously listed detail and status
actions that need a selection. It now leads with creation, quick timer when
idle, and section navigation. A filter that leaves no results retains the
clear-filter hint. The Work lists hide issue or habit actions when their
selection is empty. Protected Away uses the visible view's help and hides
bulk capture. While a timer runs or protection is enabled, help lists the
tabs reachable by their number keys. The active-session hints no longer
advertise status changes during a session, and a paused stopwatch shows Resume.

The isolated keyboard walkthrough used telemetry disabled and captured each
onboarding screen at 60x20. From Today, `?` opened help with creation and quick
timer hints above the fold. Escape closed it, `a` opened New Issue, Escape
cancelled, `A` opened Start Timer, and Escape returned to the empty Today view.
The walkthrough stopped its TUI and daemon. Automated coverage checks filtered,
protected, timer-running, and paused help states.

## Task 3 outcome (2026-11-25)

Verified issue creation, timer setup, check-in, Privacy & Diagnostics, and
Start of Day at 60x20 and 120x36 with isolated runtime data and telemetry
disabled. The issue form now advertises reverse tabbing and Enter to select
a repo or stream. Short terminals keep the filter and selected option visible.
Check-in shows reverse tabbing. Countdown labels Escape as Back, and Pomodoro
shows Tab and Shift+Tab for field navigation and Ctrl+S to start or extend.
Start of Day shows the reset shortcut on weekday rows and keeps the discard
prompt visible on short terminals. The privacy wizard names the selected
save or cancel action.

In both walkthroughs, Ctrl+S flagged an invalid issue estimate and invalid
check-in energy. Correcting those fields saved the issue and persisted mood 4
and energy 3. Escape returned from Countdown to the timer picker. Ctrl+S
started Countdown. Enter saved the privacy choices from Review. In Start of
Day, Escape opened the discard prompt, a second Escape kept the edit, and
Ctrl+S saved 10:00. Both test runtimes were stopped.

Follow-up finding: pressing `w` after saving a check-in reopened a blank
New Check-In form during both walkthroughs. The saved values were present in
the local database. Check-in reload behavior needs a separate investigation.

## Task 4 outcome (2026-11-25)

Today uses **Issues** for the same records shown under **Work → Issues**.
The pane, create and bulk-add hints, selection preview, and empty states use
that terminology. The empty plan identifies Backlog as a status and points
to Work to plan issues.

The planning guide distinguishes Today's **Planned** list from the
**Planned** lifecycle status. The list reflects the selected date and
includes open pinned issues. **Ready** remains a lifecycle status.
**Review** summarizes work by date, and **Momentum** tracks progress toward
goals. Both keep their names.

Help uses **start focus session** for `f`, and the picker opens as **Start
Focus Session** with the issue title. `A` still opens the quick timer picker.
The focus guide names all three timer profiles.

An isolated 60x20 walkthrough created an issue for the current date, opened
keyboard help, checked the `f` and `A` pickers, and found the same issue in
Work. Both test processes were stopped.

## Task 5 outcome (2026-11-25)

Counted Enter confirmations after opening Export Report. Daily PDF, weekly
Markdown, and daily clipboard previously required three confirmations
(category, report, style). Each now takes one after highlighting the output.
The active template handles the default export. Press `s` to choose another
style. With PDF available and the cursor starting at the first row, total
keypresses after opening Export fell from 4 to 1 for daily PDF, 8 to 5 for
weekly Markdown, and 7 to 4 for the daily clipboard. Repo, stream, and issue
reports retain three confirmations for category, report, and scope. The repo
path takes four extra Down presses to reach Project Reports. Summary and
Data Exports remain on the first menu.

Escape returns to the prior selection. Tests cover reopening a repo scope
and returning from Summary style and output to the entered custom range.
File exports show the full path in a completion dialog. Clipboard exports
show a status confirmation.

An isolated 60x20 walkthrough generated a daily PDF, weekly Markdown, and a
repo-scoped Markdown file. It also copied a daily report to an isolated
clipboard substitute and verified that Back restored the selected weekly
report after visiting its styles. Verified the files and stopped both test
processes.

## Task 6 outcome (2026-11-25)

Captured empty Today, Today with a dated issue, and focused issue details at
60x20, 120x36, and 200x46 in one isolated session. Before the fix, the compact
week strip cut off Sunday's date at 60 columns, and the selected issue's
focus hint fell off the end of its action row. The week now uses shorter
day labels when needed and saves the two blank rows on short screens. The
issue action row keeps Focus next to Details.

The renderer tests also reproduced the fixed split giving empty Habits six
rows while 30 issues competed for seven at minimum height. When one list is
empty and the other needs more space, Today reserves the empty list's heading
and guidance and gives the remaining rows to the populated list. The wide
column split keeps the calendar, issue preview, and day summary readable.
During the keyboard walkthrough, Enter opened the selected issue, resizing
kept the same issue title in the detail pane, and Escape returned to Issues.

## Task 7 outcome (2026-11-25)

Momentum names the unit it counts and the period it uses, such as focus
time per week across contexts or completions per day for each habit. The
chart's arrow hint names the selected day, week, or month. The empty view
explains what a goal tracks, and filtered results tell the user how to
clear the filter. A goal with no counted activity in the window explains
what to log. The status and chart distinguish that state from past periods
that missed a target after activity was recorded.

The isolated session captured empty Momentum at all three sizes. Renderer
tests cover zero-count history, the first matching focus session, and
populated habit and focus goals. Both test processes were stopped.
