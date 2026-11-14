package habitlist

import (
	"fmt"
	"strings"

	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	"crona.work/tui/internal/tui/views/issuecore"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func habitIndices(state types.ContentState) []int {
	filter := helperpkg.ParseItemFilter(state.Filters["habits"])
	out := []int{}
	for i, habit := range state.ScopedHabits {
		if filter.Match(habit.Name, habit.Tags) {
			out = append(out, i)
		}
	}
	return out
}

func renderHabits(theme types.Theme, state types.ContentState, width, height int) string {
	active := state.Pane == "habits"
	cur := state.Cursors["habits"]
	indices := habitIndices(state)
	lines := []string{}
	if active {
		lines = append(
			lines,
			viewchrome.FitHints([]string{viewchrome.KeyHint(theme, "a", "new habit")}, width-2),
		)
	}
	if len(indices) == 0 {
		lines = append(lines, theme.StyleDim.Render("No habits in this scope"))
	}
	budget := height - 1 - len(lines)
	slots := max(1, (budget+1)/3)
	start, end := viewchrome.ListWindow(cur, len(indices), slots)
	if start > 0 {
		budget--
	}
	if end < len(indices) {
		budget--
	}
	start, end = viewchrome.ListWindow(cur, len(indices), max(1, (budget+1)/3))
	if start > 0 {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↑ %d more", start)))
	}
	for pos := start; pos < end; pos++ {
		if pos > start {
			lines = append(lines, "")
		}
		habit := state.ScopedHabits[indices[pos]]
		marker := "  "
		style := theme.StyleNormal
		if active && pos == cur {
			marker = theme.StyleCursor.Render(viewchrome.SelectionCursor + " ")
			style = style.Bold(true)
		}
		lines = append(lines, marker+style.Render(ansi.Truncate(habit.Name, max(1, width-4), "…")))
		if active && pos == cur && !state.ItemDetailActions {
			hints := []string{
				viewchrome.KeyHint(theme, "enter", "open"),
				viewchrome.KeyHint(theme, "e", "edit"),
			}
			if state.Timer == nil || state.Timer.State == "idle" {
				hints = append(
					hints,
					viewchrome.KeyHint(theme, "D", "delete"),
					viewchrome.KeyHint(theme, "m", "log"),
				)
			}
			hints = append(hints, viewchrome.KeyHint(theme, "+", "tags"))
			lines = append(lines, "  "+viewchrome.FitHints(hints, width-4))
			continue
		}
		metaLine := theme.StyleDim.Render("  " + ansi.Truncate(habitMeta(habit), max(1, width-4), "…"))
		if chips := issuecore.TagChips(theme, habit.Tags, state.Tags, max(0, (width-4)/2)); chips != "" {
			metaLine = theme.StyleDim.Render("  "+ansi.Truncate(habitMeta(habit), max(1, width-6-ansi.StringWidth(chips)), "…")) + "  " + chips
		}
		lines = append(lines, metaLine)
	}
	if end < len(indices) {
		lines = append(lines, theme.StyleDim.Render(fmt.Sprintf("↓ %d more", len(indices)-end)))
	}
	return viewchrome.RenderSection(
		theme,
		viewchrome.Section{
			Title:   "Habits",
			Meta:    viewhelpers.FilterMeta(fmt.Sprintf("%d", len(indices)), state.Filters["habits"]),
			Focused: active,
			Dimmed:  !active,
			Width:   width,
			Height:  height,
			Lines:   lines,
		},
	)
}
func habitMeta(habit api.HabitWithMeta) string {
	schedule := string(habit.ScheduleType)
	if schedule == "weekly" {
		names := []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"}
		days := []string{}
		for _, day := range habit.Weekdays {
			if day >= 0 && day < len(names) {
				days = append(days, names[day])
			}
		}
		schedule = strings.Join(days, ",")
	}
	target := "-"
	if habit.TargetMinutes != nil {
		target = helperpkg.FormatCompactDurationMinutes(*habit.TargetMinutes)
	}
	status := "active"
	if !habit.Active {
		status = "paused"
	}
	return habit.RepoName + " › " + habit.StreamName + " | " + schedule + " | " + target + " | " + status
}
