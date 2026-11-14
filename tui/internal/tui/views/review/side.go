package review

import (
	"fmt"
	"math"
	"strings"

	"crona.work/tui/internal/api"
	uistate "crona.work/tui/internal/tui/state"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"
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

// estimateDrift says how far estimates were off: over-estimates "ran high", under "ran low".
func estimateDrift(progress *api.GoalProgressSummary) string {
	pct := math.Abs(progress.AverageDeltaPercent)
	kind, plural := "issue", "issues"
	switch progress.GroupBy {
	case "repo":
		kind, plural = "repo", "repos"
	case "stream":
		kind, plural = "stream", "streams"
	}
	if progress.EstimatedItems != 1 {
		kind = plural
	}
	items := fmt.Sprintf("%d compared %s", progress.EstimatedItems, kind)
	switch strings.TrimSpace(progress.EstimateBias) {
	case "over":
		return fmt.Sprintf("estimates ran %.0f%% high across %s", pct, items)
	case "under":
		return fmt.Sprintf("estimates ran %.0f%% low across %s", pct, items)
	default:
		return "estimates held across " + items
	}
}
