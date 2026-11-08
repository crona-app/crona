package momentumview

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona.work/shared/types"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	types "crona.work/tui/internal/tui/views/types"
)

// selectedBucket is the momentum's chosen bucket and its index in the series.
func selectedBucket(card sharedtypes.MomentumCard, offset int) (sharedtypes.MomentumSeriesPoint, int, bool) {
	if len(card.Series) == 0 {
		return sharedtypes.MomentumSeriesPoint{}, -1, false
	}
	i := len(card.Series) - 1 - min(max(0, offset), len(card.Series)-1)
	return card.Series[i], i, true
}

// bucketName is a daily bucket's day, or the bucket's own label.
func bucketName(card sharedtypes.MomentumCard, point sharedtypes.MomentumSeriesPoint) string {
	if sharedtypes.NormalizeHabitStreakPeriod(card.Definition.Period) == sharedtypes.HabitStreakPeriodDay {
		if day, err := time.Parse(isoDate, point.BucketKey); err == nil {
			return day.Format("Mon 2 Jan")
		}
	}
	return momentumDisplayLabel(point)
}

// contributorsTitle names the chosen bucket and how it went, for the pane's title.
func contributorsTitle(theme types.Theme, card sharedtypes.MomentumCard, offset int) (string, string) {
	point, _, ok := selectedBucket(card, offset)
	if !ok {
		return "Contributors", ""
	}
	status := momentumStatusForPoint(point, card.Definition.Enabled)
	return bucketName(card, point), momentumRatioText(point, card.Definition) + " " +
		momentumStatusStyle(theme, status).Render(status)
}

// ContributorLines are what counted toward the chosen bucket, newest day first: each day a
// heading, its sessions or habit completions under it with their amount and context.
func ContributorLines(theme types.Theme, detail *sharedtypes.MomentumDetail, width int) []string {
	if detail == nil {
		return []string{theme.StyleDim.Render("Loading")}
	}
	lines := []string{}
	if len(detail.Contributors) == 0 {
		return append(lines, theme.StyleDim.Render("Nothing counted toward this bucket."))
	}
	date := ""
	for _, c := range detail.Contributors {
		if c.Date != date {
			if date != "" {
				lines = append(lines, "")
			}
			date = c.Date
			heading := date
			if t, err := time.Parse(isoDate, date); err == nil {
				heading = t.Format("Monday 2 Jan")
			}
			lines = append(lines, theme.StyleHeader.Render(heading))
		}
		amount := strings.TrimSpace(c.AmountLabel)
		title := viewhelpers.Truncate(c.Title, max(1, width-4-len(amount)-2))
		gap := max(1, width-2-len([]rune(title))-len([]rune(amount)))
		lines = append(lines, "  "+title+strings.Repeat(" ", gap)+theme.StyleDim.Render(amount))
		meta := []string{}
		for _, part := range []string{c.Context, c.Meta} {
			if strings.TrimSpace(part) != "" {
				meta = append(meta, strings.TrimSpace(part))
			}
		}
		if len(meta) > 0 {
			lines = append(lines, theme.StyleDim.Render(viewhelpers.Truncate("    "+strings.Join(meta, " · "), width)))
		}
	}
	return lines
}

// bucketFootnote is the chosen bucket and the window's numbers, under the chart.
func bucketFootnote(card sharedtypes.MomentumCard, offset int) string {
	def := card.Definition
	point, _, ok := selectedBucket(card, offset)
	if !ok {
		return ""
	}
	parts := []string{}
	if point.Skipped {
		parts = append(parts, fmt.Sprintf("%s skipped (%d protected days)", bucketName(card, point), point.ProtectedDayCount))
	} else {
		parts = append(parts, fmt.Sprintf("%s %s %s", bucketName(card, point), momentumRatioText(point, def),
			momentumStatusForPoint(point, def.Enabled)))
	}
	if notes := bucketNotes(point, def); notes != "" && !point.Skipped {
		parts = append(parts, notes)
	}
	shared, hasShared := momentumSharedTarget(card.Series)
	parts = append(parts, bucketSummary(def, card.Series, shared, hasShared), "[←/→] bucket")
	return strings.Join(parts, " · ")
}

func selectedIndex(card sharedtypes.MomentumCard, offset int) int {
	_, i, _ := selectedBucket(card, offset)
	return i
}

func selectedKey(card sharedtypes.MomentumCard, offset int) string {
	point, _, _ := selectedBucket(card, offset)
	return point.BucketKey
}
