package daily

import (
	viewcalendar "crona.work/tui/internal/tui/views/calendar"
	viewruntime "crona.work/tui/internal/tui/views/runtime"
	types "crona.work/tui/internal/tui/views/types"
)

// SubTabMeta is Daily's date navigation for the date row: the shown date and its week,
// flanked by the date-navigation keys. Within room it drops the key words first, then the
// keys, then the week, so the date always shows.
func SubTabMeta(theme types.Theme, state types.ContentState, room int) string {
	return viewcalendar.DateNav(theme, viewruntime.ShownDate(state), "", state.Settings, state.WeekStart, room)
}
