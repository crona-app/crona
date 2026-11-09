package review

import (
	"fmt"
	"math"
	"strings"
	"time"

	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	"crona.work/tui/internal/tui/views/charts"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Review is a dashboard of the window: headline tiles, wellbeing and work as line charts on
// the window's days with the selected day's readout under them, and beside them the risk,
// where the work went and how the plan held.
const (
	// sideMinWidth and sideMaxWidth bound the side column; under dashboardSideAt it stacks.
	sideMinWidth    = 44
	sideMaxWidth    = 64
	dashboardSideAt = 110
	// axisWidth holds a chart's y labels and its axis line.
	axisWidth = 5
	tileMinW  = 18
)

// series is the window's values for each day, oldest first; NaN where a day has none.
type daySeries struct {
	dates                              []string
	mood, energy, sleep, worked, focus []float64
	done, planned                      []int
	selected                           int
}

func collectSeries(state types.ContentState) daySeries {
	d := indexWindow(state)
	s := daySeries{dates: WindowDates(state), selected: -1}
	for i, date := range s.dates {
		if date == state.DashboardDate {
			s.selected = i
		}
		day := d.metrics[date]
		nan := math.NaN()
		mood, energy, sleep := nan, nan, nan
		if c := day.CheckIn; c != nil {
			mood, energy = float64(c.Mood), float64(c.Energy)
			if c.SleepHours != nil {
				sleep = *c.SleepHours
			}
		}
		focus := nan
		if score, ok := d.scores[date]; ok && score.HasData {
			focus = float64(score.Score)
		}
		s.mood, s.energy, s.sleep = append(s.mood, mood), append(s.energy, energy), append(s.sleep, sleep)
		s.focus = append(s.focus, focus)
		s.worked = append(s.worked, float64(day.WorkedSeconds)/3600)
		s.done = append(s.done, d.plans[date].CompletedCount)
		s.planned = append(s.planned, d.plans[date].PlannedCount)
	}
	return s
}

// rule is a section heading across width: title, then meta at the right.
func rule(theme types.Theme, title, meta string, width int) string {
	left := theme.StyleDim.Render("─ ") + theme.StylePaneTitle.Render(title) + " "
	right := ""
	if meta != "" {
		right = " " + meta + theme.StyleDim.Render(" ─")
	}
	fill := width - ansi.StringWidth(left) - ansi.StringWidth(right)
	if fill < 1 {
		return viewhelpers.TruncateANSI(left+meta, width)
	}
	return left + theme.StyleDim.Render(strings.Repeat("─", fill)) + right
}

func fit(line string, width int) string {
	line = viewhelpers.TruncateANSI(line, width)
	return line + strings.Repeat(" ", max(0, width-ansi.StringWidth(line)))
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
		lo, hi := c*len(values)/cells, (c+1)*len(values)/cells
		sum, n := 0.0, 0
		for _, v := range values[lo:hi] {
			if !math.IsNaN(v) {
				sum, n = sum+v, n+1
			}
		}
		points[c] = math.NaN()
		if n > 0 {
			points[c] = sum / float64(n)
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

// tile is a headline: label, value and change on the line, a sparkline of the days under it.
type tile struct {
	label, value, change string
	spark                []float64
	color                lipgloss.Color
}

func tilesFor(theme types.Theme, state types.ContentState, s daySeries) []tile {
	d := indexWindow(state)
	rollup := state.MetricsRollup
	prev := d.prev.Rollup
	tiles := []tile{}
	worked := tile{label: "Worked", value: "–", spark: s.worked, color: theme.ColorCyan}
	if rollup != nil {
		worked.value = viewhelpers.FormatCompactDurationSeconds(rollup.WorkedSeconds)
		if prev != nil && prev.SessionCount > 0 {
			worked.change = change(theme, float64(rollup.WorkedSeconds-prev.WorkedSeconds), func(v float64) string {
				return viewhelpers.FormatCompactDurationSeconds(int(v))
			})
		}
	}
	focus := tile{label: "Focus", value: "–", spark: s.focus, color: theme.ColorLavender}
	if state.WeeklyFocusScore != nil {
		focus.value = fmt.Sprint(state.WeeklyFocusScore.Score)
		if d.prev.Focus != nil && d.prev.Focus.SessionCount > 0 {
			focus.change = change(theme, float64(state.WeeklyFocusScore.Score-d.prev.Focus.Score), func(v float64) string {
				return fmt.Sprintf("%.0f", v)
			})
		}
	}
	ratios := make([]float64, len(s.dates))
	for i := range ratios {
		ratios[i] = math.NaN()
		if s.planned[i] > 0 {
			ratios[i] = float64(s.done[i]) / float64(s.planned[i])
		}
	}
	plan := tile{label: "Plan", value: "–", spark: ratios, color: theme.ColorGreen}
	if window := state.DashboardWindow; window != nil && window.PlannedCount > 0 {
		plan.value = fmt.Sprintf("%d/%d", window.CompletedCount, window.PlannedCount)
		if d.prev.Window != nil && d.prev.Window.PlannedCount > 0 {
			plan.change = change(theme, float64(window.CompletedCount-d.prev.Window.CompletedCount), func(v float64) string {
				return fmt.Sprintf("%.0f", v)
			})
		}
	}
	tiles = append(tiles, worked, focus, plan)
	scale := func(label string, color lipgloss.Color, values []float64, pick func(*api.MetricsRollup) *float64,
		format func(float64) string,
	) tile {
		t := tile{label: label, value: "–", spark: values, color: color}
		if rollup != nil && pick(rollup) != nil {
			t.value = format(*pick(rollup))
			if prev != nil && pick(prev) != nil {
				t.change = change(theme, *pick(rollup)-*pick(prev), format)
			}
		}
		return t
	}
	oneDecimal := func(v float64) string { return fmt.Sprintf("%.1f", v) }
	return append(tiles,
		scale("Mood", theme.ColorMagenta, s.mood, func(r *api.MetricsRollup) *float64 { return r.AverageMood }, oneDecimal),
		scale("Energy", theme.ColorYellow, s.energy, func(r *api.MetricsRollup) *float64 { return r.AverageEnergy }, oneDecimal),
		scale("Sleep", theme.ColorBlue, s.sleep, func(r *api.MetricsRollup) *float64 { return r.AverageSleepHours },
			helperpkg.FormatCompactDurationHours),
	)
}

// tileLines lays the tiles out in rows of as many as fit width, two lines per row.
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

// wellbeingLane is one series of the Wellbeing chart, on its own scale.
type wellbeingLane struct {
	name    string
	values  []float64
	lo, hi  float64
	loLabel string
	hiLabel string
	color   lipgloss.Color
	format  func(float64) string
}

// wellbeingChart draws mood, energy and sleep as lanes one above the other, each on its own
// scale with its window average dotted across, so the lines never cross. Over 14 days the
// line is a 3-day rolling average and the days themselves are faint dots.
func wellbeingChart(theme types.Theme, s daySeries, width, rows int) []string {
	oneDecimal := func(v float64) string { return fmt.Sprintf("%.1f", v) }
	lanes := []wellbeingLane{
		{"Mood", s.mood, 1, 5, "1", "5", theme.ColorMagenta, oneDecimal},
		{"Energy", s.energy, 1, 5, "1", "5", theme.ColorYellow, oneDecimal},
		{"Sleep", s.sleep, 4, 9, "4h", "9h", theme.ColorBlue, helperpkg.FormatCompactDurationHours},
	}
	laneRows := max(2, (rows-len(lanes))/len(lanes))
	lines := make([]string, 0, rows)
	for _, lane := range lanes {
		lines = append(lines, laneHeading(theme, lane, s))
		lines = append(lines, laneChart(theme, lane, s, width, laneRows)...)
	}
	return lines
}

// laneHeading names a lane with its window average and the selected day's value.
func laneHeading(theme types.Theme, lane wellbeingLane, s daySeries) string {
	parts := []string{lipgloss.NewStyle().Foreground(lane.color).Bold(true).Render(lane.name)}
	if avg := mean(lane.values); !math.IsNaN(avg) {
		parts = append(parts, theme.StyleDim.Render("avg ")+lane.format(avg))
	} else {
		parts = append(parts, theme.StyleDim.Render("none logged"))
	}
	if s.selected >= 0 && s.selected < len(lane.values) && !math.IsNaN(lane.values[s.selected]) {
		parts = append(parts, theme.StyleDim.Render(shortDay(s.dates[s.selected])+" ")+lane.format(lane.values[s.selected]))
	}
	return strings.Repeat(" ", axisWidth) + strings.Join(parts, theme.StyleDim.Render(" · "))
}

func laneChart(theme types.Theme, lane wellbeingLane, s daySeries, width, rows int) []string {
	plotW := max(1, width-axisWidth)
	canvas := charts.NewLineCanvas(plotW, rows)
	if len(lane.values) > 14 {
		canvas.Dots(lane.values, lane.lo, lane.hi, 1)
		canvas.Plot(rolling(lane.values, 1), lane.lo, lane.hi, 0)
	} else {
		canvas.Plot(lane.values, lane.lo, lane.hi, 0)
	}
	avgRow := -1
	if avg := mean(lane.values); !math.IsNaN(avg) {
		avgRow = canvas.Row(avg, lane.lo, lane.hi)
	}
	line := lipgloss.NewStyle().Foreground(lane.color)
	faint := line.Faint(true)
	labels := map[int]string{0: lane.hiLabel, rows - 1: lane.loLabel}
	return drawCanvas(theme, canvas, s, rows, labels, func(row, cell int) (string, bool) {
		if r := canvas.Rune(row, cell); r != 0 {
			if canvas.Owner[row][cell] == 1 {
				return faint.Render(string(r)), true
			}
			return line.Render(string(r)), true
		}
		if row == avgRow {
			return theme.StyleDim.Render("┄"), true
		}
		return "", false
	})
}

// mean is the average of the values that are set, NaN when none are.
func mean(values []float64) float64 {
	sum, n := 0.0, 0
	for _, v := range values {
		if !math.IsNaN(v) {
			sum, n = sum+v, n+1
		}
	}
	if n == 0 {
		return math.NaN()
	}
	return sum / float64(n)
}

// rolling averages each day with the radius days either side of it, gaps left out; a day
// with nothing in reach stays NaN.
func rolling(values []float64, radius int) []float64 {
	out := make([]float64, len(values))
	for i := range values {
		out[i] = mean(values[max(0, i-radius):min(len(values), i+radius+1)])
	}
	return out
}

func shortDay(date string) string {
	if t, err := time.Parse(isoDate, date); err == nil {
		return t.Format("Mon 2")
	}
	return date
}

// workChart draws the hours worked each day as bars, and the focus score (0–100 on the same
// height) as a line over them.
func workChart(theme types.Theme, s daySeries, width, rows int) []string {
	plotW := max(1, width-axisWidth)
	canvas := charts.NewLineCanvas(plotW, rows)
	canvas.Plot(s.focus, 0, 100, 0)
	top := 1.0
	for _, h := range s.worked {
		top = max(top, math.Ceil(h))
	}
	n := len(s.dates)
	barW := max(1, min(3, plotW/max(1, n)/2))
	bars := map[int]float64{}
	for i, h := range s.worked {
		centre := canvas.Cell(i, n)
		for x := centre - (barW-1)/2; x < centre-(barW-1)/2+barW; x++ {
			if x >= 0 && x < plotW {
				bars[x] = h
			}
		}
	}
	barStyle := lipgloss.NewStyle().Foreground(theme.ColorCyan)
	lineStyle := lipgloss.NewStyle().Foreground(theme.ColorLavender)
	labels := map[int]string{0: fmt.Sprintf("%.0fh", top), rows - 1: "0"}
	return drawCanvas(theme, canvas, s, rows, labels, func(row, cell int) (string, bool) {
		if r := canvas.Rune(row, cell); r != 0 {
			return lineStyle.Render(string(r)), true
		}
		h, ok := bars[cell]
		if !ok || h <= 0 {
			return "", false
		}
		eighths := int(math.Round(h / top * float64(rows*8)))
		fill := eighths - (rows-1-row)*8
		if fill <= 0 {
			return "", false
		}
		return barStyle.Render(charts.Block(min(8, fill))), true
	})
}

// drawCanvas renders rows of a chart: the y label and axis, then each cell from draw, or the
// selected day's dotted cursor, or blank.
func drawCanvas(theme types.Theme, canvas *charts.LineCanvas, s daySeries, rows int, labels map[int]string,
	draw func(row, cell int) (string, bool),
) []string {
	cursor := -1
	if s.selected >= 0 {
		cursor = canvas.Cell(s.selected, len(s.dates))
	}
	lines := make([]string, 0, rows)
	for row := range rows {
		axis := "│"
		if _, ok := labels[row]; ok {
			axis = "┤"
		}
		var b strings.Builder
		b.WriteString(theme.StyleDim.Render(fmt.Sprintf("%*s %s", axisWidth-2, labels[row], axis)))
		for cell := range canvas.Width {
			if text, ok := draw(row, cell); ok {
				b.WriteString(text)
			} else if cell == cursor {
				b.WriteString(theme.StyleDim.Render("┊"))
			} else {
				b.WriteString(" ")
			}
		}
		lines = append(lines, b.String())
	}
	return lines
}

// dateAxis labels the days under the charts: the first, the last, the selected (▸) and
// evenly between where they fit.
func dateAxis(theme types.Theme, s daySeries, width int) string {
	plotW := max(1, width-axisWidth)
	canvas := charts.NewLineCanvas(plotW, 1)
	n := len(s.dates)
	row := []rune(strings.Repeat(" ", plotW))
	taken := make([]bool, plotW)
	format := "Mon 2"
	if n > 14 {
		format = "Jan 2"
	}
	place := func(i int, mark bool) {
		t, err := time.Parse(isoDate, s.dates[i])
		if err != nil {
			return
		}
		label := []rune(t.Format(format))
		if mark {
			label = append([]rune("▸"), label...)
		}
		start := min(max(0, canvas.Cell(i, n)-len(label)/2), plotW-len(label))
		if start < 0 {
			return
		}
		for x := max(0, start-1); x < min(plotW, start+len(label)+1); x++ {
			if taken[x] {
				return
			}
		}
		for j, r := range label {
			row[start+j], taken[start+j] = r, true
		}
	}
	if s.selected >= 0 {
		place(s.selected, true)
	}
	if n > 0 {
		place(0, false)
		place(n-1, false)
		step := max(1, n/max(1, plotW/12))
		for i := step; i < n-1; i += step {
			place(i, false)
		}
	}
	return theme.StyleDim.Render(strings.Repeat(" ", axisWidth) + string(row))
}

// readout is the selected day's numbers.
func readout(theme types.Theme, s daySeries) string {
	if s.selected < 0 {
		return ""
	}
	i := s.selected
	parts := []string{theme.StyleHeader.Render("▸ " + dayTitle(s.dates[i]))}
	number := func(label string, v float64, format func(float64) string) {
		if !math.IsNaN(v) {
			parts = append(parts, theme.StyleDim.Render(label+" ")+format(v))
		}
	}
	whole := func(v float64) string { return fmt.Sprintf("%.0f", v) }
	number("mood", s.mood[i], whole)
	number("energy", s.energy[i], whole)
	number("sleep", s.sleep[i], helperpkg.FormatCompactDurationHours)
	number("worked", s.worked[i], func(v float64) string { return viewhelpers.FormatCompactDurationSeconds(int(v * 3600)) })
	number("focus", s.focus[i], whole)
	if s.planned[i] > 0 {
		parts = append(parts, theme.StyleDim.Render("plan ")+fmt.Sprintf("%d/%d", s.done[i], s.planned[i]))
	}
	return strings.Join(parts, theme.StyleDim.Render(" · ")) + "   " + viewchrome.KeyHint(theme, "enter", "open in Today")
}

func legend(theme types.Theme, items ...[2]string) string {
	parts := []string{}
	for _, item := range items {
		parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color(item[1])).Render("●")+" "+theme.StyleDim.Render(item[0]))
	}
	return strings.Join(parts, "  ")
}

// sideLines are the risk chart, where the work went (b groups, B selects) and how the plan
// held.
func sideLines(theme types.Theme, state types.ContentState, width int) []string {
	lines := []string{}
	if burnout := latestBurnout(state); burnout != nil {
		lines = append(lines, rule(theme, "Risk", burnoutBadge(theme, burnout), width))
		lines = append(lines, riskChart(theme, burnout, width)...)
		for _, line := range strings.Split(ansi.Wordwrap(burnoutSummary(burnout), max(1, width), ""), "\n") {
			lines = append(lines, theme.StyleDim.Render(line))
		}
		lines = append(lines, "")
	}
	work := viewsummary.WorkLines(theme, state, width, false)
	lines = append(lines, rule(theme, "Work", workMeta(theme, state, 0), width))
	lines = append(lines, work[min(1, len(work)):]...)
	lines = append(lines, "", rule(theme, "Plan", planMeta(theme, state), width))
	if window := state.DashboardWindow; window != nil && window.PlannedCount > 0 {
		lines = append(lines, fmt.Sprintf("%d/%d done", window.CompletedCount, window.PlannedCount))
	}
	if progress := state.GoalProgress; progress != nil && progress.EstimatedItems > 0 {
		lines = append(lines, fmt.Sprintf("worked %s of %s estimated",
			viewhelpers.FormatCompactDurationSeconds(progress.TotalActualSeconds),
			helperpkg.FormatCompactDurationMinutes(progress.TotalEstimateMinutes)))
		lines = append(lines, theme.StyleDim.Render(estimateDrift(progress)))
	} else {
		lines = append(lines, theme.StyleDim.Render("no estimated work"))
	}
	for i := range lines {
		lines[i] = fit(lines[i], width)
	}
	return lines
}

// chartLines are the two charts, the date axis and the readout in height lines.
func chartLines(theme types.Theme, s daySeries, width, height int) []string {
	room := height - 4 // two headings, the date axis, the readout
	wellRows := max(3, room*55/100)
	workRows := max(3, room-wellRows)
	meta := theme.StyleDim.Render("┄ window average")
	if len(s.dates) > 14 {
		meta = theme.StyleDim.Render("3-day average · dots are days · ┄ window average")
	}
	lines := []string{rule(theme, "Wellbeing", meta, width)}
	lines = append(lines, wellbeingChart(theme, s, width, wellRows)...)
	lines = append(lines, rule(theme, "Work", legend(theme,
		[2]string{"hours worked", string(theme.ColorCyan)}, [2]string{"focus 0–100", string(theme.ColorLavender)}), width))
	lines = append(lines, workChart(theme, s, width, workRows)...)
	lines = append(lines, dateAxis(theme, s, width), readout(theme, s))
	for i := range lines {
		lines[i] = fit(lines[i], width)
	}
	return lines
}

// dashboardLines lay the window out in height lines of width.
func dashboardLines(theme types.Theme, state types.ContentState, width, height int) []string {
	s := collectSeries(state)
	lines := tileLines(theme, tilesFor(theme, state, s), width)
	lines = append(lines, "")
	room := height - len(lines)
	if width < dashboardSideAt {
		side := sideLines(theme, state, width)
		charts := chartLines(theme, s, width, max(10, room-len(side)-1))
		return append(append(append(lines, charts...), ""), side...)
	}
	sideW := min(sideMaxWidth, max(sideMinWidth, width*35/100))
	mainW := width - sideW - 3
	left := chartLines(theme, s, mainW, room)
	right := sideLines(theme, state, sideW)
	sep := theme.StyleDim.Render(" │ ")
	for i := range max(len(left), len(right)) {
		l, r := strings.Repeat(" ", mainW), ""
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		lines = append(lines, l+sep+r)
	}
	return lines
}
