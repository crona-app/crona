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
	state.ItemDetailActions = state.Pane == "issues"
	return issue.Title, itemdetail.IssueLines(theme, state, issue, width), true
}

func renderDetail(theme types.Theme, state types.ContentState, width, height int) string {
	title, lines, ok := Detail(theme, state, width-2)
	var actions [][2]string
	meta := ""
	if ok && state.Pane == "issues" {
		indices, cur := issueIndices(state), state.Cursors["issues"]
		actions = itemdetail.IssueFooterKeys(state, state.DefaultIssues[indices[cur]].Issue)
		meta = viewchrome.KeyHint(theme, "enter", "expand")
	}
	return itemdetail.RenderPane(theme, viewchrome.Section{Title: title, Meta: meta, Width: width, Height: height, Lines: lines}, actions)
}
