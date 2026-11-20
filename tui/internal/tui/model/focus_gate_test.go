package model

import (
	"testing"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
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
