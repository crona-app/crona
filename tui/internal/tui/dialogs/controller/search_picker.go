package controller

import (
	tea "github.com/charmbracelet/bubbletea"
	"strings"
)

func contextOptions(state State, ctx UpdateContext, field int) []SelectorOption {
	if field == 0 {
		return DefaultRepoOptions(state.Inputs, ctx.Repos)
	}
	if state.Kind == "checkout_context" {
		return CheckoutStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)
	}
	return DefaultStreamOptions(state.Inputs, state.RepoIndex, ctx.Repos, ctx.AllIssues, ctx.Streams, ctx.Context)
}

func updateContextPicker(state State, ctx UpdateContext, msg tea.KeyMsg) (State, *Action, string, bool) {
	switch state.Kind {
	case "create_issue_default", "create_habit", "checkout_context":
	default:
		return state, nil, "", false
	}
	if len(state.Inputs) < 2 {
		return state, nil, "", false
	}
	key := msg.String()
	if key == "esc" {
		return state, nil, "", false
	}
	if state.Kind == "checkout_context" && key == "ctrl+x" {
		return Close(state), &Action{Kind: "checkout_context"}, "", true
	}
	if (key == "ctrl+s" || key == "ctrl+a" || key == "ctrl+f") && len(state.CaptureRows) == 0 {
		for field := range 2 {
			options := contextOptions(state, ctx, field)
			index, confirmed := state.RepoIndex, state.ConfirmedNewRepo
			if field == 1 {
				index, confirmed = state.StreamIndex, state.ConfirmedNewStream
			}
			if option, ok := selectedOption(options, index); ok && option.ID == "__new__" && confirmed != strings.TrimSpace(state.Inputs[field].Value()) {
				state.FocusIdx = field
				return SyncDialogFocus(state), nil, "Select the Create option with Enter before saving", true
			}
		}
		if state.Kind == "checkout_context" && key == "ctrl+s" {
			next, action, status := updateCheckoutContext(state, ctx, tea.KeyMsg{Type: tea.KeyEnter})
			return next, action, status, true
		}
		return state, nil, "", false
	}
	field := state.FocusIdx
	if field > 1 {
		if state.Kind == "checkout_context" {
			switch key {
			case "enter":
				return state, nil, "", false
			case "tab":
				state.FocusIdx = 0
			case "shift+tab":
				state.FocusIdx = 1
			}
			return SyncDialogFocus(state), nil, "", true
		}
		return state, nil, "", false
	}
	options := contextOptions(state, ctx, field)
	index := state.RepoIndex
	if field == 1 {
		index = state.StreamIndex
	}
	switch key {
	case "up", "down":
		dir := 1
		if key == "up" {
			dir = -1
		}
		if len(options) == 0 {
			return state, nil, "", true
		}
		if index < 0 && dir < 0 {
			index = len(options)
		}
		index = min(max(index+dir, 0), max(0, len(options)-1))
		if field == 0 {
			if index != state.RepoIndex {
				state.StreamIndex = 0
				state.Inputs[1].SetValue("")
				state.ConfirmedNewStream = ""
			}
			state.RepoIndex = index
		} else {
			state.StreamIndex = index
		}
		return clearDialogError(state), nil, "", true
	case "enter", "tab":
		option, ok := selectedOption(options, max(0, index))
		if !ok {
			return state, nil, "No matches: type a name or clear the filter", true
		}
		switch option.ID {
		case "__new__":
			if field == 0 {
				state.ConfirmedNewRepo = strings.TrimSpace(state.Inputs[0].Value())
			} else {
				state.ConfirmedNewStream = strings.TrimSpace(state.Inputs[1].Value())
			}
		case "__repo__":
			state.Inputs[field].SetValue("")
			state.StreamIndex = 0
		default:
			state.Inputs[field].SetValue(option.Label)
			if field == 0 {
				state.RepoIndex = 0
			} else {
				state.StreamIndex = 0
			}
		}
		state.FocusIdx++
		if field == 0 {
			state.StreamIndex = 0
		}
		return SyncDialogFocus(clearDialogError(state)), nil, "", true
	case "shift+tab":
		return state, nil, "", false
	default:
		if msg.Type != tea.KeyRunes && key != "backspace" && key != "delete" && key != "left" && key != "right" && key != "home" && key != "end" && key != "ctrl+u" && key != "ctrl+w" {
			return state, nil, "", false
		}
		old := ""
		if option, ok := selectedOption(options, index); ok {
			old = option.ID
		}
		before := state.Inputs[field].Value()
		state.Inputs[field], _ = state.Inputs[field].Update(msg)
		if before != state.Inputs[field].Value() {
			next := contextOptions(state, ctx, field)
			index = 0
			for i, option := range next {
				if option.ID == old && old != "__new__" {
					index = i
					break
				}
			}
			if field == 0 {
				if len(next) == 0 || next[index].ID != old {
					state.StreamIndex = 0
					state.Inputs[1].SetValue("")
				}
				state.RepoIndex = index
				state.ConfirmedNewRepo = ""
				state.ConfirmedNewStream = ""
			} else {
				state.StreamIndex = index
				state.ConfirmedNewStream = ""
			}
		}
		return clearDialogError(state), nil, "", true
	}
}
