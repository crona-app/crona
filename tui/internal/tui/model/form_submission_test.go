package model

import (
	"path/filepath"
	"testing"
	"time"

	"crona.work/tui/internal/api"
	"crona.work/tui/internal/tui/commands"
	dialogstate "crona.work/tui/internal/tui/dialogs/controller"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFailedIssueSubmissionKeepsInputAndAllowsRetry(t *testing.T) {
	m := NewDailyRenderModel(60, 20)
	m.client = api.NewClient("unix", filepath.Join(t.TempDir(), "missing.sock"))
	state := dialogstate.OpenCreateIssueMeta(m.dialogState(), 7, "app", "Work")
	state.Inputs[0].SetValue("Keep this unsaved issue")
	state.Description.SetValue("Keep this description")
	m = m.withDialogState(state)
	submitted, cmd := m.updateDialog(tea.KeyMsg{Type: tea.KeyCtrlS})
	pending := submitted.(Model)
	if pending.dialog == "" {
		t.Fatal("form closed before the save result")
	}
	if cmd == nil {
		t.Fatal("missing submission command")
	}
	result, _ := pending.update(cmd())
	failed := result.(Model)
	if failed.dialog != "create_issue_meta" || failed.dialogState().Inputs[0].Value() != "Keep this unsaved issue" {
		t.Fatal("failed save lost form or title")
	}
	if failed.dialogState().Description.Value() != "Keep this description" {
		t.Fatal("failed save lost description")
	}
	if failed.dialogErrorMessage == "" {
		t.Fatal("failed save has no inline error")
	}
	_, retry := failed.updateDialog(tea.KeyMsg{Type: tea.KeyCtrlS})
	if retry == nil {
		t.Fatal("failed form cannot be retried")
	}
}

func TestPendingSubmissionRejectsDuplicateAndCanBeCancelled(t *testing.T) {
	m := NewDailyRenderModel(60, 20)
	m.client = api.NewClient("unix", filepath.Join(t.TempDir(), "missing.sock"))
	state := dialogstate.OpenCreateIssueMeta(m.dialogState(), 7, "app", "Work")
	state.Inputs[0].SetValue("Submit once")
	m = m.withDialogState(state)
	submitted, cmd := m.updateDialog(tea.KeyMsg{Type: tea.KeyCtrlS})
	pending := submitted.(Model)
	if pending.dialog == "" {
		t.Fatal("form closed before the save result")
	}
	_, duplicate := pending.updateDialog(tea.KeyMsg{Type: tea.KeyCtrlS})
	if duplicate != nil {
		t.Fatal("pending form submitted twice")
	}
	cancelled, _ := pending.updateDialog(tea.KeyMsg{Type: tea.KeyEsc})
	next := cancelled.(Model).withDialogState(dialogstate.OpenCreateRepo(cancelled.(Model).dialogState()))
	late, _ := next.update(cmd())
	if late.(Model).dialog != "create_repo" {
		t.Fatal("late failure restored a cancelled form")
	}
	if late.(Model).dialogErrorMessage != "" {
		t.Fatal("late failure attached to a different form")
	}
}

func TestSuccessfulCreateKeepsFocusFollowUp(t *testing.T) {
	m := NewDailyRenderModel(120, 36)
	state := dialogstate.OpenCreateIssueMeta(m.dialogState(), 7, "app", "Work")
	state.Inputs[0].SetValue("Focus next")
	m = m.withDialogState(state)
	next := m.withDialogState(dialogstate.Close(state))
	pending, cmd := m.submitForm(next, dialogstate.Action{Kind: "create_issue_meta"}, func() tea.Msg {
		return commands.IssueCreatedMsg{Issue: api.Issue{ID: 11, StreamID: 7, Title: "Focus next"}, RepoID: 3, StreamID: 7, FollowUp: dialogstate.IssueCreateFollowUpFocus}
	})
	completed, _ := pending.(Model).update(cmd())
	if completed.(Model).dialog != "timer_start_type" {
		t.Fatalf("successful save lost focus follow-up: %q", completed.(Model).dialog)
	}
}

func TestCancelledSubmissionSuccessDoesNotOpenFollowUp(t *testing.T) {
	m := NewDailyRenderModel(120, 36)
	state := dialogstate.OpenCreateIssueMeta(m.dialogState(), 7, "app", "Work")
	m = m.withDialogState(state)
	pending, cmd := m.submitForm(m.withDialogState(dialogstate.Close(state)), dialogstate.Action{Kind: "create_issue_meta"}, func() tea.Msg {
		return commands.IssueCreatedMsg{Issue: api.Issue{ID: 11, StreamID: 7, Title: "Focus next"}, RepoID: 3, StreamID: 7, FollowUp: dialogstate.IssueCreateFollowUpFocus}
	})
	cancelled, _ := pending.(Model).updateDialog(tea.KeyMsg{Type: tea.KeyEsc})
	next := cancelled.(Model).withDialogState(dialogstate.OpenCreateRepo(cancelled.(Model).dialogState()))
	completed, _ := next.update(cmd())
	if completed.(Model).dialog != "create_repo" {
		t.Fatal("late success opened another form")
	}
}

func TestLateExportResultDoesNotReplaceAnotherForm(t *testing.T) {
	m := NewDailyRenderModel(120, 36)
	m = m.withDialogState(dialogstate.OpenCreateRepo(m.dialogState()))
	result, _ := m.update(formSubmissionResult{id: 1, message: commands.CalendarExportGeneratedMsg{Result: &api.CalendarExportResult{}}})
	if result.(Model).dialog != "create_repo" {
		t.Fatalf("late export replaced the active form: %q", result.(Model).dialog)
	}
}

func TestLateClipboardResultDoesNotCloseNewExport(t *testing.T) {
	m := NewDailyRenderModel(120, 36)
	state := m.dialogState()
	state.Kind = "export_preset"
	state.Processing = true
	m = m.withDialogState(state)
	result, _ := m.update(formSubmissionResult{id: 1, message: commands.ClipboardCopiedMsg{Message: "Old export copied"}})
	if result.(Model).dialog != state.Kind {
		t.Fatal("old export result closed the newer export")
	}
}

func TestExportPathSaveFailureKeepsEnteredPath(t *testing.T) {
	for _, open := range []func(dialogstate.State, string) dialogstate.State{dialogstate.OpenExportReportsDir, dialogstate.OpenExportICSDir} {
		m := NewDailyRenderModel(60, 20)
		m.client = api.NewClient("unix", filepath.Join(t.TempDir(), "missing.sock"))
		state := open(m.dialogState(), "/tmp/my-reports")
		m = m.withDialogState(state)
		pending, cmd := m.updateDialog(tea.KeyMsg{Type: tea.KeyCtrlS})
		if pending.(Model).dialog != state.Kind {
			t.Fatal("settings form closed before save result")
		}
		failed, _ := pending.(Model).update(cmd())
		if failed.(Model).dialogState().Inputs[0].Value() != "/tmp/my-reports" {
			t.Fatal("failed settings save lost entered path")
		}
		if failed.(Model).dialogErrorMessage == "" {
			t.Fatal("failed settings save has no inline error")
		}
	}
}

func TestCompletedExportShowsFullPath(t *testing.T) {
	m := NewDailyRenderModel(60, 20)
	state := m.dialogState()
	state.Kind = "export_report_category"
	m = m.withDialogState(state)
	state.Processing = true
	path := "/tmp/crona-export-test/reports/2026/weekly-summary-2026-11-25.md"
	pending, cmd := m.submitForm(m.withDialogState(state), dialogstate.Action{Kind: "export_report"}, func() tea.Msg {
		return commands.DailyReportGeneratedMsg{Result: &api.DailyReportResult{
			Label: "Weekly summary", OutputMode: "file", FilePath: &path,
		}}
	})
	completed, _ := pending.(Model).update(cmd())
	result := completed.(Model).dialogState()
	if result.ViewTitle != "Export Complete" || result.ViewBody != path {
		t.Fatalf("expected an export receipt with the complete path, got %q, %q", result.ViewTitle, result.ViewBody)
	}
}

func TestSavedCheckInReopensWithPersistedValues(t *testing.T) {
	m := NewDailyRenderModel(60, 20)
	m.date = "2026-11-26"
	m = m.openCheckInDialogForDate(m.date)
	saved := &api.DailyCheckIn{Date: m.date, Mood: 4, Energy: 3, Notes: new("Persisted notes")}
	pending, cmd := m.submitForm(
		m.withDialogState(dialogstate.Close(m.dialogState())),
		dialogstate.Action{Kind: "create_checkin"},
		func() tea.Msg {
			return commands.DailyCheckInSavedMsg{CheckIn: saved, CompletedAt: time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)}
		},
	)
	result, _ := pending.(Model).Update(cmd())
	reopened, _ := press(result.(Model), "w")
	state := reopened.dialogState()
	if state.Kind != "edit_checkin" || state.Inputs[0].Value() != "4" ||
		state.Inputs[1].Value() != "3" || state.Inputs[5].Value() != "Persisted notes" {
		t.Fatalf("check-in did not reopen with the saved values: %q", state.Kind)
	}
}

