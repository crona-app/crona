# Focus Recovery and Check-In Implementation Plan

> **For agentic workers:** Use the Superpowers executing-plans skill to work through the steps. Verify each regression before changing production code.

**Goal:** Preserve open forms when focus recovery finishes late and reopen saved check-ins with their persisted values.

**Architecture:** Route focus recovery through the existing correlated form-submission lifecycle. Return a typed check-in save result to the model and apply it to the matching date before refreshing dependent views.

**Tech stack:** Go, Bubble Tea, Crona local IPC.

**Evidence:** The release audit reproduced the delayed focus reply with a model test. The first-run assessment records the check-in walkthrough and persisted database values.

## Constraints

- Keep the existing IPC methods, payloads, and issue lifecycle rules.
- Escape can close a pending form while its request finishes. A successful status change remains saved.
- Keep the two fixes independently reviewable. Use disposable runtime data with telemetry disabled for live verification.

## Task 1 — Correlate focus recovery with its dialog

The `IssueReadyForFocusMsg` handler in `model/update.go` matches only an issue ID and opens the timer picker immediately. That lets a reply replace a form opened after recovery began. The release audit reproduced this by delivering the reply while New Repo was open.

**Files:** `tui/internal/tui/model/{core,dialogs,update,form_submission,focus_gate_test}.go` and `tui/internal/tui/commands/focus_prepare.go`.

- [x] Add model regression tests that confirm recovery, close the pending dialog, open New Repo with a typed name, and deliver the successful result. Assert the new form and typed name survive. Also cover confirming twice, late failure, and the normal success path opening the timer picker for the original issue.
- [x] Run the focused model tests and confirm the delayed reply reproduces the overwrite.
- [x] Register `prepare_focus_status` as a recoverable form action and remove the separate pending-issue-ID slot. Let the existing submission ID guard duplicate and out-of-order replies, including a new recovery for the same issue.
- [x] Keep the timer picker follow-up on the active success path only. Use the completed request's issue ID, resolve that issue again, and avoid opening the picker if an active timer or protected mode has made focus unavailable. Route active failures through inline form errors. On late success, refresh issue and daily data without changing the current dialog. On late failure, use the existing global error handling.
- [x] Run the focused model tests. In the live TUI, verify normal recovery from Backlog and verify that Escape from the pending dialog leaves a newly opened form intact when the reply arrives.

## Task 2 — Apply successful check-ins before refreshing views

`commands.UpsertDailyCheckIn` discards the saved record and returns a Bubble Tea batch message. The submission wrapper delivers that batch as an internal result, and the model calls its own `Update` with it. That bypasses the Bubble Tea program dispatch that would schedule its commands. The date's check-in cache can therefore remain empty. In addition, a check-in load currently carries only the record, so a nil reply has no date identity.

**Files:** `tui/internal/tui/commands/{actions,load,messages}.go`, `tui/internal/tui/dispatch/messages.go`, and `tui/internal/tui/model/{update,dialogs,form_submission_test}.go`.

- [x] Add regression tests that deliver the real save-command result shape through the form wrapper. After success, send `w` and assert the Edit Check-In form opens with the persisted mood, energy, and optional fields. Cover a failed save retaining inputs, a save reply after changing date, and a delayed empty load requested before the save.
- [x] Run the focused tests and confirm the current wrapper fails to update the selected date's check-in.
- [x] Return a typed success message with the record from `UpsertDailyCheckIn`. On receipt, apply the record immediately if its date is still selected, mark check-in history stale, and return the existing wellbeing and rollup refresh commands from the model as a real `tea.Cmd`. Preserve the dedicated Check-In History save workflow.
- [x] Add the requested date to check-in load replies, including empty replies. Carry load start time and save completion time so an in-flight read started before the save cannot erase the newly saved record. Reject replies for another selected date, verify the cached record's date before opening Edit Check-In, and keep a fresh load after deletion able to clear the cache.
- [x] Run the focused tests. In a disposable runtime, save a new check-in and press `w` immediately. Confirm the form opens in Edit mode with saved values. Repeat after editing an optional field, changing dates, and deleting the record.

## Release verification

- [x] Run `make ci` and `make test-e2e`. Confirm the IPC tests ran rather than skipping due to socket restrictions.
- [x] Record the isolated walkthrough outcomes in the first-run assessment.
- [x] Keep the focus-recovery and check-in changes in separate commits.
