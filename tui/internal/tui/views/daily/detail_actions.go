package daily

import (
	"strings"

	viewchrome "crona.work/tui/internal/tui/views/chrome"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// detailActionStrip wraps whole key hints, so narrow detail panes never lose actions.
func detailActionStrip(theme types.Theme, title string, width int, actions ...[2]string) []string {
	if len(actions) == 0 {
		return nil
	}
	lines := []string{}
	line := ""
	if title != "" {
		line = theme.StyleDim.Bold(true).Render(title) + "  "
	}
	for _, action := range actions {
		hint := viewchrome.KeyHint(theme, action[0], action[1])
		if ansi.StringWidth(line+hint) > width && strings.TrimSpace(ansi.Strip(line)) != title {
			lines = append(lines, strings.TrimRight(line, " "))
			line = "  "
		}
		line += hint + "  "
	}
	return append(lines, strings.TrimRight(line, " "))
}

func detailFooter(theme types.Theme, state types.ContentState, width int) []string {
	if !state.DailyDetailFocus || state.DailyDetailSummary {
		return nil
	}
	var actions [][2]string
	if state.Pane == "habits" {
		if habit, ok := selectedDailyHabit(state); ok && len(habit.Tags) == 0 {
			actions = append(actions, [2]string{"+", "tags"})
		}
		actions = append(actions, [2]string{"e", "edit"})
		if timerIdle(state) {
			actions = append(actions, [2]string{"D", "delete"})
		}
	} else if issue, ok := selectedDailyIssue(state); ok {
		if issue.Adhoc {
			if timerIdle(state) {
				actions = append(actions, [2]string{"e", "convert"})
			}
		} else {
			if !issue.PinnedDaily {
				actions = append(actions, [2]string{"P", "pin"})
			}
			if len(issue.Tags) == 0 {
				actions = append(actions, [2]string{"+", "tags"})
			}
			if state.DailyTaskSection == "overdue" {
				actions = append(actions, [2]string{"r", "carry"})
			}
			if timerIdle(state) {
				actions = append(actions, [2]string{"f", "focus"}, [2]string{"e", "edit"}, [2]string{"D", "delete"})
				if state.DetailSessionsIssueID != issue.ID {
					actions = append(actions, [2]string{"m", "log"})
				}
			}
		}
	}
	if len(actions) == 0 {
		return nil
	}
	return detailActionStrip(theme, "", width, actions...)
}
