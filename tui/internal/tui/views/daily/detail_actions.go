package daily

import (
	"crona.work/tui/internal/tui/views/itemdetail"
	types "crona.work/tui/internal/tui/views/types"
)

func detailFooter(theme types.Theme, state types.ContentState, width int) []string {
	if !state.DailyDetailFocus || state.DailyDetailSummary {
		return nil
	}
	var actions [][2]string
	if state.Pane == "habits" {
		if habit, ok := selectedDailyHabit(state); ok {
			actions = itemdetail.HabitFooterKeys(state, habit.Habit)
		}
	} else if issue, ok := selectedDailyIssue(state); ok {
		actions = itemdetail.IssueFooterKeys(state, issue)
	}
	if len(actions) == 0 {
		return nil
	}
	return itemdetail.ActionLines(theme, width, actions)
}
