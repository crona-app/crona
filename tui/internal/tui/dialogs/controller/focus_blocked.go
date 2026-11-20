package controller

import tea "github.com/charmbracelet/bubbletea"

func updateFocusBlocked(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "up", "down":
		if state.ChoiceCursor == 0 {
			state.ChoiceCursor = 1
		} else {
			state.ChoiceCursor = 0
		}
		return state, nil, ""
	case "enter":
		if state.ChoiceCursor == 1 {
			return Close(state), nil, ""
		}
		return Close(state), &Action{
			Kind:    "prepare_focus_status",
			IssueID: state.IssueID,
			Status:  state.RecoveryStatus,
		}, ""
	default:
		return state, nil, ""
	}
}
