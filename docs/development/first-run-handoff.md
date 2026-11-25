# First-run assessment handoff

Branch: `fix/first-run-focus-flow`. Updated 2026-11-25.

Completed: actionable Today empty states, explicit focus-status recovery,
readable CLI errors, and TUI error visibility/recovery.

The TUI now keeps global failures visible until Escape dismisses them. Failed
form saves retain input and show inline errors. Pending submissions reject
duplicates; late replies cannot replace another form. Machine error codes stay
in diagnostics. Lost replies warn that a change may already have been saved;
mutations are never automatically retried.

Validation passed: `make test-unit`, `make lint`, `go vet ./tui/...`, and focused
race tests for model, layout, dialog_runtime, and helpers. Independent review
findings were fixed with regression tests. A disposable daemon/TUI walkthrough
verified failure, retained input, notice dismissal, and successful retry by
temporarily moving only its test socket. Both processes were stopped. Installed
apps and normal user data were untouched.

Next: assess help/key-hint accuracy against the actual input handlers. Include
the newly observed High-priority issue: onboarding content clips at 60x20.
Reproduce each onboarding page at that size before changing its layout.
Later passes: form hints, vocabulary, exports, column ratios, Momentum copy.

Read [assessment](first-run-assessment.md) for priorities and evidence, and
[error design](error-visibility-design.md) / [implementation record](error-visibility-plan.md)
for the completed change. The older `.handoff/2026-11-20-p1-first-run-focus.md`
has a stale unchecked checklist; do not redo its committed implementation.

Keep follow-up work on this branch unless asked otherwise. Use isolated runtime
data with telemetry disabled, keep commits focused, and use `gofmt` only on
touched files. The separate macOS sound fix lives in the sibling macOS repo.
