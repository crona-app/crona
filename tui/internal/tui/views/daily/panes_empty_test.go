package daily

import (
	"strings"
	"testing"
)

func TestDailyEmptyTasksExplainsBacklogAndBrowseAction(t *testing.T) {
	state := dailyState(100, 20)
	state.DailyIssues = nil
	state.DailyIssueIndices = nil
	state.DailyBacklogCount = 12
	state.DailyTaskCounts = map[string]int{"overdue": 3, "upcoming": 5}
	state.DailyHasIssues = true
	rendered := renderIssues(testTheme(), state, 100, 10)

	for _, want := range []string{
		"No issues planned for this date",
		"12 in Backlog",
		"3 overdue",
		"5 upcoming",
		"[2] plan issues in Work",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("empty task view missing %q in %q", want, rendered)
		}
	}
}

func TestDailyEmptyTasksWithFilterHasFilterSpecificCopy(t *testing.T) {
	state := dailyState(100, 20)
	state.Filters["issues"] = "missing"
	state.DailyIssues = nil
	state.DailyIssueIndices = nil
	rendered := renderIssues(testTheme(), state, 100, 10)
	if !strings.Contains(rendered, "No issues match the filter") {
		t.Fatalf("filtered empty view lacks filter-specific explanation: %q", rendered)
	}
}

func TestDailyFirstRunEmptyTasksOffersCreation(t *testing.T) {
	state := dailyState(100, 20)
	state.DailyIssues = nil
	state.DailyIssueIndices = nil
	state.DailyBacklogCount = 0
	state.DailyTaskCounts = map[string]int{}
	rendered := renderIssues(testTheme(), state, 100, 10)
	if !strings.Contains(rendered, "No issues yet") ||
		!strings.Contains(rendered, "[a] new issue") ||
		!strings.Contains(rendered, "[I] bulk add") {
		t.Fatalf("first-run empty state is not actionable: %q", rendered)
	}
}

func TestDailyEmptyTasksSuppressesZeroCountsAndHints(t *testing.T) {
	state := dailyState(100, 20)
	state.DailyIssues = nil
	state.DailyIssueIndices = nil
	state.DailyHasIssues = true
	state.DailyTaskCounts = map[string]int{"overdue": 0, "upcoming": 0}
	rendered := renderIssues(testTheme(), state, 100, 10)
	if strings.Contains(rendered, "Backlog") || strings.Contains(rendered, "plan issues in Work") {
		t.Fatalf("zero counts should not add dangling hints: %q", rendered)
	}
}
