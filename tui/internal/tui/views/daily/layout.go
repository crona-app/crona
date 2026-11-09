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

// dailyMinBodyRows is the room Tasks and Habits keep before the strip gives up rows.
const dailyMinBodyRows = 12

// layout is how Today splits the width: wide screens hold the detail pane (and with it the
// Day summary and the calendar) beside the lists.
type layout struct {
	wide           bool
	listW, detailW int
}

func layoutFor(state types.ContentState) layout {
	if viewhelpers.ClassFor(state.Width) != viewhelpers.WidthWide {
		return layout{listW: state.Width}
	}
	cols := viewhelpers.Columns(state.Width, 3, 2)
	return layout{wide: true, listW: cols[0], detailW: cols[1]}
}

// renderView lays Today out as a day strip over Tasks and Habits, each holding half the
// height so neither moves as the other changes, and on wide screens a detail pane.
func renderView(theme types.Theme, state types.ContentState) string {
	if state.HistoricalAwayDate != "" {
		return viewaway.RenderHistorical(theme, state)
	}
	l := layoutFor(state)
	dayLine, strip, rest := dailyChrome(theme, state, l)
	body := dailyLists(theme, state, l.listW, rest, !l.wide)
	if l.wide {
		body = viewchrome.JoinColumns(theme, rest, body, dailyRightColumn(theme, state, l.detailW, rest))
	}
	return lipgloss.JoinVertical(lipgloss.Left, append(append(dayLine, strip...), body)...)
}

// dailyChrome is the day line and strip over Today's lists, and the rows left beneath them.
// The detail pane's Day summary holds the strip's meters and signals, so with that pane on
// screen the strip is left out; without it, the summary shows as one line over the strip.
func dailyChrome(theme types.Theme, state types.ContentState, l layout) (dayLine, strip []string, rest int) {
	height := state.Height
	if l.wide {
		return nil, nil, height
	}
	if height-1-dayStripMaxRows >= dailyMinBodyRows {
		if line := viewsummary.DayLine(theme, state, state.Width); line != "" {
			dayLine, height = []string{line}, height-1
		}
	}
	stripRows := dayStripMaxRows
	for stripRows > 1 && height-stripRows < dailyMinBodyRows {
		stripRows--
	}
	strip = dayStripRows(theme, state, state.Width, stripRows)
	return dayLine, strip, height - len(strip)
}

// calendarSlot is the calendar's width and the test of whether a calendar of a given
// height fits where Today puts it: under the details on wide screens, under Habits otherwise.
func calendarSlot(state types.ContentState) (int, func(int) bool) {
	l := layoutFor(state)
	_, _, rest := dailyChrome(types.Theme{}, state, l)
	if l.wide {
		return l.detailW, detailCalendarFits(rest)
	}
	return l.listW, listsCalendarFits(state, rest)
}

// detailCalendarFits keeps the details at least half the column.
func detailCalendarFits(height int) func(int) bool {
	return func(h int) bool { return height-h >= max(dailyMinDetailRows, height/2) }
}

// listsCalendarFits keeps Habits its rows in the lower half, or, with Habits folded, Tasks
// its rows.
func listsCalendarFits(state types.ContentState, height int) func(int) bool {
	if dailyCollapsed(state, uistate.CollapseDailyHabits) {
		return func(h int) bool { return height-1-h >= dailyMinTaskRows }
	}
	return func(h int) bool { return height/2-h >= dailyMinHabitRows }
}

// dailyRightColumn is the score calendar over the detail pane. The calendar gives way first,
// to its week and then folded, so details keep half the column.
func dailyRightColumn(theme types.Theme, state types.ContentState, width, height int) string {
	calendar, calendarH := calendarPane(theme, state, width, detailCalendarFits(height))
	return lipgloss.JoinVertical(lipgloss.Left, calendar, renderDetail(theme, state, width, height-calendarH))
}

// dailyLists is Tasks over Habits. Tasks keeps the top half and Habits the bottom half;
// folded, Habits keeps its title row and Tasks takes the rest. withCalendar puts the
// calendar under Habits, where it shares Habits' half while Habits keeps dailyMinHabitRows.
func dailyLists(theme types.Theme, state types.ContentState, width, height int, withCalendar bool) string {
	var below []string
	belowH := 0
	if withCalendar {
		calendar, calendarH := calendarPane(theme, state, width, listsCalendarFits(state, height))
		below, belowH = []string{calendar}, calendarH
	}
	habits, habitsH := collapsedPane(theme, "Habits", uistate.CollapseDailyHabits, width), 1
	if !dailyCollapsed(state, uistate.CollapseDailyHabits) {
		habitsH = max(1, height/2-belowH)
		habits = renderHabits(theme, state, width, habitsH)
	}
	parts := append([]string{renderIssues(theme, state, width, height-habitsH-belowH), habits}, below...)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
