package review

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"

	"crona.work/tui/internal/api"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// riskLines are the Risk heading with the score, what moves it, and what to do about it.
func riskLines(theme types.Theme, burnout *api.BurnoutIndicator, width int) []string {
	color, advice := theme.ColorGreen, "Current signals look stable. Maintain your recovery rhythm."
	switch burnout.Level {
	case "guarded":
		color, advice = theme.ColorYellow, "Signals are mixed. Keep scope tight and protect breaks."
	case "high":
		color, advice = theme.ColorRed, "Risk is elevated. Reduce load and prioritize recovery today."
	}
	badge := lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%d %s", burnout.Score, strings.ToUpper(string(burnout.Level))))
	lines := []string{rule(theme, "Risk", badge, width)}
	lines = append(lines, riskChart(theme, burnout, width)...)
	return append(lines, wrapDim(theme, advice, width)...)
}

// burnoutDriver is one factor of the risk score, in points: positive raises it.
type burnoutDriver struct {
	name   string
	points int
}

// burnoutDrivers are the factors easing the score (most first) and raising it (most first),
// up to three of each; factors under a point say nothing and are left out.
func burnoutDrivers(burnout *api.BurnoutIndicator) (easing, raising []burnoutDriver) {
	for name, score := range burnout.Factors {
		points := int(math.Round(score * 100))
		switch {
		case points > 0:
			raising = append(raising, burnoutDriver{factorName(name), points})
		case points < 0:
			easing = append(easing, burnoutDriver{factorName(name), points})
		}
	}
	slices.SortFunc(raising, func(a, b burnoutDriver) int {
		return cmp.Or(cmp.Compare(b.points, a.points), cmp.Compare(a.name, b.name))
	})
	slices.SortFunc(easing, func(a, b burnoutDriver) int {
		return cmp.Or(cmp.Compare(a.points, b.points), cmp.Compare(a.name, b.name))
	})
	return easing[:min(3, len(easing))], raising[:min(3, len(raising))]
}

// factorNames are the daemon's burnout factor keys in words; an unknown key shows as is.
var factorNames = map[string]string{
	"workloadPressure":    "Workload pressure",
	"breakDebt":           "Break debt",
	"moodEnergyDrag":      "Mood/energy drag",
	"sleepDebt":           "Sleep debt",
	"recoveryConsistency": "Recovery consistency",
	"recoveryBreaks":      "Break recovery",
	"loadStability":       "Load stability",
}

func factorName(key string) string { return cmp.Or(factorNames[key], key) }

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
	for _, driver := range slices.Concat(easing, raising) {
		largest = max(largest, int(math.Abs(float64(driver.points))))
	}
	length := func(points int) int {
		return max(1, int(math.Round(math.Abs(float64(points))/float64(largest)*float64(barWidth))))
	}
	padLeft := func(text string, w int) string {
		return strings.Repeat(" ", max(0, w-ansi.StringWidth(text))) + text
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

func wrapDim(theme types.Theme, text string, width int) []string {
	lines := strings.Split(ansi.Wordwrap(text, max(1, width), ""), "\n")
	for i := range lines {
		lines[i] = theme.StyleDim.Render(lines[i])
	}
	return lines
}
