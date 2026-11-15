package viewchrome

import (
	"slices"
	"strings"

	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// AwayDayStrip announces an explicitly recorded away date without adding header rows.
func AwayDayStrip(theme Theme, date string, settings *api.CoreSettings, width int) string {
	if width < 1 || settings == nil || !slices.Contains(settings.AwayDates, date) {
		return ""
	}
	accent := lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true)
	text := accent.Render("▌ Away day") + " · " + helperpkg.FormatDisplayDate(date, settings) + "  " + theme.StyleDim.Render("You chose to rest and recover.")
	text = ansi.Truncate(text, width, "…")
	return text + strings.Repeat(" ", max(0, width-ansi.StringWidth(text)))
}

// AwayBand adds a subdued background while retaining the chart's foreground colors.
func AwayBand(text string) string {
	return lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{Light: "#dcefe4", Dark: "#17392d"}).Render(text)
}
