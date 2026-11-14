package itemdetail

import (
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	types "crona.work/tui/internal/tui/views/types"
)

// detailKey attaches an action to its data only in Daily's expanded item details.
func detailKey(theme types.Theme, state types.ContentState, key, label string) string {
	if state.View != "daily" || !state.DailyDetailFocus || state.DailyDetailSummary {
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
