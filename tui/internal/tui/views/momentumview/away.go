package momentumview

import (
	"fmt"
	"time"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
)

func awayDateSet(settings *api.CoreSettings) map[string]bool {
	dates := map[string]bool{}
	if settings != nil {
		for _, date := range settings.AwayDates {
			if _, err := time.Parse(isoDate, date); err == nil {
				dates[date] = true
			}
		}
	}
	return dates
}

func awayDayCount(point sharedtypes.MomentumSeriesPoint, dates map[string]bool) int {
	start, end := point.StartDate, point.EndDate
	if start == "" || end == "" {
		start, end = point.BucketKey, point.BucketKey
	}
	if _, err := time.Parse(isoDate, start); err != nil {
		return 0
	}
	if _, err := time.Parse(isoDate, end); err != nil {
		return 0
	}
	count := 0
	for date, away := range dates {
		if away && date >= start && date <= end {
			count++
		}
	}
	return count
}

func awayBucketLabel(count int) string {
	if count == 1 {
		return "1 away day"
	}
	return fmt.Sprintf("%d away days", count)
}
