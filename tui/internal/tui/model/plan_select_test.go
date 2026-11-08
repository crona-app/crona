package model

import (
	"testing"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	inputpkg "crona.work/tui/internal/tui/input"
)

func TestPlanCursorFollowsItsIssueAsTheListReorders(t *testing.T) {
	m := Model{
		dailyDetailFocus:   true,
		dailyDetailSummary: true,
		summarySnapshot: &api.SummarySnapshot{Plan: &api.DailyPlan{Contributing: []api.DailyPlanEntry{
			{IssueID: 1, Status: sharedtypes.DailyPlanEntryStatusPlanned, FailScore: 0.9},
			{IssueID: 2, Status: sharedtypes.DailyPlanEntryStatusPlanned, FailScore: 0.5},
		}}},
		allIssues: []api.IssueWithMeta{
			{Issue: api.Issue{ID: 1, Title: "One", Status: "ready"}},
			{Issue: api.Issue{ID: 2, Title: "Two", Status: "ready"}},
		},
	}
	m = m.applyPlanSelect(inputpkg.State{DailyPlanSelect: true, DailyPlanCursor: 1})
	if issue, ok := m.selectedPlanIssue(); !ok || issue.ID != 2 {
		t.Fatalf("expected issue 2 selected, got %+v", issue)
	}
	// Issue 2 grows riskier and moves to the top; the cursor goes with it.
	m.summarySnapshot.Plan.Contributing[1].FailScore = 1.5
	if cursor := m.planCursor(m.planItems()); cursor != 0 {
		t.Fatalf("expected the cursor on issue 2 at the top, got row %d", cursor)
	}
	// Leaving the summary ends the selection.
	m.dailyDetailSummary = false
	if m = m.applyPlanSelect(inputpkg.State{DailyPlanSelect: true}); m.dailyPlanSelect {
		t.Fatal("expected the plan selection to end with the summary")
	}
}
