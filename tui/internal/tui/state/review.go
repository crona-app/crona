package state

import (
	"time"

	sharedtypes "crona.work/shared/types"
)

// DefaultReviewWindowDays is the Review window when none has been chosen.
const DefaultReviewWindowDays = 7

// ReviewRange is the Review window: days ending on anchor while date is inside it,
// otherwise days ending on date, so moving the date out of the window carries it along.
func ReviewRange(date, anchor string, days int) (string, string) {
	if days < 1 {
		days = DefaultReviewWindowDays
	}
	end := anchor
	if end == "" || date > end || date < shiftDate(end, -(days-1)) {
		end = date
	}
	return shiftDate(end, -(days - 1)), end
}

func shiftDate(date string, days int) string {
	t, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return date
	}
	return t.AddDate(0, 0, days).Format(time.DateOnly)
}

// ReviewMaxDayColumns is the longest Review window shown a column per day; longer ones
// go by week.
const ReviewMaxDayColumns = 14

// ReviewColumns splits start..end into Review's columns: a day each, or a week each
// (starting on weekStart) when the window is longer than ReviewMaxDayColumns.
func ReviewColumns(start, end string, weekStart sharedtypes.WeekStart) [][]string {
	first, err1 := time.Parse(time.DateOnly, start)
	last, err2 := time.Parse(time.DateOnly, end)
	if err1 != nil || err2 != nil || last.Before(first) {
		return nil
	}
	days := int(last.Sub(first).Hours()/24) + 1
	startDay := time.Monday
	if sharedtypes.NormalizeWeekStart(weekStart) == sharedtypes.WeekStartSunday {
		startDay = time.Sunday
	}
	columns := [][]string{}
	for t := first; !t.After(last); t = t.AddDate(0, 0, 1) {
		if len(columns) == 0 || days <= ReviewMaxDayColumns || t.Weekday() == startDay {
			columns = append(columns, nil)
		}
		columns[len(columns)-1] = append(columns[len(columns)-1], t.Format(time.DateOnly))
	}
	return columns
}
