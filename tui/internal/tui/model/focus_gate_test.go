package model

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	"crona.work/tui/internal/tui/commands"
	dialogstate "crona.work/tui/internal/tui/dialogs/controller"
)

func TestFocusOnBacklogOpensGuardDialogWithoutStarting(t *testing.T) {
	issue := api.IssueWithMeta{
		Issue:      sharedtypes.Issue{ID: 22, StreamID: 7, Title: "Ship onboarding", Status: sharedtypes.IssueStatusBacklog},
		RepoID:     3,
		RepoName:   "Work",
		StreamName: "app",
	}
	repoID := int64(3)
	m := Model{
		view:      ViewDefault,
		pane:      PaneIssues,
		issues:    []api.Issue{issue.Issue},
		allIssues: []api.IssueWithMeta{issue},
		context:   &api.ActiveContext{RepoID: &repoID},
		cursor:    map[Pane]int{PaneIssues: 0},
	}

	got, cmd := m.handleInputStartFocusFromSelection()
	if cmd != nil {
		t.Fatal("backlog focus must not start a session before explicit confirmation")
	}
	next := got.(Model)
	if next.dialogState().Kind != "focus_blocked" {
		t.Fatalf("dialog = %q, want focus_blocked", next.dialogState().Kind)
	}
	if next.dialogState().IssueID != issue.ID {
		t.Fatalf("issue id = %d, want %d", next.dialogState().IssueID, issue.ID)
	}
	if next.dialogSnapshot().Dialog.RecoveryStatus != "ready" {
		t.Fatalf("recovery status = %q, want ready", next.dialogSnapshot().Dialog.RecoveryStatus)
	}
}

func TestFocusRecoveryReplyRespectsSubmission(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cancel  bool
		message tea.Msg
		want    string
	}{
		{"active success", false, commands.IssueReadyForFocusMsg{IssueID: 22}, "timer_start_type"},
		{"active failure", false, commands.ErrMsg{Err: errors.New("daemon unavailable")}, "focus_blocked"},
		{"late success", true, commands.IssueReadyForFocusMsg{IssueID: 22}, "create_repo"},
		{"late failure", true, commands.ErrMsg{Err: errors.New("daemon unavailable")}, "create_repo"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := NewDailyRenderModel(60, 20)
			m.allIssues = []api.IssueWithMeta{{
				Issue:  sharedtypes.Issue{ID: 22, StreamID: 7, Title: "Focus"},
				RepoID: 3,
			}}
			m = m.withDialogState(dialogstate.OpenFocusBlocked(m.dialogState(), 22, "Focus", "backlog", "ready"))
			submitted, cmd := m.updateDialog(tea.KeyMsg{Type: tea.KeyEnter})
			m = submitted.(Model)
			if cmd == nil || m.formSubmission == nil || !m.dialogState().Processing {
				t.Fatal("recovery should keep a pending form until the reply")
			}
			result := formSubmissionResult{id: m.formSubmission.id, message: tc.message, nextDialog: dialogstate.Close(m.dialogState())}
			if _, duplicate := m.updateDialog(tea.KeyMsg{Type: tea.KeyEnter}); duplicate != nil {
				t.Fatal("duplicate recovery submitted while pending")
			}
			if tc.cancel {
				closed, _ := m.updateDialog(tea.KeyMsg{Type: tea.KeyEsc})
				m = closed.(Model)
				form := dialogstate.OpenCreateRepo(m.dialogState())
				form.Inputs[0].SetValue("Keep my repo name")
				m = m.withDialogState(form)
			}
			updated, _ := m.Update(result)
			state := updated.(Model).dialogState()
			if state.Kind != tc.want {
				t.Fatalf("got dialog %q, want %q", state.Kind, tc.want)
			}
			if tc.cancel && state.Inputs[0].Value() != "Keep my repo name" {
				t.Fatal("late reply lost the new form's input")
			}
			if tc.want == "timer_start_type" && state.IssueID != 22 {
				t.Fatal("timer picker lost the recovery target")
			}
			if tc.want == "focus_blocked" && state.ErrorMessage == "" {
				t.Fatal("recovery failure should be visible inline")
			}
		})
	}
}
