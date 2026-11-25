package dialogruntime

import (
	"crona.work/tui/internal/tui/commands"

	tea "github.com/charmbracelet/bubbletea"
)

// runFormMutations completes each save before returning the refresh messages.
// A failure stops later mutations; refreshes themselves remain asynchronous.
func runFormMutations(cmds ...tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		var refresh []tea.Cmd
		for _, cmd := range cmds {
			if cmd == nil {
				continue
			}
			msg := cmd()
			if _, failed := msg.(commands.ErrMsg); failed {
				return msg
			}
			if msg != nil {
				refresh = append(refresh, func() tea.Msg { return msg })
			}
		}
		if batch := tea.Batch(refresh...); batch != nil {
			return batch()
		}
		return nil
	}
}
