// Package review renders the Insights tab's Review sub-tab: one window of days. Today owns
// the single day; Review shows the days, their trends and what the window adds up to.
package review

import (
	"math"

	"crona.work/tui/internal/api"
	"fmt"
	"strings"
	"time"

	helperpkg "crona.work/tui/internal/tui/helpers"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// Render shows the window as one timeline under its title: Wellbeing, Work, Plan and Risk,
// every section on the same columns.
func Render(theme types.Theme, state types.ContentState) string {
	if state.Width < 1 || state.Height < 1 {
		return ""
	}
	if state.HistoricalAwayDate != "" {
		return renderAway(theme, state, state.Width, state.Height)
	}
	hints := []string{viewchrome.KeyHint(theme, "←/→", "day"), viewchrome.KeyHint(theme, "enter", "open in Today"),
		viewchrome.KeyHint(theme, "+/-/0", fmt.Sprintf("%dd window", len(WindowDates(state))))}
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: windowTitle(state), Meta: viewchrome.FitHints(hints, max(0, state.Width-24)),
		Width: state.Width, Height: state.Height, Lines: dashboardLines(theme, state, state.Width-2, state.Height-2),
	})
}

// windowTitle names the window: "Last 7 days" ending today, else its dates.
func windowTitle(state types.ContentState) string {
	dates := WindowDates(state)
	if len(dates) == 0 {
		return "Review"
	}
	return fmt.Sprintf("%s – %s · %d days", dayTitle(dates[0]), dayTitle(dates[len(dates)-1]), len(dates))
}

func renderAway(theme types.Theme, state types.ContentState, width, height int) string {
	lines := make([]string, max(0, (height-2)/2))
	lines = append(
		lines,
		lipgloss.PlaceHorizontal(
			max(1, width-2),
			lipgloss.Center,
			theme.StyleHeader.Render(awayMessage(state)),
		),
	)
	return viewchrome.RenderSection(
		theme,
		viewchrome.Section{Title: "Away", Width: width, Height: height, Lines: lines},
	)
}

func awayMessage(state types.ContentState) string {
	return fmt.Sprintf(
		"You chose to rest and recover on %s.",
		helperpkg.FormatDisplayDate(state.HistoricalAwayDate, state.Settings),
	)
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

func dayTitle(date string) string {
	if t, err := time.Parse(isoDate, date); err == nil {
		return t.Format("Mon 2 Jan")
	}
	return date
}
