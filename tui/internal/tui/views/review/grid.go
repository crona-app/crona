package review

import (
	"fmt"
	"math"
	"strings"
	"time"

	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	uistate "crona.work/tui/internal/tui/state"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"

	sharedtypes "crona.work/shared/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Review is one timeline: every section shares the window's columns (a day each, or a
// week each in a long window), then the window's total and its change on the window before.
const (
	gridLabelWidth = 12
	// gridMinCell and gridMaxCell bound a column; columns widen to fill the screen.
	gridMinCell    = 7
	gridMaxCell    = 14
	gridTotalWidth = 11
	gridPrevWidth  = 10
	// gridMaxWorkRows is the most Work rows shown; a short screen shows fewer.
	gridMaxWorkRows = 15
)

// gridColumn is one column of the timeline.
type gridColumn struct {
	label    string
	dates    []string
	selected bool
}

func gridColumns(state types.ContentState) []gridColumn {
	dates := WindowDates(state)
	if len(dates) == 0 {
		return nil
	}
	runs := uistate.ReviewColumns(dates[0], dates[len(dates)-1], state.WeekStart)
	columns := make([]gridColumn, 0, len(runs))
	for _, run := range runs {
		col := gridColumn{dates: run}
		if t, err := time.Parse(isoDate, run[0]); err == nil {
			if len(run) == 1 && len(runs) == len(dates) {
				col.label = t.Format("Mon 2")
			} else {
				_, week := t.ISOWeek()
				col.label = fmt.Sprintf("W%d", week)
			}
		}
		for _, date := range run {
			col.selected = col.selected || date == state.DashboardDate
		}
		columns = append(columns, col)
	}
	return columns
}

// grid draws rows against the columns, dropping the comparison then the total when narrow.
type grid struct {
	theme              types.Theme
	columns            []gridColumn
	width              int
	showTotal, showPrv bool
	// hasPrev is whether a previous window loaded; the last column also holds Work shares.
	hasPrev bool
	cell    int
}

func newGrid(theme types.Theme, state types.ContentState, width int) grid {
	g := grid{theme: theme, columns: gridColumns(state), width: width, showTotal: true, showPrv: true, hasPrev: state.ReviewPrev != nil, cell: gridMinCell}
	need := func() int {
		w := gridLabelWidth + len(g.columns)*g.cell
		if g.showTotal {
			w += gridTotalWidth
		}
		if g.showPrv {
			w += gridPrevWidth
		}
		return w
	}
	if n := len(g.columns); n > 0 {
		g.cell = min(gridMaxCell, max(gridMinCell, (width-gridLabelWidth-gridTotalWidth-gridPrevWidth)/n))
	}
	if need() > width {
		g.showPrv = false
	}
	if need() > width {
		g.showTotal = false
	}
	return g
}

func padLeft(text string, w int) string {
	return strings.Repeat(" ", max(0, w-ansi.StringWidth(text))) + text
}

// line is a row: its label, a styled cell per column (each g.cell wide), its total and
// its comparison.
func (g grid) line(label string, cells []string, total, prev string) string {
	var b strings.Builder
	b.WriteString(label)
	for _, cell := range cells {
		b.WriteString(cell)
	}
	if g.showTotal {
		b.WriteString(padLeft(total, gridTotalWidth))
	}
	if g.showPrv {
		b.WriteString(padLeft(prev, gridPrevWidth))
	}
	return viewhelpers.TruncateANSI(b.String(), g.width)
}

func (g grid) label(text string) string {
	return g.theme.StyleDim.Render(viewchrome.LabelCell(" "+text, gridLabelWidth-1) + " ")
}

func (g grid) header() string {
	cells := make([]string, 0, len(g.columns))
	for _, col := range g.columns {
		if col.selected {
			cells = append(cells, g.theme.StyleCursor.Render(padLeft("▸"+col.label, g.cell)))
		} else {
			cells = append(cells, g.theme.StyleDim.Render(padLeft(col.label, g.cell)))
		}
	}
	return g.line(strings.Repeat(" ", gridLabelWidth), cells,
		g.theme.StyleDim.Render("Window"), g.theme.StyleDim.Render(map[bool]string{true: "vs prev"}[g.hasPrev]))
}

// heading is a section rule across the width: its title, then meta at the right.
func (g grid) heading(title, meta string) string {
	left := g.theme.StyleDim.Render("─ ") + g.theme.StylePaneTitle.Render(title) + " "
	right := ""
	if meta != "" {
		right = " " + meta + g.theme.StyleDim.Render(" ─")
	}
	fill := g.width - ansi.StringWidth(left) - ansi.StringWidth(right)
	if fill < 1 {
		return viewhelpers.TruncateANSI(left+meta, g.width)
	}
	return left + g.theme.StyleDim.Render(strings.Repeat("─", fill)) + right
}

// heat is a value cell tinted by how good it is, so the window's shape shows in colour.
func (g grid) heat(text string, color lipgloss.Color, selected bool) string {
	if text == "" {
		return g.theme.StyleDim.Render(padLeft("· ", g.cell))
	}
	style := lipgloss.NewStyle().Foreground(color).Reverse(true)
	if selected {
		style = style.Bold(true)
	}
	return " " + style.Render(padLeft(text+" ", g.cell-1))
}

func (g grid) plain(text string, color lipgloss.Color, selected bool) string {
	if text == "" {
		return g.theme.StyleDim.Render(padLeft("–", g.cell))
	}
	style := lipgloss.NewStyle()
	if color != "" {
		style = style.Foreground(color)
	}
	if selected {
		style = style.Bold(true)
	}
	return style.Render(padLeft(text, g.cell))
}

func band(theme types.Theme, v, high, low float64) lipgloss.Color {
	switch {
	case v >= high:
		return theme.ColorGreen
	case v >= low:
		return theme.ColorYellow
	default:
		return theme.ColorRed
	}
}

// change is the difference on the previous window: ▲ more (green), ▼ less (red), = level.
// For every metric here, more is better.
func change(theme types.Theme, delta float64, format func(float64) string) string {
	switch {
	case math.Abs(delta) < 0.05:
		return theme.StyleDim.Render("=")
	case delta > 0:
		return lipgloss.NewStyle().Foreground(theme.ColorGreen).Render("▲ " + format(delta))
	default:
		return lipgloss.NewStyle().Foreground(theme.ColorRed).Render("▼ " + format(-delta))
	}
}

func trimFloat(v float64) string {
	if v == math.Trunc(v) {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.1f", v)
}

// windowData indexes the window's loaded data by date.
type windowData struct {
	metrics map[string]api.DailyMetricsDay
	plans   map[string]sharedtypes.DashboardWindowDay
	scores  map[string]api.FocusScoreRangeDay
	prev    types.ReviewPrev
}

func indexWindow(state types.ContentState) windowData {
	d := windowData{
		metrics: map[string]api.DailyMetricsDay{},
		plans:   map[string]sharedtypes.DashboardWindowDay{},
		scores:  map[string]api.FocusScoreRangeDay{},
	}
	for _, day := range state.MetricsRange {
		d.metrics[day.Date] = day
	}
	if state.DashboardWindow != nil {
		for _, day := range state.DashboardWindow.Days {
			d.plans[day.Date] = day
		}
	}
	for _, day := range state.ReviewScores {
		d.scores[day.Date] = day
	}
	if state.ReviewPrev != nil {
		d.prev = *state.ReviewPrev
	}
	return d
}

// average is a metric over a column's dates, skipping days without it.
func average(dates []string, value func(string) (float64, bool)) (float64, bool) {
	total, n := 0.0, 0
	for _, date := range dates {
		if v, ok := value(date); ok {
			total, n = total+v, n+1
		}
	}
	if n == 0 {
		return 0, false
	}
	return total / float64(n), true
}

// wellbeingLines are mood, energy, sleep and focus as a heatmap, each with its window
// average and change. Days without a check-in show "·".
func wellbeingLines(g grid, state types.ContentState, d windowData) []string {
	theme, rollup := g.theme, state.MetricsRollup
	checkIn := func(date string) *api.DailyCheckIn { return d.metrics[date].CheckIn }
	row := func(label string, value func(string) (float64, bool), format func(float64) string,
		color func(float64) lipgloss.Color, total, prev string,
	) string {
		cells := make([]string, 0, len(g.columns))
		for _, col := range g.columns {
			v, ok := average(col.dates, value)
			if !ok {
				cells = append(cells, g.heat("", "", false))
				continue
			}
			cells = append(cells, g.heat(format(v), color(v), col.selected))
		}
		return g.line(g.label(label), cells, theme.StyleHeader.Render(total), prev)
	}
	scale := func(label string, pick func(*api.DailyCheckIn) int, avg func(*api.MetricsRollup) *float64) string {
		total, prev := "–", ""
		if rollup != nil && avg(rollup) != nil {
			total = "avg " + fmt.Sprintf("%.1f", *avg(rollup))
			if d.prev.Rollup != nil && avg(d.prev.Rollup) != nil {
				prev = change(theme, *avg(rollup)-*avg(d.prev.Rollup), func(v float64) string { return fmt.Sprintf("%.1f", v) })
			}
		}
		return row(label, func(date string) (float64, bool) {
			c := checkIn(date)
			if c == nil {
				return 0, false
			}
			return float64(pick(c)), true
		}, trimFloat, func(v float64) lipgloss.Color { return band(theme, v, 4, 3) }, total, prev)
	}
	lines := []string{
		scale("Mood", func(c *api.DailyCheckIn) int { return c.Mood }, func(r *api.MetricsRollup) *float64 { return r.AverageMood }),
		scale("Energy", func(c *api.DailyCheckIn) int { return c.Energy }, func(r *api.MetricsRollup) *float64 { return r.AverageEnergy }),
	}
	sleepTotal, sleepPrev := "–", ""
	if rollup != nil && rollup.AverageSleepHours != nil {
		sleepTotal = "avg " + helperpkg.FormatCompactDurationHours(*rollup.AverageSleepHours)
		if d.prev.Rollup != nil && d.prev.Rollup.AverageSleepHours != nil {
			sleepPrev = change(theme, *rollup.AverageSleepHours-*d.prev.Rollup.AverageSleepHours, helperpkg.FormatCompactDurationHours)
		}
	}
	lines = append(lines, row("Sleep", func(date string) (float64, bool) {
		c := checkIn(date)
		if c == nil || c.SleepHours == nil {
			return 0, false
		}
		return *c.SleepHours, true
	}, helperpkg.FormatCompactDurationHours, func(v float64) lipgloss.Color { return band(theme, v, 7, 6) }, sleepTotal, sleepPrev))
	focusTotal, focusPrev := "–", ""
	if state.WeeklyFocusScore != nil {
		focusTotal = fmt.Sprintf("avg %d", state.WeeklyFocusScore.Score)
		if d.prev.Focus != nil {
			focusPrev = change(theme, float64(state.WeeklyFocusScore.Score-d.prev.Focus.Score), func(v float64) string { return fmt.Sprintf("%.0f", v) })
		}
	}
	lines = append(lines, row("Focus", func(date string) (float64, bool) {
		day, ok := d.scores[date]
		return float64(day.Score), ok && day.HasData
	}, func(v float64) string { return fmt.Sprintf("%.0f", v) },
		func(v float64) lipgloss.Color { return band(theme, v, 60, 40) }, focusTotal, focusPrev))
	return lines
}

// wellbeingMeta is the window's check-in count and its change, for the Wellbeing heading.
func wellbeingMeta(theme types.Theme, state types.ContentState, d windowData) string {
	rollup := state.MetricsRollup
	if rollup == nil {
		return ""
	}
	meta := theme.StyleDim.Render(fmt.Sprintf("%d/%d check-ins", rollup.CheckInDays, rollup.Days))
	if d.prev.Rollup != nil && rollup.CheckInDays != d.prev.Rollup.CheckInDays {
		meta += " " + change(theme, float64(rollup.CheckInDays-d.prev.Rollup.CheckInDays), func(v float64) string { return fmt.Sprintf("%.0f", v) })
	}
	return meta
}

var heatGlyphs = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

// workRows are the window's work groups (or a drill level's) with a heat block per column for
// when the time went, their window total and share; then the day's total worked.
func workRows(g grid, state types.ContentState, d windowData, limit int) (lines []string, hidden int) {
	theme := g.theme
	total := workTotal(g, state, d)
	summary := state.WorkBreakdown
	if summary == nil {
		return []string{g.label("") + theme.StyleDim.Render("Loading"), total}, 0
	}
	rows := make([]sharedtypes.TimeDistributionRow, 0, len(summary.Rows))
	for _, row := range summary.Rows {
		if row.WorkedSeconds > 0 {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		return []string{g.label("") + theme.StyleDim.Render("No work logged."), total}, 0
	}
	// cell[i][key] is a column's seconds per row; the busiest cell sets the scale.
	cells := make([]map[string]int, len(g.columns))
	busiest := 0
	for i := range g.columns {
		cells[i] = map[string]int{}
		if i < len(state.WorkColumns) && state.WorkColumns[i] != nil {
			for _, row := range state.WorkColumns[i].Rows {
				cells[i][row.Key] = row.WorkedSeconds
				busiest = max(busiest, row.WorkedSeconds)
			}
		}
	}
	cursor := min(max(0, state.DailyWorkCursor), len(rows)-1)
	start, end := 0, min(limit, len(rows))
	if state.DailyWorkSelect {
		start, end = viewchrome.ListWindow(cursor, len(rows), max(1, limit))
	}
	for i := start; i < end; i++ {
		row := rows[i]
		label := g.label(row.Label)
		if state.DailyWorkSelect && i == cursor {
			label = theme.StyleCursor.Render(viewchrome.LabelCell("▶"+row.Label, gridLabelWidth-1) + " ")
		}
		columns := make([]string, 0, len(g.columns))
		for c, col := range g.columns {
			seconds := cells[c][row.Key]
			if seconds <= 0 || busiest == 0 {
				columns = append(columns, theme.StyleDim.Render(padLeft("·", g.cell)))
				continue
			}
			level := min(len(heatGlyphs)-1, int(float64(seconds)/float64(busiest)*float64(len(heatGlyphs))))
			style := lipgloss.NewStyle().Foreground(theme.ColorCyan)
			if col.selected {
				style = style.Bold(true)
			}
			columns = append(columns, "  "+style.Render(strings.Repeat(heatGlyphs[level], g.cell-3))+" ")
		}
		share := fmt.Sprintf("%3.0f%% ", row.Percent) + shareBar(theme, row.Percent, 4)
		lines = append(lines, g.line(label, columns,
			helperpkg.FormatCompactDurationSeconds(row.WorkedSeconds), theme.StyleDim.Render(share)))
	}
	return append(lines, total), len(rows) - (end - start)
}

// workTotal is the worked time per column, over the window and against the window before.
func workTotal(g grid, state types.ContentState, d windowData) string {
	theme := g.theme
	totals := make([]string, 0, len(g.columns))
	for _, col := range g.columns {
		seconds := 0
		for _, date := range col.dates {
			seconds += d.metrics[date].WorkedSeconds
		}
		text := ""
		if seconds > 0 {
			text = viewhelpers.FormatCompactDurationSeconds(seconds)
		}
		totals = append(totals, g.plain(text, "", col.selected))
	}
	total, prev := "–", ""
	if rollup := state.MetricsRollup; rollup != nil {
		total = viewhelpers.FormatCompactDurationSeconds(rollup.WorkedSeconds)
		if d.prev.Rollup != nil {
			prev = change(theme, float64(rollup.WorkedSeconds-d.prev.Rollup.WorkedSeconds), func(v float64) string {
				return viewhelpers.FormatCompactDurationSeconds(int(v))
			})
		}
	}
	return g.line(theme.StyleHeader.Render(viewchrome.LabelCell(" Total", gridLabelWidth-1)+" "),
		totals, theme.StyleHeader.Render(total), prev)
}

func shareBar(theme types.Theme, percent float64, width int) string {
	filled := min(width, max(0, int(percent/100*float64(width)+0.5)))
	if percent > 0 && filled == 0 {
		filled = 1
	}
	return lipgloss.NewStyle().Foreground(theme.ColorCyan).Render(strings.Repeat("█", filled)) +
		theme.StyleDim.Render(strings.Repeat("░", width-filled))
}

// workMeta is the Work heading's keys: b and B, or the selection's own keys.
func workMeta(theme types.Theme, state types.ContentState, hidden int) string {
	group := uistate.WorkLevelGroup(state.DailyWorkGroup, state.DailyWorkPath)
	hints := []string{}
	if hidden > 0 && !state.DailyWorkSelect {
		hints = append(hints, theme.StyleDim.Render(fmt.Sprintf("+%d more", hidden)))
	}
	if state.DailyWorkSelect {
		if uistate.WorkDrillable(group) {
			hints = append(hints, viewchrome.KeyHint(theme, "enter", "open"))
		}
		hints = append(hints, viewchrome.KeyHint(theme, "esc", "back"))
	} else {
		hints = append(hints, viewchrome.KeyHint(theme, "b", "group"))
		if uistate.WorkDrillable(group) {
			hints = append(hints, viewchrome.KeyHint(theme, "B", "select"))
		}
	}
	return strings.Join(hints, "  ")
}

// planLines are the plan's done count per column and how estimates held over the window.
func planLines(g grid, state types.ContentState, d windowData) []string {
	theme := g.theme
	cells := make([]string, 0, len(g.columns))
	for _, col := range g.columns {
		done, planned := 0, 0
		for _, date := range col.dates {
			done += d.plans[date].CompletedCount
			planned += d.plans[date].PlannedCount
		}
		if planned == 0 {
			cells = append(cells, g.plain("", "", false))
			continue
		}
		cells = append(cells, g.plain(fmt.Sprintf("%d/%d", done, planned),
			band(theme, float64(done)/float64(planned), 1, 0.01), col.selected))
	}
	total, prev := "–", ""
	if window := state.DashboardWindow; window != nil && window.PlannedCount > 0 {
		total = fmt.Sprintf("%d/%d", window.CompletedCount, window.PlannedCount)
		if d.prev.Window != nil {
			prev = change(theme, float64(window.CompletedCount-d.prev.Window.CompletedCount), func(v float64) string {
				return fmt.Sprintf("%.0f", v)
			})
		}
	}
	lines := []string{g.line(g.label("Done"), cells, theme.StyleHeader.Render(total), prev)}
	estimates := theme.StyleDim.Render("no estimated work")
	if progress := state.GoalProgress; progress != nil && progress.EstimatedItems > 0 {
		estimates = fmt.Sprintf("worked %s of %s estimated", viewhelpers.FormatCompactDurationSeconds(progress.TotalActualSeconds),
			helperpkg.FormatCompactDurationMinutes(progress.TotalEstimateMinutes)) +
			theme.StyleDim.Render(" — "+estimateDrift(progress))
	}
	return append(lines, viewhelpers.TruncateANSI(g.label("Estimates")+estimates, g.width))
}

// planMeta is the window's carried and missed counts, for the Plan heading.
func planMeta(theme types.Theme, state types.ContentState) string {
	window := state.DashboardWindow
	if window == nil || window.PlannedCount == 0 {
		return theme.StyleDim.Render("nothing planned")
	}
	colored := func(color lipgloss.Color, n int, word string) string {
		return lipgloss.NewStyle().Foreground(color).Render(fmt.Sprint(n)) + theme.StyleDim.Render(" "+word)
	}
	return strings.Join([]string{
		colored(theme.ColorRed, window.MissedCount, "missed"),
		colored(theme.ColorYellow, window.CarryOverCount, "carried"),
	}, theme.StyleDim.Render(" · "))
}

// riskChart draws what moves the risk score as bars from a centre line: easing to the left,
// raising to the right, each named on its outer side.
func riskChart(theme types.Theme, burnout *api.BurnoutIndicator, width int) []string {
	easing, raising := burnoutDrivers(burnout)
	if len(easing)+len(raising) == 0 {
		return nil
	}
	const nameWidth = 22
	barWidth := max(4, min(24, (width-2*nameWidth-6)/2))
	largest := 1
	for _, driver := range append(append([]burnoutDriver{}, easing...), raising...) {
		largest = max(largest, int(math.Abs(float64(driver.points))))
	}
	length := func(points int) int {
		return max(1, int(math.Round(math.Abs(float64(points))/float64(largest)*float64(barWidth))))
	}
	axis := theme.StyleDim.Render("│")
	green := lipgloss.NewStyle().Foreground(theme.ColorGreen)
	red := lipgloss.NewStyle().Foreground(theme.ColorRed)
	left := strings.Repeat(" ", nameWidth+1+barWidth)
	lines := []string{
		theme.StyleDim.Render(padLeft("easing ◀", nameWidth+1+barWidth)) + axis + theme.StyleDim.Render("▶ raising"),
	}
	for _, driver := range easing {
		n := length(driver.points)
		lines = append(lines, padLeft(driver.name, nameWidth)+" "+strings.Repeat(" ", barWidth-n)+
			green.Render(strings.Repeat("█", n))+axis+theme.StyleDim.Render(fmt.Sprintf(" %d", driver.points)))
	}
	for _, driver := range raising {
		n := length(driver.points)
		lines = append(lines, left+axis+red.Render(strings.Repeat("█", n))+
			theme.StyleDim.Render(fmt.Sprintf(" +%d ", driver.points))+driver.name)
	}
	for i := range lines {
		lines[i] = viewhelpers.TruncateANSI(lines[i], width)
	}
	return lines
}

// timelineLines lays the window out in height lines: Wellbeing, Work, Plan and Risk. Work
// gets as many rows as fit (up to gridMaxWorkRows); a short screen drops the risk chart
// first, keeping its heading.
func timelineLines(theme types.Theme, state types.ContentState, width, height int) []string {
	g := newGrid(theme, state, width)
	d := indexWindow(state)
	group := uistate.WorkLevelGroup(state.DailyWorkGroup, state.DailyWorkPath)
	wellbeing := wellbeingLines(g, state, d)
	plan := planLines(g, state, d)
	var chart []string
	riskTitle, riskMeta := "Risk", ""
	burnout := latestBurnout(state)
	if burnout != nil {
		riskMeta = burnoutBadge(theme, burnout) + theme.StyleDim.Render("  "+burnoutSummary(burnout))
		chart = riskChart(theme, burnout, width)
	}
	// Fixed: header, four headings, the wellbeing rows, the work chips and total, the plan.
	fixed := 1 + 4 + len(wellbeing) + 2 + len(plan)
	if burnout == nil {
		fixed--
	}
	room := height - fixed - len(chart)
	if room < 3 {
		chart, room = nil, height-fixed
	}
	work, hidden := workRows(g, state, d, max(1, min(gridMaxWorkRows, room)))
	// Left-over height goes between the sections, a line each.
	gap := []string{}
	if height-fixed-len(chart)-len(work) >= 3 {
		gap = []string{""}
	}
	workTitle := "Work by " + strings.ToLower(workGroupName(group))
	lines := []string{g.header(), g.heading("Wellbeing", wellbeingMeta(theme, state, d))}
	lines = append(lines, wellbeing...)
	lines = append(lines, gap...)
	lines = append(lines, g.heading(workTitle, workMeta(theme, state, hidden)),
		viewsummary.WorkHeading(theme, state, group, width))
	lines = append(lines, work...)
	lines = append(lines, gap...)
	lines = append(lines, g.heading("Plan", planMeta(theme, state)))
	lines = append(lines, plan...)
	if burnout != nil {
		lines = append(lines, gap...)
		lines = append(lines, g.heading(riskTitle, riskMeta))
		lines = append(lines, chart...)
	}
	return lines
}

func workGroupName(group string) string {
	switch group {
	case "segment_type":
		return "segment"
	case "":
		return "repo"
	}
	return group
}
