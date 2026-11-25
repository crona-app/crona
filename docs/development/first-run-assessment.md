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

Next: help accuracy, including the minimum-size onboarding finding. Export
simplification and wider layout redesign should wait for their own walkthroughs.
