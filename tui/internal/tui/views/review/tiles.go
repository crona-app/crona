package review

import (
	"cmp"
	"math"
	"strings"

	"crona.work/tui/internal/api"
	"crona.work/tui/internal/tui/views/charts"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

const tileMinW = 18

type tile struct {
	label, value, change string
	spark                []float64
	color                lipgloss.Color
}

// headline is a tile of now, with its change on before when the previous window had any.
func headline(theme types.Theme, label string, color lipgloss.Color, format func(float64) string, spark []float64,
	now, before *float64,
) tile {
	t := tile{label: label, value: "–", spark: spark, color: color}
	if now != nil {
		t.value = format(*now)
		if before != nil {
			t.change = change(theme, *now-*before, format)
		}
	}
	return t
}

// change is the difference on the previous window: ▲ more (green), ▼ less (red), = level.
// For every metric here, more is better.
func change(theme types.Theme, delta float64, format func(float64) string) string {
	switch {
	case format(math.Abs(delta)) == format(0):
		return theme.StyleDim.Render("=")
	case delta > 0:
		return lipgloss.NewStyle().Foreground(theme.ColorGreen).Render("▲ " + format(delta))
	default:
		return lipgloss.NewStyle().Foreground(theme.ColorRed).Render("▼ " + format(-delta))
	}
}

func tilesFor(theme types.Theme, state types.ContentState, s daySeries, m metricSet) []tile {
	value := func(v float64) *float64 { return &v }
	var prev types.ReviewPrev
	if state.ReviewPrev != nil {
		prev = *state.ReviewPrev
	}
	rollup := cmp.Or(state.MetricsRollup, &api.MetricsRollup{})
	prev.Rollup = cmp.Or(prev.Rollup, &api.MetricsRollup{})

	var worked, workedBefore *float64
	if state.MetricsRollup != nil {
		worked = value(hours(rollup.WorkedSeconds))
	}
	if prev.Rollup.SessionCount > 0 {
		workedBefore = value(hours(prev.Rollup.WorkedSeconds))
	}
	var focus, focusBefore *float64
	if state.WeeklyFocusScore != nil {
		focus = value(float64(state.WeeklyFocusScore.Score))
	}
	if prev.Focus != nil && prev.Focus.SessionCount > 0 {
		focusBefore = value(float64(prev.Focus.Score))
	}
	var done, doneBefore *float64
	window := state.DashboardWindow
	if window != nil && window.PlannedCount > 0 {
		done = value(float64(window.CompletedCount))
	}
	if prev.Window != nil && prev.Window.PlannedCount > 0 {
		doneBefore = value(float64(prev.Window.CompletedCount))
	}

	ratios := make([]float64, len(s.days))
	for i := range ratios {
		ratios[i] = math.NaN()
		if s.planned[i] > 0 {
			ratios[i] = float64(s.done[i]) / float64(s.planned[i])
		}
	}
	plan := headline(theme, "Plan", theme.ColorGreen, whole, ratios, done, doneBefore)
	if done != nil {
		plan.value = planCount(window.CompletedCount, window.PlannedCount)
	}
	of := func(series metric, spark []float64, now, before *float64) tile {
		return headline(theme, series.title(), series.color, series.format, spark, now, before)
	}
	return []tile{
		of(m.worked, s.worked, worked, workedBefore),
		of(m.focus, s.focus, focus, focusBefore),
		plan,
		of(m.mood, s.mood, rollup.AverageMood, prev.Rollup.AverageMood),
		of(m.energy, s.energy, rollup.AverageEnergy, prev.Rollup.AverageEnergy),
		of(m.sleep, s.sleep, rollup.AverageSleepHours, prev.Rollup.AverageSleepHours),
	}
}

// tileLines lays the tiles out in rows of as many as fit width.
func tileLines(theme types.Theme, tiles []tile, width int) []string {
	perRow := max(1, min(len(tiles), (width+3)/(tileMinW+3)))
	lines := []string{}
	sep := theme.StyleDim.Render(" │ ")
	for start := 0; start < len(tiles); start += perRow {
		row := tiles[start:min(len(tiles), start+perRow)]
		w := (width - 3*(perRow-1)) / perRow
		top, bottom := []string{}, []string{}
		for _, t := range row {
			head := theme.StyleDim.Render(t.label) + " " + theme.StyleHeader.Render(t.value)
			if t.change != "" {
				head += " " + t.change
			}
			top = append(top, fit(head, w))
			bottom = append(bottom, fit(sparkline(t.spark, w, t.color), w))
		}
		lines = append(lines, strings.Join(top, sep), strings.Join(bottom, sep))
	}
	return lines
}

// sparkline draws values in width cells of eighth blocks, averaging days when they outnumber
// the cells; NaN days are blank.
func sparkline(values []float64, width int, color lipgloss.Color) string {
	if len(values) == 0 || width < 1 {
		return ""
	}
	cells := min(width, len(values))
	points := make([]float64, cells)
	top := 0.0
	for c := range cells {
		points[c] = mean(values[c*len(values)/cells : (c+1)*len(values)/cells])
		if !math.IsNaN(points[c]) {
			top = max(top, points[c])
		}
	}
	var b strings.Builder
	for _, v := range points {
		switch {
		case math.IsNaN(v):
			b.WriteString(" ")
		case top <= 0:
			b.WriteString(charts.Block(1))
		default:
			b.WriteString(charts.Block(max(1, int(math.Round(v/top*8)))))
		}
	}
	return lipgloss.NewStyle().Foreground(color).Render(b.String())
}
