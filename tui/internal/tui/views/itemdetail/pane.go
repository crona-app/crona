package itemdetail

import (
	"strings"

	viewchrome "crona.work/tui/internal/tui/views/chrome"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// ActionLines wraps complete key hints without dropping actions at narrow widths.
func ActionLines(theme types.Theme, width int, actions [][2]string) []string {
	lines := []string{}
	line := ""
	for _, action := range actions {
		hint := viewchrome.KeyHint(theme, action[0], action[1])
		if line != "" && ansi.StringWidth(line+"  "+hint) > width {
			lines = append(lines, line)
			line = ""
		}
		if line != "" {
			line += "  "
		}
		line += hint
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

// RenderPane keeps general actions below the content, matching interactive item dialogs.
func RenderPane(theme types.Theme, section viewchrome.Section, actions [][2]string) string {
	footer := ActionLines(theme, max(1, section.Width-2), actions)
	if len(footer) > 0 && len(footer) <= section.Height-3 {
		visible := section.Height - 1 - len(footer)
		lines := append([]string{}, section.Lines[:min(len(section.Lines), visible)]...)
		if len(section.Lines) > visible {
			section.Meta = strings.TrimSpace(theme.StyleDim.Render("↓") + "  " + section.Meta)
		}
		for len(lines) < visible {
			lines = append(lines, "")
		}
		section.Lines = append(lines, footer...)
	}
	return viewchrome.RenderSection(theme, section)
}
