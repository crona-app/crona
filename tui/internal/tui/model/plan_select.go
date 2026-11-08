package model

import (
	"crona.work/tui/internal/api"
	inputpkg "crona.work/tui/internal/tui/input"
	viewsummary "crona.work/tui/internal/tui/views/summary"
)

// planItems are the Day summary's open plan entries in the order its plan list shows them.
func (m Model) planItems() []viewsummary.PlanItem {
	return viewsummary.PlanItems(m.summarySnapshot, m.allIssues)
}

// planCursor is the selected row of the plan list: where the selected issue now sits, or
// the last row index clamped to the list when that issue has left it (done, abandoned).
func (m Model) planCursor(items []viewsummary.PlanItem) int {
	for i, item := range items {
		if item.Entry.IssueID == m.dailyPlanIssueID {
			return i
		}
	}
	return min(max(0, m.dailyPlanCursor), max(0, len(items)-1))
}

// selectedPlanIssue is the issue under the plan list's cursor.
func (m Model) selectedPlanIssue() (*api.IssueWithMeta, bool) {
	items := m.planItems()
	if !m.dailyPlanSelect || len(items) == 0 {
		return nil, false
	}
	return m.issueMetaByID(items[m.planCursor(items)].Entry.IssueID)
}

// applyPlanSelect takes the plan list's selection from input: it lasts only while the Day
// summary is focused, and remembers the issue so the cursor follows it.
func (m Model) applyPlanSelect(state inputpkg.State) Model {
	m.dailyPlanSelect = state.DailyPlanSelect && m.dailyDetailFocus && m.dailyDetailSummary
	if !m.dailyPlanSelect {
		m.dailyPlanCursor, m.dailyPlanIssueID = 0, 0
		return m
	}
	items := m.planItems()
	if len(items) == 0 {
		m.dailyPlanSelect = false
		return m
	}
	// Input's cursor indexes the list as last shown, which planCursor derived from the issue.
	m.dailyPlanCursor = min(max(0, state.DailyPlanCursor), len(items)-1)
	m.dailyPlanIssueID = items[m.dailyPlanCursor].Entry.IssueID
	return m
}

func (m Model) openPlanIssueStatus(state *inputpkg.State) bool {
	next := m.applyInputState(*state)
	issue, ok := next.selectedPlanIssue()
	if !ok {
		return false
	}
	next = next.withDialogState(next.dialogSnapshot().OpenIssueStatusFor(
		issue.ID,
		issue.StreamID,
		string(issue.Status),
		issue.Title,
		issue.RepoName,
		issue.StreamName,
		issue.EstimateMinutes,
		issue.WorkedSeconds,
	))
	*state = next.inputState()
	return true
}

func (m Model) openPlanIssueTodoDate(state *inputpkg.State) bool {
	next := m.applyInputState(*state)
	issue, ok := next.selectedPlanIssue()
	if !ok {
		return false
	}
	next = next.withDialogState(next.dialogSnapshot().OpenDatePicker("", issue.ID, 0, issue.TodoForDate))
	*state = next.inputState()
	return true
}
