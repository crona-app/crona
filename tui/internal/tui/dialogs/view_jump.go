package dialogs

import (
	"strings"

	controllerpkg "crona.work/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/x/ansi"
)

// viewJumpWidth is the jump dialog's body width when the screen allows it.
const viewJumpWidth = 50

// viewJumpGridMin is how many views a group needs before it folds into a grid.
const viewJumpGridMin = 5

// renderViewJump draws the jump dialog: views grouped under their tab, the tab's number key on
// the right, each row its key and name. The selected row carries its description; the view the
// dialog was opened from is marked. A large group (System) folds into a grid of three.
func renderViewJump(theme Theme, state controllerpkg.State) string {
	width := min(viewJumpWidth, max(24, state.Width-10))
	groups := viewJumpGroups(state)
	body := []string{}
	focus := -1
	for gi, group := range groups {
		if gi > 0 {
			body = append(body, "")
		}
		name, number, _ := strings.Cut(group.head, "\t")
		gap := max(1, width-ansi.StringWidth(name)-ansi.StringWidth(number))
		body = append(body, theme.StylePaneTitle.Render(name)+strings.Repeat(" ", gap)+theme.StyleDim.Render(number))
		if len(group.items) >= viewJumpGridMin && width >= 3*16 {
			lines, detail, at := viewJumpGrid(theme, state, group.items, width)
			if at >= 0 {
				focus = len(body) + at
			}
			body = append(body, lines...)
			if detail != "" {
				body = append(body, "    "+theme.StyleDim.Render(ansi.Truncate(detail, width-4, "…")))
			}
			continue
		}
		for _, idx := range group.items {
			if idx == state.ChoiceCursor {
				focus = len(body)
			}
			body = append(body, viewJumpRow(theme, state, idx, width))
		}
	}
	return renderListDialog(theme, state, listDialog{
		title: state.ViewTitle, body: body, focus: focus, primary: "enter jump",
	})
}

type viewJumpGroup struct {
	head  string
	items []int
}

// viewJumpGroups splits the choices at each group heading.
func viewJumpGroups(state controllerpkg.State) []viewJumpGroup {
	var groups []viewJumpGroup
	for idx := range state.ChoiceItems {
		head := ""
		if idx < len(state.ChoiceGroups) {
			head = state.ChoiceGroups[idx]
		}
		if head != "" || len(groups) == 0 {
			groups = append(groups, viewJumpGroup{head: head})
		}
		groups[len(groups)-1].items = append(groups[len(groups)-1].items, idx)
	}
	return groups
}

// viewJumpCell is a choice's pointer, key and name, the opened-from view marked with a dot.
func viewJumpCell(theme Theme, state controllerpkg.State, idx int) string {
	key := viewJumpKey(state.ChoiceItems[idx])
	name := viewJumpName(state.ChoiceItems[idx])
	pointer := "  "
	nameStyle := theme.StyleNormal
	if idx == state.ChoiceCursor {
		pointer = theme.StyleCursor.Render("▶ ")
		nameStyle = theme.StyleSelected
	}
	cell := pointer + theme.StyleHeader.Render(key) + "  " + nameStyle.Render(name)
	if idx < len(state.ChoiceValues) && state.ChoiceValues[idx] == state.ChoiceCurrent {
		cell += theme.StyleDim.Render(" ●")
	}
	return cell
}

// viewJumpRow is one listed view; the selected one adds its description, right-aligned.
func viewJumpRow(theme Theme, state controllerpkg.State, idx, width int) string {
	row := viewJumpCell(theme, state, idx)
	if idx != state.ChoiceCursor || idx >= len(state.ChoiceDetails) {
		return row
	}
	room := width - ansi.StringWidth(row) - 3
	detail := strings.TrimSpace(state.ChoiceDetails[idx])
	if room < 8 || detail == "" {
		return row
	}
	detail = ansi.Truncate(detail, room, "…")
	return row + strings.Repeat(" ", width-ansi.StringWidth(row)-ansi.StringWidth(detail)) + theme.StyleDim.Render(detail)
}

// viewJumpGrid lays a group out three to a row. It returns the lines, the selected view's
// description when the selection is in the grid, and the selected line's index (-1 if none).
func viewJumpGrid(theme Theme, state controllerpkg.State, items []int, width int) ([]string, string, int) {
	cellWidth := width / 3
	var lines []string
	detail, at := "", -1
	for start := 0; start < len(items); start += 3 {
		line := ""
		for _, idx := range items[start:min(len(items), start+3)] {
			cell := viewJumpCell(theme, state, idx)
			line += cell + strings.Repeat(" ", max(1, cellWidth-ansi.StringWidth(cell)))
			if idx == state.ChoiceCursor {
				at = len(lines)
				if idx < len(state.ChoiceDetails) {
					detail = strings.TrimSpace(state.ChoiceDetails[idx])
				}
			}
		}
		lines = append(lines, strings.TrimRight(line, " "))
	}
	return lines, detail, at
}

// viewJumpKey and viewJumpName split a "[k] Name" choice label.
func viewJumpKey(item string) string {
	if end := strings.Index(item, "]"); strings.HasPrefix(item, "[") && end > 1 {
		return item[1:end]
	}
	return ""
}

func viewJumpName(item string) string {
	if end := strings.Index(item, "]"); strings.HasPrefix(item, "[") && end > 0 {
		return strings.TrimSpace(item[end+1:])
	}
	return item
}
