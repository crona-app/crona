package review

import (
	"fmt"
	"math"

	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// planLines explains plan outcomes and estimate accuracy across the review window.
func planLines(theme types.Theme, state types.ContentState, width int) []string {
	window := state.DashboardWindow
	meta := theme.StyleDim.Render("nothing planned")
	if window != nil && window.PlannedCount > 0 {
		meta = fmt.Sprintf("%.0f%% completed", 100*float64(window.CompletedCount)/float64(window.PlannedCount))
	}
	lines := []string{rule(theme, "Plan", meta, width)}
	if window == nil {
		lines = append(lines, theme.StyleDim.Render("Plan outcomes unavailable."))
	} else if window.PlannedCount == 0 {
		lines = append(lines, theme.StyleDim.Render("No committed plan entries in this window."))
	} else {
		count := func(n int, color lipgloss.Color) string {
			return lipgloss.NewStyle().Foreground(color).Render(fmt.Sprint(n))
		}
		lines = append(lines,
			count(window.CompletedCount, theme.ColorGreen)+fmt.Sprintf("/%d done · committed plan entries", window.PlannedCount),
			count(window.FailedCount, theme.ColorRed)+fmt.Sprintf(" failed · %d missed their planned day", window.MissedCount),
			count(window.AbandonedCount, theme.ColorOrange)+" abandoned",
			count(window.CarryOverCount, theme.ColorYellow)+" carried from an earlier plan",
		)
		plannedDays, completeDays := 0, 0
		for _, day := range window.Days {
			if day.PlannedCount > 0 {
				plannedDays++
				if day.CompletedCount == day.PlannedCount {
					completeDays++
				}
			}
		}
		if plannedDays > 0 {
			lines = append(lines, fmt.Sprintf("%d/%d planned days completed in full", completeDays, plannedDays))
		}
		if state.ReviewPrev != nil && state.ReviewPrev.Window != nil && state.ReviewPrev.Window.PlannedCount > 0 {
			prev := state.ReviewPrev.Window
			delta := 100 * (float64(window.CompletedCount)/float64(window.PlannedCount) - float64(prev.CompletedCount)/float64(prev.PlannedCount))
			lines = append(lines, fmt.Sprintf("Completion %+.0f pp vs previous window", delta))
		}
		lines = append(lines, wrapDim(theme, "Counts are plan entries, not unique issues; carryovers can appear on multiple days.", width)...)
	}
	lines = append(lines, "")
	return append(lines, planEstimateLines(theme, state.GoalProgress, width)...)
}

func planEstimateLines(theme types.Theme, progress *api.GoalProgressSummary, width int) []string {
	if progress == nil {
		return []string{theme.StyleDim.Render("Estimate comparison unavailable.")}
	}
	if progress.TotalEstimateMinutes <= 0 {
		return wrapDim(theme, "No estimates recorded for this window. Add issue estimates to compare planned and tracked time.", width)
	}
	estimated := helperpkg.FormatCompactDurationMinutes(progress.TotalEstimateMinutes)
	actual := viewhelpers.FormatCompactDurationSeconds(progress.TotalActualSeconds)
	lines := []string{theme.StylePaneTitle.Render("Estimates"), "Estimated " + estimated + " · tracked " + actual}
	delta := progress.TotalActualSeconds - progress.TotalEstimateMinutes*60
	gap, color := "below", theme.ColorCyan
	if delta > 0 {
		gap, color = "above", theme.ColorOrange
	}
	text := fmt.Sprintf("Tracked time %s %s total estimate", viewhelpers.FormatCompactDurationSeconds(int(math.Abs(float64(delta)))), gap)
	if delta == 0 {
		text = "Tracked time matches total estimate"
	}
	lines = append(lines, lipgloss.NewStyle().Foreground(color).Render(text))
	if progress.EstimatedItems > 0 {
		lines = append(lines, wrapDim(theme, estimateDrift(progress), width)...)
		average := int(math.Round(math.Abs(progress.AverageDeltaMinutes) * 60))
		lines = append(lines, theme.StyleDim.Render("Average gap "+viewhelpers.FormatCompactDurationSeconds(average)+" per compared item"))
	} else {
		lines = append(lines, wrapDim(theme, "No estimated items have tracked time yet; accuracy is not available.", width)...)
	}
	return append(lines, wrapDim(theme, "Tracked time is time spent, not completion progress.", width)...)
}
