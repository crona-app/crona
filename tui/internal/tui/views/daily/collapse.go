package daily

import (
	"strings"

	uistate "crona.work/tui/internal/tui/state"
	viewcalendar "crona.work/tui/internal/tui/views/calendar"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewruntime "crona.work/tui/internal/tui/views/runtime"
	types "crona.work/tui/internal/tui/views/types"
)

// The calendar shrinks to its week, then folds, before these panes drop below their rows.
const (
	dailyMinDetailRows = 8
)

func dailyCollapsed(state types.ContentState, pane string) bool {
	if state.Settings == nil {
		return false
	}
	return uistate.PaneCollapsed(state.Settings.CollapsedPanes, pane)
}

// foldHint is a pane header's `[⌥1] fold` (or `expand` when folded).
func foldHint(theme types.Theme, pane, action string) string {
	return viewchrome.KeyHint(theme, uistate.KeyLabel(uistate.FoldKey(pane)), action)
}

// collapsedPane is a folded pane: its title row alone, with the key that opens it.
func collapsedPane(theme types.Theme, title, pane string, width int, keys ...string) string {
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title:  title + " ▸",
		Meta:   strings.Join(append(keys, foldHint(theme, pane, "expand")), "  "),
		Dimmed: true,
		Width:  width,
		Height: 1,
	})
}

// scoreCalendarLines is the Summary Calendar for the shown date, boxed as the C dialog draws
// it; week keeps only the shown week.
func scoreCalendarLines(theme types.Theme, state types.ContentState, width int, week bool) []string {
	lines := viewcalendar.ScoreGrid(theme, scoreSelection(state), max(1, width-2*panePad), week)
	if len(lines) == 0 {
		return nil
	}
	// Blank rows part the month caption from the grid, and the grid from the pane below.
	return append(padded(append([]string{lines[0], ""}, lines[1:]...)), "")
}

func scoreSelection(state types.ContentState) viewcalendar.ScoreSelection {
	selection := viewcalendar.ScoreSelection{AnchorDate: viewruntime.ShownDate(state)}
	if state.Health != nil {
		selection.Today = state.Health.CurrentDate
	}
	if state.Settings != nil {
		selection.WeekStart = state.Settings.WeekStart
		selection.AwayDates = state.Settings.AwayDates
	}
	if state.SummarySnapshot != nil {
		selection.Scores = state.SummarySnapshot.CalendarScores
	}
	return selection
}

// panePad is the inset the calendar and Day summary keep from their pane's sides.
const panePad = 1

// padded insets lines by panePad and opens them with a blank row, so the content breathes
// under the pane's title rule.
func padded(lines []string) []string {
	out := make([]string, 0, len(lines)+1)
	out = append(out, "")
	for _, line := range lines {
		if line == "" {
			out = append(out, line)
			continue
		}
		out = append(out, strings.Repeat(" ", panePad)+line)
	}
	return out
}

// calendarPane is the score calendar at the largest size fits allows: the month, then the
// shown week, else folded to its title row.
func calendarPane(
	theme types.Theme,
	state types.ContentState,
	width int,
	fits func(calendarH int) bool,
) (string, int) {
	mode := calendarMode(state)
	if mode == uistate.CalendarHidden {
		return collapsedPane(theme, "Calendar", uistate.CollapseDailyCalendar, width), 1
	}
	shapes := []bool{false, true}
	if mode == uistate.CalendarWeek {
		shapes = []bool{true}
	}
	for _, week := range shapes {
		lines := scoreCalendarLines(theme, state, width-2, week)
		height := len(lines) + 1
		if len(lines) > 0 && fits(height) {
			// The fold key steps a month to its week, and a week to hidden.
			next := "week"
			if week {
				next = "hide"
			}
			// The calendar never takes focus, so it is not faded like an unfocused list.
			return viewchrome.RenderSection(theme, viewchrome.Section{
				Title:  "Calendar",
				Meta:   foldHint(theme, uistate.CollapseDailyCalendar, next),
				Width:  width,
				Height: height,
				Lines:  lines,
			}), height
		}
	}
	return collapsedPane(theme, "Calendar", uistate.CollapseDailyCalendar, width), 1
}

func calendarMode(state types.ContentState) uistate.CalendarMode {
	if state.Settings == nil {
		return uistate.CalendarMonth
	}
	return uistate.DailyCalendarMode(state.Settings.CollapsedPanes)
}

// CalendarMonthFits reports whether Today has room for the calendar's full month, so the
// fold key can skip a month that would only show as its week. Only the wide layout has a
// month; narrower ones show the week on a line.
func CalendarMonthFits(theme types.Theme, state types.ContentState) bool {
	l := layoutFor(state)
	if state.HistoricalAwayDate != "" || !l.wide {
		return false
	}
	_, rest := dailyChrome(theme, state, l)
	lines := scoreCalendarLines(theme, state, l.detailW-2, false)
	return len(lines) > 0 && detailCalendarFits(rest)(len(lines)+1)
}

// weekPane is the shown week over the lists, each day with its focus score: boxed as on wide
// screens while the lists keep narrowListRows each, else on one line. C opens the month.
// Folded, it keeps its title row.
func weekPane(theme types.Theme, state types.ContentState, width, height int) (string, int) {
	if calendarMode(state) == uistate.CalendarHidden {
		return collapsedPane(theme, "Week", uistate.CollapseDailyCalendar, width), 1
	}
	lines := scoreCalendarLines(theme, state, width-2, true)
	if len(lines) == 0 || height-(len(lines)+1) < 2*narrowListRows {
		lines = []string{"", " " + viewcalendar.ScoreWeek(theme, scoreSelection(state)), ""}
	}
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title:  "Week",
		Meta:   viewchrome.KeyHint(theme, "C", "month") + "  " + foldHint(theme, uistate.CollapseDailyCalendar, "hide"),
		Width:  width,
		Height: len(lines) + 1,
		Lines:  lines,
	}), len(lines) + 1
}

// narrowListRows is the room Tasks and Habits each keep before the boxed week gives way to
// its line.
const narrowListRows = 12
