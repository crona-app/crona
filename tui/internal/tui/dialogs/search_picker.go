package dialogs

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// SearchPicker exposes choices immediately rather than hiding them behind arrows.
func SearchPicker(theme Theme, spec PickerSpec, width int) []string {
	if !spec.Focused {
		return []string{Field(theme, spec.Label, spec.Column, spec.Current, false)}
	}
	rows := []string{Field(theme, spec.Label, spec.Column, spec.Input, true)}
	selected := min(max(0, spec.Selected), max(0, len(spec.Options)-1))
	start := min(max(0, selected-2), max(0, len(spec.Options)-5))
	for i := range 5 {
		at := start + i
		text := ""
		if at < len(spec.Options) {
			text = "  " + ansi.Truncate(spec.Options[at], max(1, width-4), "…")
			if at == selected {
				text = theme.StyleCursor.Render("▶ " + ansi.Truncate(spec.Options[at], max(1, width-4), "…"))
			}
		} else if i == 0 {
			text = theme.StyleDim.Render("  No matches")
		}
		rows = append(rows, text)
	}
	detail := ""
	if len(spec.Options) > 0 && ansi.StringWidth(spec.Options[selected]) > max(1, width-4) {
		detail = spec.Options[selected]
	}
	full := strings.Split(ansi.Wordwrap(detail, max(1, width-2), ""), "\n")
	for i := range 2 {
		text := ""
		if i < len(full) {
			text = "  " + ansi.Truncate(full[i], max(1, width-2), "…")
		}
		if i == 1 && len(full) > 2 {
			text = ansi.Truncate(text, max(1, width-1), "") + "…"
		}
		rows = append(rows, theme.StyleDim.Render(text))
	}
	return rows
}
