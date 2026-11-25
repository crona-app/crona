# Error visibility and form recovery

Status: approved 2026-11-25. Scope: TUI first-run assessment task 1.

## Findings

- `model/status.go` clears shell errors after five seconds. `layout/render.go`
  collapses them to one truncated footer line, replacing normal help text.
- Shell status is absent while a dialog or session overlay is open because
  those layouts return before rendering the footer.
- Dialog validation already has a dedicated inline error area. However, issue
  create and edit submissions can close their form before the daemon replies.
- Generic asynchronous errors use `err.Error()`, which can expose RPC codes.
  A background load error can also be attached to whichever dialog is open,
  even if that dialog did not cause it.

## Chosen approach

The user selected persistent notices for failures outside forms and inline form
errors. Keeping errors only in the footer would still truncate explanations.
Modal error dialogs would interrupt navigation and compete with existing forms.

### Persistent notices

- Render a visually distinct, wrapped error notice above the ordinary footer.
  Preserve the normal action hints. Show the notice with dialogs and session
  overlays as well as ordinary views.
- Errors do not expire or disappear when an unrelated success message arrives.
  Escape dismisses a visible global notice and consumes that key, leaving any
  underlying dialog, selection, or filter unchanged. Display this dismissal hint.
- Keep one current notice; a newer failure replaces it. Do not queue repeated
  connection/load failures or automatically retry mutations.
- Bound notice height so it fits at 60x20. Wrap by terminal display width;
  split long tokens. If a message exceeds the available notice space, indicate
  that the full details are in logs rather than letting it push controls offscreen.
- Display RPC messages without machine-code prefixes. Keep original errors and
  operation context in logs and telemetry. Provide operation context or a safe
  next step where known; do not invent a universal retry action for unknown failures.

### Form recovery

- Keep create/edit form values until the associated mutation succeeds. Show a
  saving state while it is pending and prevent duplicate submissions.
- On failure, keep the form open with the same values and an inline error;
  restore the ordinary submit/cancel hints so users can correct or retry it.
- Correlate results with their submission. Background load errors become global
  notices and do not overwrite form validation. Late replies must not close or
  restore a different form.
- Preserve existing success behavior, including issue follow-up choices and
  reloads. Treat failures after a partial mutation cautiously: show the actual
  failure, retain input, and do not replay the operation automatically.
- Start with issue creation/editing, repo/stream and habit forms, check-in,
  timer setup, and settings saves. Inspect each command's result boundary;
  do not assume a batch of reloads proves a mutation failed or succeeded.

## Verification

Regression tests must reproduce the current defects before fixes: timeout and
success overwrite, long/narrow messages, overlay visibility, retained failed
form input, submit deduplication, and unrelated/late reply handling. Verify RPC
message presentation separately from diagnostic logging.

Run focused package tests, then unit tests, daemon e2e tests if IPC behavior is
touched, and lint. Inspect rendered 60x20 and wider frames. Exercise failure and
retry in an isolated runtime without changing the user's live data.

## Non-goals

No notification history, daemon reconnection redesign, automatic mutation retry,
new protocol methods, export wizard redesign, or changes to other assessment
passes. The native macOS companion is outside this TUI task.
