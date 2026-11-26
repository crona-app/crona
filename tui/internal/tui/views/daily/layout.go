package daily

import (
	uistate "crona.work/tui/internal/tui/state"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"
	viewui "crona.work/tui/internal/tui/views/ui"

	"github.com/charmbracelet/lipgloss"
)

// Render draws Today: the day strip over Tasks and Habits, and on wide screens the calendar
// over a selection preview and the Day summary beside them.
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

// renderView puts Issues above Habits, beside the detail pane on wide screens
// and under the day line otherwise.
func renderView(theme types.Theme, state types.ContentState) string {
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
		if state.HistoricalAwayDate != "" {
			line = ""
		}
		// A blank row parts the line from the panes under it.
		return []string{line, ""}, state.Height - 2
	}
	return nil, state.Height
}

// detailCalendarFits reserves a compact preview and at least eight summary rows.
func detailCalendarFits(height int) func(int) bool {
	return func(h int) bool { return height-h >= max(15, height/2) }
}

// dailyRightColumn keeps the item and the day together. Focus temporarily expands one
// surface without changing saved calendar visibility.
func dailyRightColumn(theme types.Theme, state types.ContentState, width, height int) string {
	if height < 3 {
		return renderDetail(theme, state, width, height)
	}
	if state.DailyDetailFocus {
		calendar, calendarH := focusedCalendar(theme, state, width, height)
		available := height - calendarH
		if state.DailyDetailSummary {
			return lipgloss.JoinVertical(lipgloss.Left, calendar, renderPreview(theme, state, width, 1),
				renderDetail(theme, state, width, available-1))
		}
		summaryH := max(1, min(6, available-dailyMinDetailRows))
		return lipgloss.JoinVertical(lipgloss.Left, calendar,
			renderDetail(theme, state, width, available-summaryH), compactSummary(theme, state, width, summaryH))
	}
	calendar, calendarH := calendarPane(theme, state, width, detailCalendarFits(height))
	available := height - calendarH
	previewH := 7
	if available < 15 {
		previewH = min(4, available-8)
		if previewH < 4 {
			previewH = 1
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, calendar, renderPreview(theme, state, width, previewH),
		renderDetail(theme, state, width, available-previewH))
}

// focusedCalendar keeps the week visible when either side surface expands. Saved hidden
// calendars remain hidden; short terminals use the one-line week instead of boxed days.
func focusedCalendar(theme types.Theme, state types.ContentState, width, height int) (string, int) {
	if height < 6 {
		return collapsedPane(theme, "Calendar", uistate.CollapseDailyCalendar, width), 1
	}
	return weekPane(theme, state, width, height)
}

// dailyLists splits Issues and Habits, reserving only the help rows for an empty list
// when the other contains items. Narrow screens put the week above Issues and use the
// Habits area for focused details.
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
		if !state.DailyDetailFocus && state.HistoricalAwayDate == "" {
			issueCount := len(dailyIssueIndices(state))
			habitCount := len(dailyHabitIndices(state))
			issueRows := height - aboveH - lowerH
			switch {
			case habitCount == 0 && issueCount > max(1, (issueRows-1)/3):
				// Keep the empty heading, creation hint and explanation.
				lowerH = min(lowerH, 3)
			case issueCount == 0 && habitCount > max(1, (lowerH-1)/3):
				// Keep the selector and room for backlog counts and their shortcut.
				lowerH = max(lowerH, height-aboveH-5)
			}
		}
		lower = renderHabits(theme, state, width, lowerH)
	}
	parts := append(above, renderIssues(theme, state, width, height-lowerH-aboveH), lower)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
