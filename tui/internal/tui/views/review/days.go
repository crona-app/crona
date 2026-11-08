package review

import (
	"time"

	types "crona.work/tui/internal/tui/views/types"
)

const isoDate = "2006-01-02"

// WindowDates lists the window's ISO dates, oldest first.
func WindowDates(state types.ContentState) []string {
	end := state.ReviewWindowEnd
	if end == "" {
		end = state.DashboardDate
	}
	last, err := time.Parse(isoDate, end)
	if err != nil {
		return nil
	}
	days := state.ReviewWindowDays
	if days < 1 {
		days = 7
	}
	dates := make([]string, days)
	for i := range days {
		dates[i] = last.AddDate(0, 0, i-days+1).Format(isoDate)
	}
	return dates
}
