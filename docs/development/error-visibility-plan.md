# Error Visibility Implementation Plan

Goal: persistent shell errors and recoverable form submissions in the TUI.
Spec: [approved design](error-visibility-design.md).
Execution: implement inline on the existing `fix/first-run-focus-flow` branch.
Use the executing-plans and test-driven-development skills for each task.

## Constraints and review focus

No daemon protocol changes, automatic mutation retries, or new dependencies.
Preserve original errors in diagnostics. Keep notices visible at 60x20 without
obscuring controls. Correlate submission results; do not confuse a mutation
failure with a later reload failure. Preserve existing success follow-ups.

## Task 1: Persistent error notice

Files: model/status.go, model/update.go, model/core.go, layout/render.go;
new layout/error_notice.go and focused layout/model tests.

- [x] Add regression tests: an error survives a newer success and stale clear;
  Escape dismisses it without closing a dialog; notices render with overlays
  and at 60x20; long tokens never exceed the viewport.
- [x] Run the tests and verify failure.
- [x] Keep a separate persistent error string in model state. Route global
  asynchronous errors there, reserving inline errors for correlated submissions.
- [x] Render a bounded wrapped notice with dismissal hint above the footer;
  reserve its height in ordinary and overlay layouts. Keep success status timed.
- [x] Format typed RPC errors as their message; log the original error.
- [x] Run focused model/layout/dispatch tests.

## Task 2: Correlated form submissions

Files: model/update.go and model/core.go; new model/form_submission.go and
form_submission_test.go; affected command mutation/result boundaries as needed.

- [x] Add regression tests: failed form submission retains values; success closes
  the form and preserves result handling; resubmission while pending is ignored;
  unrelated errors remain global; late replies cannot close another dialog.
- [x] Run tests and verify failure.
- [x] Keep a submitted dialog snapshot until its command completes. Wrap the
  mutation result with a submission identifier. Disable repeated submission while
  pending; permit cancel and handle subsequent results without restoring a form.
- [x] Associate synchronous form validation and mutation failure with the form;
  leave refresh failures global. Audit issue, repo/stream, habit, check-in, timer,
  and settings commands for result boundaries and preserve success follow-ups.
- [x] Run focused model/commands/dialog tests.

## Task 3: Verify and document

- [x] Run `make test-unit` and `make lint`; use `gofmt` on touched files only.
- [x] Inspect rendered narrow and wide frames and exercise failure/correction in
  an isolated runtime with telemetry disabled. Do not touch the user's live data.
- [x] Review the complete diff, update the assessment status, and report any
  validation limitations. Do not commit or deploy unless requested.

## Review and implementation notes

- Independent review found late export successes replacing newer dialogs and
  long inline messages hiding submit/cancel hints. Both reproduced in failing
  tests and now pass; export-path settings saves also retain their input.
- Ruling: multi-field settings and tagged edits finish their mutations before
  returning refresh messages, so submission tracking can distinguish save failure
  from a later refresh failure. Partial saves are not automatically replayed.
- Ruling: modal notice screens show only the global quit hint below the notice;
  underlying pane shortcuts are inactive while a form is open.
- No protocol changes or native companion changes were needed. Live checks used
  disposable runtime data; the installed TUI was not replaced.
