package model

import (
	"crona.work/tui/internal/logger"
	"crona.work/tui/internal/tui/commands"
	dialogstate "crona.work/tui/internal/tui/dialogs/controller"
	helperpkg "crona.work/tui/internal/tui/helpers"

	tea "github.com/charmbracelet/bubbletea"
)

type formSubmission struct {
	id   uint64
	kind string
}
type formSubmissionResult struct {
	id         uint64
	operation  string
	message    tea.Msg
	nextDialog dialogstate.State
}

func recoverableFormAction(kind string) bool {
	switch kind {
	case "create_repo", "edit_repo", "create_stream", "edit_stream",
		"create_issue_meta", "create_issue_default", "edit_issue",
		"create_habit", "edit_habit", "create_checkin", "edit_checkin",
		"promote_adhoc_issue", "move_inbox_issue", "patch_setting",
		"patch_rest_protection", "patch_telemetry_settings", "complete_onboarding",
		"start_focus_session", "extend_hard_limit", "manual_session",
		"edit_session_history", "create_alert_reminder", "edit_alert_reminder",
		"create_momentum", "update_momentum", "set_tags", "rename_tag", "export_report", "set_export_reports_dir", "set_export_ics_dir":
		return true
	}
	return false
}

func (m Model) submitForm(next Model, action dialogstate.Action, cmd tea.Cmd) (tea.Model, tea.Cmd) {
	if cmd == nil || !recoverableFormAction(action.Kind) {
		return next, cmd
	}
	next.formSubmissionSeq++
	id := next.formSubmissionSeq
	successDialog := next.dialogState()
	state := m.dialogState()
	state.ErrorMessage = ""
	state.Processing = true
	state.ProcessingLabel = successDialog.ProcessingLabel
	if state.ProcessingLabel == "" {
		state.ProcessingLabel = "Saving…"
	}
	next = next.withDialogState(state)
	next.formSubmission = &formSubmission{id: id, kind: state.Kind}
	return next, func() tea.Msg {
		return formSubmissionResult{id: id, operation: action.Kind, message: cmd(), nextDialog: successDialog}
	}
}

func (m Model) applyFormSubmission(result formSubmissionResult) (tea.Model, tea.Cmd) {
	active := m.formSubmission != nil && m.formSubmission.id == result.id && m.dialog == m.formSubmission.kind
	var failure *commands.ErrMsg
	switch msg := result.message.(type) {
	case commands.ErrMsg:
		if msg.Operation == "" {
			msg.Operation = result.operation
		}
		failure = &msg
	case commands.CheckInHistoryMutationMsg:
		if msg.Err != nil {
			failure = &commands.ErrMsg{Err: msg.Err, Operation: "save check-in"}
		}
	}
	if !active {
		if m.formSubmission != nil && m.formSubmission.id == result.id {
			m.formSubmission = nil
		}
		if failure != nil {
			return m.update(*failure)
		}
		// A late create result may refresh data, but must not open its follow-up form.
		if created, ok := result.message.(commands.IssueCreatedMsg); ok {
			created.FollowUp = dialogstate.IssueCreateFollowUpDefault
			result.message = created
		}
		priorDialog := m.dialogState()
		updated, cmd := m.Update(result.message)
		next := updated.(Model).withDialogState(priorDialog)
		return next, cmd
	}
	m.formSubmission = nil
	if failure != nil {
		logger.Errorf("form submission %s failed: %v", failure.Operation, failure.Err)
		state := m.dialogState()
		state.Processing = false
		state.ProcessingLabel = ""
		state.ErrorMessage = helperpkg.ErrorMessage(failure.Err)
		m = m.withDialogState(state)
		return m, m.dispatchMessageDeps().ReportHandledError(failure.Err, failure.Operation)
	}
	m = m.withDialogState(result.nextDialog)
	return m.Update(result.message)
}
