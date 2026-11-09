package review

import (
	"fmt"
	"math"
	"strings"

	"crona.work/tui/internal/api"
	uistate "crona.work/tui/internal/tui/state"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	types "crona.work/tui/internal/tui/views/types"

	sharedtypes "crona.work/shared/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func padLeft(text string, w int) string {
	return strings.Repeat(" ", max(0, w-ansi.StringWidth(text))) + text
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
