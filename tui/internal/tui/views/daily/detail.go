package daily

import (
	"time"

	"crona.work/tui/internal/api"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	"crona.work/tui/internal/tui/views/itemdetail"
	viewruntime "crona.work/tui/internal/tui/views/runtime"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// renderDetail is the detail pane: the Day summary, or with the pane focused the selection's
// details, titled with what it shows.
func renderDetail(theme types.Theme, state types.ContentState, width, height int) string {
	title, lines := detailBody(theme, state, max(1, width-2))
	// Unfocused, the lists' action lines already offer enter, so the title carries no key.
	meta := ""
	scroll := 0
	if state.DailyDetailFocus {
		meta = viewchrome.KeyHint(theme, "esc", "back")
		scroll = state.DailyDetailScroll
	}
	visible := max(1, height-1)
	scroll = min(scroll, max(0, len(lines)-visible))
	lines = lines[scroll:]
	if more := len(lines) > visible; more || scroll > 0 {
		arrows := theme.StyleDim.Render(scrollArrows(scroll > 0, more))
		if meta == "" {
			meta = arrows
		} else {
			meta = arrows + "  " + meta
		}
	}
	// A long title gives way so the key always fits: lead, padded title, a fill cell, meta.
	title = ansi.Truncate(title, max(1, width-5-ansi.StringWidth(meta)), "…")
	titleStyle := theme.StyleHeader.Bold(true)
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title:      title,
		Meta:       meta,
		TitleStyle: &titleStyle,
		Width:      width,
		Height:     height,
		Lines:      lines,
		Focused:    state.DailyDetailFocus,
		Dimmed:     !state.DailyDetailFocus,
	})
}

func detailBody(theme types.Theme, state types.ContentState, width int) (string, []string) {
	inner := max(1, width-2*panePad)
	if !state.DailyDetailFocus || state.DailyDetailSummary {
		return dayTitle(state), padded(viewsummary.DayLines(theme, state, inner))
	}
	if state.Pane == "habits" {
		if habit, ok := selectedDailyHabit(state); ok {
			return habit.Name, padded(itemdetail.HabitLines(theme, state, habit, inner))
		}
	} else if issue, ok := selectedDailyIssue(state); ok {
		return issue.Title, padded(itemdetail.IssueLines(theme, state, issue, inner))
	}
	return "Selected", []string{theme.StyleDim.Render("Nothing selected")}
}

// dayTitle names the summarised day, "Sun 8 Nov (Today)". Before the summary loads it names
// the dashboard's day.
func dayTitle(state types.ContentState) string {
	iso := state.SummaryDate
	if iso == "" {
		iso = viewruntime.ShownDate(state)
	}
	date, err := time.Parse(time.DateOnly, iso)
	if err != nil {
		return "Day"
	}
	title := date.Format("Mon 2 Jan")
	if state.Health != nil && state.Health.CurrentDate == iso {
		title += " (Today)"
	}
	return title
}

// DetailScrollLimit is the furthest the focused detail pane scrolls: its last line at the top.
func DetailScrollLimit(theme types.Theme, state types.ContentState) int {
	_, lines := detailBody(theme, state, max(1, layoutFor(state).detailW-2))
	return max(0, len(lines)-1)
}

func scrollArrows(up, down bool) string {
	switch {
	case up && down:
		return "↑↓"
	case up:
		return "↑"
	default:
		return "↓"
	}
}

func selectedDailyIssue(state types.ContentState) (api.Issue, bool) {
	indices := dailyIssueIndices(state)
	cur := state.Cursors["issues"]
	if cur < 0 || cur >= len(indices) || indices[cur] >= len(state.DailyIssues) {
		return api.Issue{}, false
	}
	return state.DailyIssues[indices[cur]], true
}

func selectedDailyHabit(state types.ContentState) (api.HabitDailyItem, bool) {
	indices := dailyHabitIndices(state)
	cur := state.Cursors["habits"]
	if cur < 0 || cur >= len(indices) || indices[cur] >= len(state.DueHabits) {
		return api.HabitDailyItem{}, false
	}
	return state.DueHabits[indices[cur]], true
}
