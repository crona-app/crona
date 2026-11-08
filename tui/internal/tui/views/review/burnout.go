package review

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"

	"crona.work/tui/internal/api"
	types "crona.work/tui/internal/tui/views/types"
	"github.com/charmbracelet/lipgloss"
)

func latestBurnout(state types.ContentState) *api.BurnoutIndicator {
	if state.MetricsRollup == nil || state.MetricsRollup.LatestBurnout == nil {
		return nil
	}
	return state.MetricsRollup.LatestBurnout
}

func burnoutBadge(theme types.Theme, burnout *api.BurnoutIndicator) string {
	style := lipgloss.NewStyle().Foreground(theme.ColorGreen)
	switch burnout.Level {
	case "guarded":
		style = lipgloss.NewStyle().Foreground(theme.ColorYellow)
	case "high":
		style = lipgloss.NewStyle().Foreground(theme.ColorRed)
	}
	return style.Render(fmt.Sprintf("%d %s", burnout.Score, strings.ToUpper(string(burnout.Level))))
}

func burnoutSummary(burnout *api.BurnoutIndicator) string {
	switch burnout.Level {
	case "high":
		return "Risk is elevated. Reduce load and prioritize recovery today."
	case "guarded":
		return "Signals are mixed. Keep scope tight and protect breaks."
	default:
		return "Current signals look stable. Maintain your recovery rhythm."
	}
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
			raising = append(raising, burnoutDriver{prettifyBurnoutFactor(name), points})
		case points < 0:
			easing = append(easing, burnoutDriver{prettifyBurnoutFactor(name), points})
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

func prettifyBurnoutFactor(name string) string {
	switch name {
	case "workloadPressure":
		return "Workload pressure"
	case "breakDebt":
		return "Break debt"
	case "moodEnergyDrag":
		return "Mood/energy drag"
	case "sleepDebt":
		return "Sleep debt"
	case "recoveryConsistency":
		return "Recovery consistency"
	case "recoveryBreaks":
		return "Break recovery"
	case "loadStability":
		return "Load stability"
	default:
		return name
	}
}
