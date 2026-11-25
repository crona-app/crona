package layout

import (
	"strings"

	"crona.work/tui/internal/tui/chrome"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Render reserves space for a notice in every screen, including modal screens.
func Render(state State) string {
	if state.Width != 0 && (state.Width < MinWidth || state.Height < MinHeight) {
		return renderMinimumSizeWarning(state.Width, state.Height)
	}
	message := state.ErrorNotice
	if message == "" && state.StatusErr {
		message = state.StatusMsg
	}
	if message == "" || state.Width < MinWidth || state.Height < MinHeight {
		return renderFrame(state)
	}
	notice := renderErrorNotice(message, state.Width, state.Height)
	rows := lipgloss.Height(notice)
	state.Height -= rows
	state.ContentState.Height = ContentHeight(state)
	state.DialogState.Height = state.Height
	if state.StatusErr {
		state.StatusMsg = ""
		state.StatusErr = false
	}
	if state.DialogOpen || state.SessionDetailOpen || state.SessionContextOpen {
		state.Height--
		state.DialogState.Height = state.Height
		body := renderFrame(state)
		return clipViewportString(strings.Join([]string{body, notice, chrome.StyleDim.Render("  ctrl+c quit")}, "\n"), state.Width, state.Height+rows+1)
	}
	lines := strings.Split(renderFrame(state), "\n")
	last := len(lines) - 1
	footer := lines[last]
	return strings.Join(append(append(lines[:last], notice), footer), "\n")
}

func renderErrorNotice(message string, width, height int) string {
	message = strings.Join(strings.Fields(ansi.Strip(message)), " ")
	message = strings.TrimPrefix(message, "Error: ")
	lines := strings.Split(ansi.Wrap(message, width-4, ""), "\n")
	maxBody := max(1, min(4, height/3-2))
	if len(lines) > maxBody {
		lines = lines[:maxBody]
		lines[maxBody-1] = ansi.Truncate(lines[maxBody-1], width-4-len("… details in logs"), "") + "… details in logs"
	}
	rows := []string{chrome.StyleError.Render("  Error")}
	for _, line := range lines {
		rows = append(rows, chrome.StyleError.Render("  "+line))
	}
	rows = append(rows, chrome.StyleDim.Render("  esc dismiss"))
	return strings.Join(rows, "\n")
}
