package review

import (
	"fmt"
	"math"
	"strings"

	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	uistate "crona.work/tui/internal/tui/state"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// sideLines are the risk, where the work went and how the plan held.
func sideLines(theme types.Theme, state types.ContentState, width int) []string {
	lines := []string{}
	if state.MetricsRollup != nil && state.MetricsRollup.LatestBurnout != nil {
		lines = append(lines, riskLines(theme, state.MetricsRollup.LatestBurnout, width)...)
		lines = append(lines, "")
	}
	lines = append(lines, rule(theme, "Work", workMeta(theme, state), width))
	lines = append(lines, viewsummary.WorkBody(theme, state, width)...)
	lines = append(lines, "")
	lines = append(lines, planLines(theme, state, width)...)
	for i := range lines {
		lines[i] = fit(lines[i], width)
	}
	return lines
}

// workMeta is the Work heading's keys: b and B, or the selection's own keys.
func workMeta(theme types.Theme, state types.ContentState) string {
	group := uistate.WorkLevelGroup(state.DailyWorkGroup, state.DailyWorkPath)
	hints := []string{}
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

func planLines(theme types.Theme, state types.ContentState, width int) []string {
	window := state.DashboardWindow
	planned := window != nil && window.PlannedCount > 0
	meta := theme.StyleDim.Render("nothing planned")
	if planned {
		count := func(color lipgloss.Color, n int, word string) string {
			return lipgloss.NewStyle().Foreground(color).Render(fmt.Sprint(n)) + theme.StyleDim.Render(" "+word)
		}
		meta = count(theme.ColorRed, window.MissedCount, "missed") + theme.StyleDim.Render(" · ") +
			count(theme.ColorYellow, window.CarryOverCount, "carried")
	}
	lines := []string{rule(theme, "Plan", meta, width)}
	if planned {
		lines = append(lines, planCount(window.CompletedCount, window.PlannedCount)+" done")
	}
	if progress := state.GoalProgress; progress != nil && progress.EstimatedItems > 0 {
		lines = append(lines, fmt.Sprintf("worked %s of %s estimated",
			viewhelpers.FormatCompactDurationSeconds(progress.TotalActualSeconds),
			helperpkg.FormatCompactDurationMinutes(progress.TotalEstimateMinutes)),
			theme.StyleDim.Render(estimateDrift(progress)))
	} else {
		lines = append(lines, theme.StyleDim.Render("no estimated work"))
	}
	return lines
}

// estimateDrift says how far estimates were off: over-estimates "ran high", under "ran low".
func estimateDrift(progress *api.GoalProgressSummary) string {
	pct := math.Abs(progress.AverageDeltaPercent)
	items := fmt.Sprintf("%d %s", progress.EstimatedItems,
		map[bool]string{true: "issue", false: "issues"}[progress.EstimatedItems == 1])
	switch strings.TrimSpace(progress.EstimateBias) {
	case "over":
		return fmt.Sprintf("estimates ran %.0f%% high across %s", pct, items)
	case "under":
		return fmt.Sprintf("estimates ran %.0f%% low across %s", pct, items)
	default:
		return "estimates held across " + items
	}
}
