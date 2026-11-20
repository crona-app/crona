package dialogs

import (
	"strings"

	controllerpkg "crona.work/tui/internal/tui/dialogs/controller"
)

func renderFocusBlocked(theme Theme, state controllerpkg.State) string {
	body := []string{
		theme.StyleNormal.Render(state.ViewName + " is in " + strings.ReplaceAll(state.IssueStatus, "_", " ") + "."),
		theme.StyleNormal.Render("Focus sessions start from Planned, Ready, or In Progress."),
		"",
	}
	body = append(body, selectList(theme, state.ChoiceItems, state.ChoiceCursor)...)
	return renderListDialog(theme, state, listDialog{
		title:   "Can't focus yet",
		context: state.ViewName,
		border:  theme.ColorYellow,
		primary: "enter choose",
		hints:   []string{"↑/↓ move"},
		focus:   len(body) - len(state.ChoiceItems) + state.ChoiceCursor,
		body:    body,
	})
}
