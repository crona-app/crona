package review

import (
	"fmt"
	"math"
	"strings"

	"crona.work/tui/internal/tui/views/charts"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// axisWidth holds a chart's y labels and its axis line; every chart row and the date axis
// start after it, so the days line up down the column.
const axisWidth = 5

// smoothRadius is how many days either side a long window's lines average over.
const smoothRadius = 1

// The scales the lanes are drawn on: check-in ratings, and the sleep hours that matter.
const (
	ratingLo, ratingHi = 1, 5
	sleepLo, sleepHi   = 4, 9
)

const shortDayFormat = "Mon 2"

// chartLines are the Wellbeing lanes, the Work chart, the date axis and the readout in
// height lines.
func chartLines(theme types.Theme, s daySeries, m metricSet, width, height int) []string {
	room := height - 4 // two headings, the date axis, the readout
	wellRows := max(3, room*55/100)
	workRows := max(3, room-wellRows)
	meta := "┄ window average"
	if s.long {
		meta = fmt.Sprintf("%d-day average · dots are days · ", 2*smoothRadius+1) + meta
	}
	lines := []string{rule(theme, "Wellbeing", theme.StyleDim.Render(meta), width)}
	lines = append(lines, wellbeingChart(theme, s, m, width, wellRows)...)
	lines = append(lines, rule(theme, "Work", legend(theme,
		legendItem{"hours worked", m.worked.color}, legendItem{"focus 0–100", m.focus.color}), width))
	lines = append(lines, workChart(theme, s, m, width, workRows)...)
	lines = append(lines, dateAxis(theme, s, width), readout(theme, s, m))
	for i := range lines {
		lines[i] = fit(lines[i], width)
	}
	return lines
}

// lane is one series of the Wellbeing chart on its own scale, so the lines never cross.
type lane struct {
	metric
	values []float64
	lo, hi float64
}

func wellbeingChart(theme types.Theme, s daySeries, m metricSet, width, rows int) []string {
	lanes := []lane{{m.mood, s.mood, ratingLo, ratingHi}, {m.energy, s.energy, ratingLo, ratingHi}, {m.sleep, s.sleep, sleepLo, sleepHi}}
	laneRows := max(2, (rows-len(lanes))/len(lanes))
	lines := make([]string, 0, rows)
	for _, l := range lanes {
		lines = append(lines, laneHeading(theme, l, s))
		lines = append(lines, laneChart(theme, l, s, width, laneRows)...)
	}
	return lines
}

// laneHeading names a lane with its window average and the selected day's value.
func laneHeading(theme types.Theme, l lane, s daySeries) string {
	parts := []string{lipgloss.NewStyle().Foreground(l.color).Bold(true).Render(l.title())}
	if avg := mean(l.values); !math.IsNaN(avg) {
		parts = append(parts, theme.StyleDim.Render("avg ")+l.format(avg))
	} else {
		parts = append(parts, theme.StyleDim.Render("none logged"))
	}
	if i := s.selected; i >= 0 && !math.IsNaN(l.values[i]) {
		parts = append(parts, theme.StyleDim.Render(s.days[i].Format(shortDayFormat)+" ")+l.format(l.values[i]))
	}
	return strings.Repeat(" ", axisWidth) + strings.Join(parts, theme.StyleDim.Render(" · "))
}

// laneChart draws a lane's line, or over a long window its rolling average with the
// days as faint dots, and its window average dotted across.
func laneChart(theme types.Theme, l lane, s daySeries, width, rows int) []string {
	canvas := charts.NewLineCanvas(max(1, width-axisWidth), rows)
	const line, day = 0, 1
	if s.long {
		canvas.Dots(l.values, l.lo, l.hi, day)
		canvas.Plot(rolling(l.values, smoothRadius), l.lo, l.hi, line)
	} else {
		canvas.Plot(l.values, l.lo, l.hi, line)
	}
	avgRow := -1
	if avg := mean(l.values); !math.IsNaN(avg) {
		avgRow = canvas.Row(avg, l.lo, l.hi)
	}
	solid := lipgloss.NewStyle().Foreground(l.color)
	faint := solid.Faint(true)
	labels := map[int]string{0: l.format(l.hi), rows - 1: l.format(l.lo)}
	return drawCanvas(theme, canvas, s, rows, labels, func(row, cell int) (string, bool) {
		if r := canvas.Rune(row, cell); r != 0 {
			if canvas.Owner[row][cell] == day {
				return faint.Render(string(r)), true
			}
			return solid.Render(string(r)), true
		}
		if row == avgRow {
			return theme.StyleDim.Render("┄"), true
		}
		return "", false
	})
}

// workChart draws the hours worked each day as bars, and the focus score (0–100 on the same
// height) as a line over them.
func workChart(theme types.Theme, s daySeries, m metricSet, width, rows int) []string {
	plotW := max(1, width-axisWidth)
	canvas := charts.NewLineCanvas(plotW, rows)
	canvas.Plot(s.focus, 0, 100, 0)
	top := 1.0
	for _, h := range s.worked {
		top = max(top, math.Ceil(h))
	}
	n := len(s.days)
	barW := max(1, min(3, plotW/max(1, n)/2))
	bars := map[int]float64{}
	for i, h := range s.worked {
		left := canvas.Cell(i, n) - (barW-1)/2
		for x := max(0, left); x < min(plotW, left+barW); x++ {
			bars[x] = h
		}
	}
	barStyle := lipgloss.NewStyle().Foreground(m.worked.color)
	lineStyle := lipgloss.NewStyle().Foreground(m.focus.color)
	labels := map[int]string{0: fmt.Sprintf("%.0fh", top), rows - 1: "0"}
	return drawCanvas(theme, canvas, s, rows, labels, func(row, cell int) (string, bool) {
		if r := canvas.Rune(row, cell); r != 0 {
			return lineStyle.Render(string(r)), true
		}
		h := bars[cell]
		fill := int(math.Round(h/top*float64(rows*8))) - (rows-1-row)*8
		if h <= 0 || fill <= 0 {
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
		cursor = canvas.Cell(s.selected, len(s.days))
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

// dateAxis labels the days under the charts: the selected (▸) first so it always shows, then
// the first, the last and evenly between where they fit.
func dateAxis(theme types.Theme, s daySeries, width int) string {
	plotW := max(1, width-axisWidth)
	n := len(s.days)
	row := []rune(strings.Repeat(" ", plotW))
	taken := make([]bool, plotW)
	format := shortDayFormat
	if s.long {
		format = "Jan 2"
	}
	place := func(i int, mark bool) {
		label := []rune(s.days[i].Format(format))
		if mark {
			label = append([]rune("▸"), label...)
		}
		start := min(max(0, charts.PointCell(i, n, plotW)-len(label)/2), plotW-len(label))
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
func readout(theme types.Theme, s daySeries, m metricSet) string {
	i := s.selected
	if i < 0 {
		return ""
	}
	parts := []string{theme.StyleHeader.Render("▸ " + s.days[i].Format(dayFormat))}
	for _, v := range []struct {
		metric
		value float64
	}{{m.mood, s.mood[i]}, {m.energy, s.energy[i]}, {m.sleep, s.sleep[i]}, {m.worked, s.worked[i]}, {m.focus, s.focus[i]}} {
		if !math.IsNaN(v.value) {
			parts = append(parts, theme.StyleDim.Render(v.name+" ")+v.format(v.value))
		}
	}
	if s.planned[i] > 0 {
		parts = append(parts, theme.StyleDim.Render("plan ")+planCount(s.done[i], s.planned[i]))
	}
	return strings.Join(parts, theme.StyleDim.Render(" · ")) + "   " + viewchrome.KeyHint(theme, "enter", "open in Today")
}

type legendItem struct {
	text  string
	color lipgloss.Color
}

func legend(theme types.Theme, items ...legendItem) string {
	parts := []string{}
	for _, item := range items {
		parts = append(parts, lipgloss.NewStyle().Foreground(item.color).Render("●")+" "+theme.StyleDim.Render(item.text))
	}
	return strings.Join(parts, "  ")
}
