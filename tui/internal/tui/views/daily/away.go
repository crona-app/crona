package daily

import (
	"strings"

	viewchrome "crona.work/tui/internal/tui/views/chrome"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// awaySection preserves the pane's footprint, with one recovery message in Tasks.
func awaySection(theme types.Theme, title string, width, height int, message bool) string {
	var lines []string
	if message && height > 1 {
		body := lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true).Render("Away day")
		if height > 3 {
			body += "\n" + theme.StyleNormal.Render("You chose to rest and recover.")
		}
		centered := lipgloss.Place(max(1, width-2), height-1, lipgloss.Center, lipgloss.Center, body)
		lines = strings.Split(centered, "\n")
	}
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: title, Width: width, Height: height, Lines: lines,
	})
}

func blankRegion(width, height int) string {
	if width < 1 || height < 1 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat(strings.Repeat(" ", width)+"\n", height), "\n")
}
