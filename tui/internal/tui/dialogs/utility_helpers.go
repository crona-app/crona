package dialogs

import (
	"slices"
	"strings"

	sharedtypes "crona.work/shared/types"

	"github.com/charmbracelet/x/ansi"
)

func exportCategoryTitle(category string) string {
	switch category {
	case "summary":
		return "Summary"
	case "project":
		return "Project Reports"
	case "data":
		return "Data Exports"
	default:
		return "Narrative Reports"
	}
}

func hasStreakKind(values []sharedtypes.StreakKind, target sharedtypes.StreakKind) bool {
	return slices.Contains(values, target)
}




func renderViewEntityBody(theme Theme, body string) string {
	lines := strings.Split(body, "\n")
	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch trimmed {
		case "Description", "Notes":
			rendered = append(rendered, theme.StyleDim.Render(trimmed))
		default:
			rendered = append(rendered, line)
		}
	}
	return strings.Join(rendered, "\n")
}

// renderViewMeta lays a view dialog's "Key value" meta fields out in up to 3 aligned
// columns that fit width, like the detail pane's field grid.
func renderViewMeta(theme Theme, meta string, width int) []string {
	const gap = 3
	type field struct{ label, value string }
	fields := []field{}
	labelWidth := 0
	for _, part := range strings.Split(meta, "   ") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, value, _ := strings.Cut(part, " ")
		key = strings.ToLower(key)
		fields = append(fields, field{key, strings.TrimSpace(value)})
		labelWidth = max(labelWidth, ansi.StringWidth(key)+1)
	}
	cell := 0
	for _, f := range fields {
		cell = max(cell, labelWidth+ansi.StringWidth(f.value))
	}
	cols := max(1, min(3, len(fields), (width+gap)/(cell+gap)))
	lines := []string{}
	for start := 0; start < len(fields); start += cols {
		cells := []string{}
		for _, f := range fields[start:min(len(fields), start+cols)] {
			label := f.label + strings.Repeat(" ", labelWidth-ansi.StringWidth(f.label))
			text := theme.StyleDim.Render(label) + theme.StyleHeader.Render(f.value)
			cells = append(cells, text+strings.Repeat(" ", max(0, cell-ansi.StringWidth(text))))
		}
		lines = append(lines, strings.TrimRight(strings.Join(cells, strings.Repeat(" ", gap)), " "))
	}
	return lines
}
