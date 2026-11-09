package issues

import (
	"strings"

	viewchrome "crona.work/tui/internal/tui/views/chrome"
	"crona.work/tui/internal/tui/views/scopepane"
	types "crona.work/tui/internal/tui/views/types"
)

func renderView(theme types.Theme, state types.ContentState) string {
	if state.Width < 1 || state.Height < 1 {
		return ""
	}
	rows := StripRows(theme, state, state.Width)
	rows = rows[:min(len(rows), state.Height)]
	height := state.Height - len(rows)
	if height == 0 {
		return strings.Join(rows, "\n")
	}
	if state.Width < 100 {
		rows = append(rows, scopepane.RenderRow(theme, state, state.Width))
		if height > 1 {
			rows = append(rows, listColumn(theme, state, state.Width, height-1))
		}
		return strings.Join(rows, "\n")
	}
	scopeWidth := 24
	listWidth := state.Width - scopeWidth - 3
	columns := []string{
		scopepane.Render(theme, state, scopepane.IssueCounts(state.AllIssues), scopeWidth, height),
	}
	detailWidth := 0
	if state.Width >= 160 {
		detailWidth = (listWidth - 3) / 2
		listWidth -= detailWidth + 3
	}
	if detailWidth == 0 {
		columns = append(columns, listColumn(theme, state, listWidth, height))
	} else {
		columns = append(columns, renderIssues(theme, state, listWidth, height))
	}
	if detailWidth > 0 {
		columns = append(columns, renderDetail(theme, state, detailWidth, height))
	}
	rows = append(rows, viewchrome.JoinColumns(theme, height, columns...))
	return strings.Join(rows, "\n")
}

// listColumn is the list with the details stacked under it when the screen is too narrow for
// them beside it and tall enough to hold both.
func listColumn(theme types.Theme, state types.ContentState, width, height int) string {
	if height < viewchrome.StackMinHeight {
		return renderIssues(theme, state, width, height)
	}
	listH, detailH := viewchrome.StackHeights(height)
	return renderIssues(theme, state, width, listH) + "\n" + renderDetail(theme, state, width, detailH)
}
