package daily

import (
	uistate "crona.work/tui/internal/tui/state"
	viewaway "crona.work/tui/internal/tui/views/away"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"
	viewui "crona.work/tui/internal/tui/views/ui"

	"github.com/charmbracelet/lipgloss"
)

// Render draws Today: the day strip over Tasks and Habits, and on wide screens the calendar
// over the detail pane beside them.
func Render(theme types.Theme, state types.ContentState) string {
	return viewui.NewLayout(theme, state, renderView).RenderView()
}

// layout is how Today splits the width: wide screens hold the detail pane (and with it the
// Day summary and the month calendar) beside the lists; narrower ones stack it in the lower
// half while it has focus, under the day line and over the week.
type layout struct {
	wide           bool
	listW, detailW int
}

func layoutFor(state types.ContentState) layout {
	if viewhelpers.ClassFor(state.Width) != viewhelpers.WidthWide {
		return layout{listW: state.Width, detailW: state.Width}
	}
	cols := viewhelpers.Columns(state.Width, 3, 2)
	return layout{wide: true, listW: cols[0], detailW: cols[1]}
}

// renderView lays Today out as Tasks over Habits, each holding half the height so neither
// moves as the other changes: beside the detail pane on wide screens, under the day line
// otherwise.
func renderView(theme types.Theme, state types.ContentState) string {
	if state.HistoricalAwayDate != "" {
		return viewaway.RenderHistorical(theme, state)
	}
	l := layoutFor(state)
	dayLine, rest := dailyChrome(theme, state, l)
	body := dailyLists(theme, state, l.listW, rest, !l.wide)
	if l.wide {
		body = viewchrome.JoinColumns(theme, rest, body, dailyRightColumn(theme, state, l.detailW, rest))
	}
	return lipgloss.JoinVertical(lipgloss.Left, append(dayLine, body)...)
}

// dailyChrome is the day line over Today's lists, and the rows left beneath it. With the
// detail pane on screen its Day summary says it all, so there is no line.
func dailyChrome(theme types.Theme, state types.ContentState, l layout) (dayLine []string, rest int) {
	if l.wide {
		return nil, state.Height
	}
	if line := viewsummary.DayLine(theme, state, state.Width); line != "" {
		// A blank row parts the line from the panes under it.
		return []string{line, ""}, state.Height - 2
	}
	return nil, state.Height
}

// detailCalendarFits keeps the details at least half the column.
func detailCalendarFits(height int) func(int) bool {
	return func(h int) bool { return height-h >= max(dailyMinDetailRows, height/2) }
}

// dailyRightColumn is the score calendar over the detail pane. The calendar gives way first,
// to its week and then folded, so details keep half the column.
func dailyRightColumn(theme types.Theme, state types.ContentState, width, height int) string {
	calendar, calendarH := calendarPane(theme, state, width, detailCalendarFits(height))
	return lipgloss.JoinVertical(lipgloss.Left, calendar, renderDetail(theme, state, width, height-calendarH))
}

// dailyLists is Tasks over Habits. Tasks keeps the top half and Habits the bottom half;
// folded, Habits keeps its title row and Tasks takes the rest. withCalendar (narrow screens)
// puts the week over Tasks, and gives the focused detail pane Habits' half.
func dailyLists(theme types.Theme, state types.ContentState, width, height int, withCalendar bool) string {
	var above []string
	aboveH := 0
	if withCalendar {
		calendar, calendarH := weekPane(theme, state, width, height)
		above, aboveH = []string{calendar}, calendarH
	}
	var lower string
	lowerH := max(1, (height-aboveH)/2)
	switch {
	case withCalendar && state.DailyDetailFocus:
		lower = renderDetail(theme, state, width, lowerH)
	case dailyCollapsed(state, uistate.CollapseDailyHabits):
		lower, lowerH = collapsedPane(theme, "Habits", uistate.CollapseDailyHabits, width, summaryHint(theme, state)...), 1
	default:
		lower = renderHabits(theme, state, width, lowerH)
	}
	parts := append(above, renderIssues(theme, state, width, height-lowerH-aboveH), lower)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
