package itemdetail

import (
	"crona.work/tui/internal/api"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	types "crona.work/tui/internal/tui/views/types"
)

// detailKey attaches an action to its data in an interactive item pane or dialog.
func detailKey(theme types.Theme, state types.ContentState, key, label string) string {
	if !state.ItemDetailActions && (state.View != "daily" || !state.DailyDetailFocus || state.DailyDetailSummary) {
		return ""
	}
	return "  " + viewchrome.KeyHint(theme, key, label)
}

func detailLogKey(theme types.Theme, state types.ContentState) string {
	if state.Timer != nil && state.Timer.State != "idle" {
		return ""
	}
	return detailKey(theme, state, "m", "log")
}

// IssueFooterKeys lists actions that do not sit beside a field in issue details.
func IssueFooterKeys(state types.ContentState, issue api.Issue) [][2]string {
	idle := state.Timer == nil || state.Timer.State == "idle"
	if issue.Adhoc {
		if idle {
			return [][2]string{{"e", "convert"}}
		}
		return nil
	}
	keys := [][2]string{}
	if !issue.PinnedDaily {
		keys = append(keys, [2]string{"P", "pin"})
	}
	if len(issue.Tags) == 0 {
		keys = append(keys, [2]string{"+", "tags"})
	}
	if state.View == "daily" && state.DailyTaskSection == "overdue" {
		keys = append(keys, [2]string{"r", "carry"})
	}
	if idle {
		keys = append(keys, [2]string{"f", "focus"}, [2]string{"e", "edit"}, [2]string{"D", "delete"})
		if state.DetailSessionsIssueID != issue.ID {
			keys = append(keys, [2]string{"m", "log"})
		}
	}
	return keys
}

// HabitFooterKeys lists general habit actions; logging appears with recent history.
func HabitFooterKeys(state types.ContentState, habit api.Habit) [][2]string {
	keys := [][2]string{}
	if len(habit.Tags) == 0 {
		keys = append(keys, [2]string{"+", "tags"})
	}
	keys = append(keys, [2]string{"e", "edit"})
	if state.Timer == nil || state.Timer.State == "idle" {
		keys = append(keys, [2]string{"D", "delete"})
	}
	return keys
}
