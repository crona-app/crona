package viewchrome

// StackMinHeight is the shortest body that stacks a details pane under its list; shorter,
// the list keeps the room and enter shows the same details in a dialog.
const StackMinHeight = 24

// StackHeights splits height between a list and the details stacked under it, the details
// taking half.
func StackHeights(height int) (list, detail int) {
	detail = height / 2
	return height - detail, detail
}