func TestCheckInRepliesRespectDateAndReadOrder(t *testing.T) {
	m := NewDailyRenderModel(60, 20)
	m.date = "2026-11-26"
	at := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	saved := &api.DailyCheckIn{Date: m.date, Mood: 4, Energy: 3}
	apply := func(message tea.Msg) {
		t.Helper()
		next, _ := m.Update(message)
		m = next.(Model)
	}
	apply(commands.DailyCheckInSavedMsg{CheckIn: saved, CompletedAt: at})
	apply(commands.DailyCheckInLoadedMsg{Date: m.date, RequestedAt: at.Add(-time.Second)})
	if m.dailyCheckIn != saved {
		t.Fatal("a stale empty read erased the saved check-in")
	}

	m.date = "2026-11-25"
	if state := m.openCheckInDialogForDate(m.date).dialogState(); state.Kind != "create_checkin" {
		t.Fatal("a cached record from the previous date should not open in edit mode")
	}
	other := &api.DailyCheckIn{Date: m.date, Mood: 2, Energy: 1}
	apply(commands.DailyCheckInLoadedMsg{Date: m.date, CheckIn: other, RequestedAt: at.Add(time.Second)})
	apply(commands.DailyCheckInSavedMsg{CheckIn: saved, CompletedAt: at.Add(2 * time.Second)})
	apply(commands.DailyCheckInLoadedMsg{Date: saved.Date, RequestedAt: at.Add(3 * time.Second)})
	if m.dailyCheckIn != other {
		t.Fatal("a response for a different date replaced the current record")
	}
	apply(commands.DailyCheckInLoadedMsg{Date: m.date, RequestedAt: at.Add(4 * time.Second)})
	if m.dailyCheckIn != nil {
		t.Fatal("a fresh empty load should clear a deleted check-in")
	}
}

func TestFailedCheckInSaveKeepsEnteredValues(t *testing.T) {
	m := NewDailyRenderModel(60, 20)
	m.date = "2026-11-26"
	m.client = api.NewClient("unix", filepath.Join(t.TempDir(), "missing.sock"))
	form := dialogstate.OpenCreateCheckIn(m.dialogState(), m.date)
	form.Inputs[0].SetValue("4")
	form.Inputs[1].SetValue("3")
	m = m.withDialogState(form)
	pending, cmd := m.updateDialog(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatal("missing check-in save command")
	}
	result, _ := pending.(Model).Update(cmd())
	state := result.(Model).dialogState()
	if state.Kind != "create_checkin" || state.ErrorMessage == "" ||
		state.Inputs[0].Value() != "4" || state.Inputs[1].Value() != "3" {
		t.Fatalf("failed check-in save lost entered values or its inline error: %q", state.Kind)
	}
}
