package habitlist

import (
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	"crona.work/tui/internal/tui/views/itemdetail"
	types "crona.work/tui/internal/tui/views/types"
)

// Detail is the selected habit's name and details, as the details pane and the details
// dialog both show them; ok is false with nothing selected.
func Detail(theme types.Theme, state types.ContentState, width int) (title string, lines []string, ok bool) {
	indices := habitIndices(state)
	cur := state.Cursors["habits"]
	if cur < 0 || cur >= len(indices) {
		return "Selected", []string{theme.StyleDim.Render("Nothing selected")}, false
	}
	habit := state.ScopedHabits[indices[cur]]
	return habit.Name, itemdetail.DefinitionLines(theme, state, habit, width), true
}

func renderDetail(theme types.Theme, state types.ContentState, width, height int) string {
	title, lines, _ := Detail(theme, state, width-2)
	return viewchrome.RenderSection(theme, viewchrome.Section{Title: title, Width: width, Height: height, Lines: lines})
}
