package controller

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestIssueStatusForANamedIssueIgnoresTheSelection(t *testing.T) {
	state := Snapshot{}.OpenIssueStatusFor(42, 7, "ready", "Plan issue", "Work", "app", nil, 0)
	ctx := UpdateContext{CurrentDate: "2026-11-08", SelectedIssueID: 1, SelectedStreamID: 2, HasSelectedIssue: true}
	for i, item := range state.StatusItems {
		if item == "in_progress" {
			state.StatusCursor = i
		}
	}
	_, action, _ := Update(state, ctx, "2026-11-08", tea.KeyMsg{Type: tea.KeyEnter})
	if action == nil || action.Kind != "change_issue_status" || action.IssueID != 42 || action.StreamID != 7 {
		t.Fatalf("expected the named issue's status changed, got %+v", action)
	}
	// Without a named issue the picker keeps acting on the selection.
	plain := Snapshot{}.OpenIssueStatus("ready", "Selected", "Work", "app", nil, 0)
	plain.StatusCursor = state.StatusCursor
	if _, action, _ := Update(plain, ctx, "2026-11-08", tea.KeyMsg{Type: tea.KeyEnter}); action == nil || action.IssueID != 1 {
		t.Fatalf("expected the selection's status changed, got %+v", action)
	}
}
