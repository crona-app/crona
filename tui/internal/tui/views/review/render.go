// Package review renders the Insights tab's Review sub-tab: one window of days. Today owns
// the single day; Review shows the days, their trends and what the window adds up to.
package review

import (
	"fmt"
	"slices"
	"strings"

	helperpkg "crona.work/tui/internal/tui/helpers"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	// The side column takes about a third of the width within these bounds, and below
	// sideColumnAt there is no room for it beside the charts, so it stacks under them.
	sideMinWidth = 44
	sideMaxWidth = 64
	sideColumnAt = 110

	dayFormat = "Mon 2 Jan"
)

// Render shows the window as a dashboard: headline tiles, the Wellbeing and Work charts over
// the window's days with the selected day's readout, and beside them Risk, Work and Plan.
func Render(theme types.Theme, state types.ContentState) string {
	if state.Width < 1 || state.Height < 1 {
		return ""
	}
	if state.HistoricalAwayDate != "" {
		return renderAway(theme, state)
	}
	s := collectSeries(state)
	hints := []string{viewchrome.KeyHint(theme, "←/→", "day"), viewchrome.KeyHint(theme, "enter", "open in Today"),
		viewchrome.KeyHint(theme, "+/-/0", fmt.Sprintf("%dd window", len(s.days)))}
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: windowTitle(s), Meta: viewchrome.FitHints(hints, max(0, state.Width-24)),
		Width: state.Width, Height: state.Height, Lines: dashboardLines(theme, state, s, state.Width-2, state.Height-2),
	})
}

func windowTitle(s daySeries) string {
	if len(s.days) == 0 {
		return "Review"
	}
	return fmt.Sprintf("%s – %s · %d days", s.days[0].Format(dayFormat), s.days[len(s.days)-1].Format(dayFormat), len(s.days))
}

func dashboardLines(theme types.Theme, state types.ContentState, s daySeries, width, height int) []string {
	m := metricsFor(theme)
	lines := tileLines(theme, tilesFor(theme, state, s, m), width)
	lines = append(lines, "")
	room := height - len(lines)
	if width < sideColumnAt {
		side := sideLines(theme, state, width)
		main := chartLines(theme, s, m, width, max(10, room-len(side)-1))
		return slices.Concat(lines, main, []string{""}, side)
	}
	sideW := min(sideMaxWidth, max(sideMinWidth, width*35/100))
	mainW := width - sideW - 3
	left := chartLines(theme, s, m, mainW, room)
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

func renderAway(theme types.Theme, state types.ContentState) string {
	message := fmt.Sprintf("You chose to rest and recover on %s.",
		helperpkg.FormatDisplayDate(state.HistoricalAwayDate, state.Settings))
	lines := make([]string, max(0, (state.Height-2)/2))
	lines = append(lines, lipgloss.PlaceHorizontal(max(1, state.Width-2), lipgloss.Center, theme.StyleHeader.Render(message)))
	return viewchrome.RenderSection(theme, viewchrome.Section{Title: "Away", Width: state.Width, Height: state.Height, Lines: lines})
}
