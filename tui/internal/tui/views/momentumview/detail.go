package momentumview

import (
	"fmt"
	"strings"

	sharedtypes "crona.work/shared/types"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	momentumhelpers "crona.work/tui/internal/tui/views/momentum"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

const (
	// pairedHeaderWidth is the narrowest detail that sets the settings side by side.
	pairedHeaderWidth = 60
	// minChartRows is the shortest chart the medium layout keeps above its buckets.
	minChartRows = 8
	isoDate      = "2006-01-02"
)

// DetailLines is the momentum's settings, then its chart, which fills the room left; with
// stacked lines (the contributors on a medium screen) the chart takes 55% and they the rest.
func DetailLines(
	theme types.Theme,
	card sharedtypes.MomentumCard,
	width, height, bucket int,
	stacked []string,
) []string {
	lines := append(headerLines(theme, card, width), "")
	room := height - len(lines)
	chartRows := room
	if stacked != nil {
		chartRows = max(minChartRows, room*55/100)
		if room-chartRows < 3 {
			chartRows, stacked = room, nil
		}
	}
	lines = append(lines, chartLines(theme, card, width, chartRows, bucket)...)
	if stacked != nil {
		lines = append(append(lines, ""), stacked...)
	}
	return lines[:min(len(lines), max(0, height))]
}

func headerLines(theme types.Theme, card sharedtypes.MomentumCard, width int) []string {
	def := card.Definition
	status := latestStatus(card)
	statusText := "–"
	if status != "" {
		statusText = momentumStatusStyle(theme, status).Render(status)
	}
	left := [][2]string{
		{"Cadence", momentumhelpers.CadenceLabel(def.Period)},
		{"Target", momentumTargetSummary(def)},
		{
			"Streak",
			streakLabel(card, card.Current) + " now · best " + streakLabel(card, card.Longest),
		},
	}
	right := [][2]string{
		{momentumCardTargetLabel(def), momentumTargetSummaryNames(card)},
		{"Status", statusText},
	}
	lines := []string{}
	if width < pairedHeaderWidth {
		for _, row := range append(left, right...) {
			lines = append(lines, viewsummary.Row(theme, row[0], row[1], width))
		}
	} else {
		// The left column takes what its values need, up to 3/5, and the right the rest.
		leftWidth := 0
		for _, row := range left {
			leftWidth = max(leftWidth, ansi.StringWidth(viewsummary.Row(theme, row[0], row[1], width)))
		}
		leftWidth = min(leftWidth, width*3/5)
		for i, row := range left {
			line := viewsummary.Row(theme, row[0], row[1], leftWidth)
			if i < len(right) {
				line += strings.Repeat(" ", max(0, leftWidth-ansi.StringWidth(line))+2) + viewsummary.Row(theme, right[i][0], right[i][1], width-leftWidth-2)
			}
			lines = append(lines, line)
		}
	}
	if def.Description != nil && strings.TrimSpace(*def.Description) != "" {
		for _, line := range strings.Split(ansi.Wordwrap(strings.TrimSpace(*def.Description), max(1, width), ""), "\n") {
			lines = append(lines, theme.StyleDim.Render(line))
		}
	}
	return lines
}

// chartLines is the chart and the latest-bucket footnote at exactly height lines.
func chartLines(theme types.Theme, card sharedtypes.MomentumCard, width, height, bucket int) []string {
	if height <= 0 {
		return nil
	}
	def := card.Definition
	footnote := theme.StyleDim.Render(
		viewhelpers.Truncate(bucketFootnote(card, bucket), width),
	)
	rows := height - 1
	var chart []string
	switch {
	case rows <= 0:
	case sharedtypes.NormalizeHabitStreakPeriod(def.Period) == sharedtypes.HabitStreakPeriodDay && momentumDailySeriesUsesSquares(def, card.Series):
		chart = calendarLines(
			theme,
			card.Series,
			width,
			rows,
			def.Enabled,
			selectedKey(card, bucket),
		)
	default:
		chart = renderMomentumBars(theme, def, card.Series, width, rows, def.Enabled, selectedIndex(card, bucket))
	}
	lines := make([]string, 0, height)
	for _, line := range chart[:min(len(chart), max(0, rows))] {
		lines = append(lines, ansi.Truncate(line, width, ""))
	}
	for len(lines) < rows {
		lines = append(lines, "")
	}
	return append(lines, footnote)
}

// bucketSummary is the table's opening line: the shared target, buckets met, and the mean.
func bucketSummary(
	def sharedtypes.HabitStreakDefinition,
	series []sharedtypes.MomentumSeriesPoint,
	shared int,
	hasShared bool,
) string {
	met, counted, total := 0, 0, 0
	for _, point := range series {
		if point.Skipped {
			continue
		}
		counted++
		total += point.Count
		if momentumStatusForPoint(point, true) == "met" {
			met++
		}
	}
	parts := []string{}
	if !def.Enabled {
		parts = append(parts, "paused")
	}
	if hasShared {
		parts = append(parts, "target "+momentumValueDisplay(shared, def))
	}
	if def.Enabled {
		parts = append(parts, fmt.Sprintf("met %d of %d", met, counted))
	}
	if counted > 0 {
		parts = append(parts, "avg "+momentumValueDisplay(total/counted, def))
	}
	return strings.Join(parts, " · ")
}

func bucketNotes(
	point sharedtypes.MomentumSeriesPoint,
	def sharedtypes.HabitStreakDefinition,
) string {
	notes := []string{}
	if point.ProtectedDayCount > 0 {
		notes = append(notes, fmt.Sprintf("protected %dd", point.ProtectedDayCount))
	}
	if point.OriginalTarget > 0 && point.OriginalTarget != point.Target {
		notes = append(notes, "prorated from "+momentumValueDisplay(point.OriginalTarget, def))
	}
	return strings.Join(notes, " · ")
}
