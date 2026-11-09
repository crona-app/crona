package issues

import (
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	"crona.work/tui/internal/tui/views/itemdetail"
	types "crona.work/tui/internal/tui/views/types"
)

// Detail is the selected issue's title and details, as the details pane and the details
// dialog both show them; ok is false with nothing selected.
func Detail(theme types.Theme, state types.ContentState, width int) (title string, lines []string, ok bool) {
	indices := issueIndices(state)
	cur := state.Cursors["issues"]
	if cur < 0 || cur >= len(indices) {
		return "Selected", []string{theme.StyleDim.Render("Nothing selected")}, false
	}
	issue := state.DefaultIssues[indices[cur]].Issue
	return issue.Title, itemdetail.IssueLines(theme, state, issue, width), true
}

func renderDetail(theme types.Theme, state types.ContentState, width, height int) string {
	title, lines, _ := Detail(theme, state, width-2)
	return viewchrome.RenderSection(theme, viewchrome.Section{Title: title, Width: width, Height: height, Lines: lines})
}
