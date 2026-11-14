package input

import (
	"maps"

	uistate "crona.work/tui/internal/tui/state"
	tea "github.com/charmbracelet/bubbletea"
)

// handleDailyDetailBrowse keeps arrows inside the expanded item's visible list.
func handleDailyDetailBrowse(s State, key tea.KeyMsg, deps Deps) (State, bool) {
	if s.ActiveView != uistate.ViewDaily || !s.DailyDetailFocus || s.DailyDetailSummary ||
		s.Dialog != "" || s.HelpOpen {
		return s, false
	}
	delta := 0
	switch key.String() {
	case "left":
		delta = -1
	case "right":
		delta = 1
	default:
		return s, false
	}
	total := deps.ListLen(s, s.ActivePane)
	if total == 0 {
		s.DailyDetailFocus, s.DailyDetailScroll = false, 0
		return s, true
	}
	cursor := max(0, min(total-1, s.Cursor[s.ActivePane]+delta))
	if cursor != s.Cursor[s.ActivePane] {
		s.Cursor = maps.Clone(s.Cursor)
		if s.Cursor == nil {
			s.Cursor = map[uistate.Pane]int{}
		}
		s.Cursor[s.ActivePane], s.DailyDetailScroll = cursor, 0
	}
	return s, true
}
