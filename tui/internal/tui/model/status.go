package model

import (
	"strings"
	"time"

	"crona.work/tui/internal/logger"
	commands "crona.work/tui/internal/tui/commands"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	statusInfoDuration = 3 * time.Second
)

func (m *Model) withStatus(message string, isError bool) Model {
	m.statusSeq++
	m.statusMsg = message
	m.statusErr = isError
	if isError {
		m.errorNotice = message
	}
	logStatusError(message, isError)
	return *m
}

func (m *Model) setStatus(message string, isError bool) tea.Cmd {
	m.statusSeq++
	m.statusMsg = message
	m.statusErr = isError
	if isError {
		m.errorNotice = message
	}
	logStatusError(message, isError)
	duration := statusInfoDuration
	if isError {
		return nil
	}
	return commands.ClearStatusAfter(m.statusSeq, duration)
}

func logStatusError(message string, isError bool) {
	if !isError {
		return
	}
	message = strings.TrimSpace(message)
	if message == "" || strings.HasPrefix(message, "Error: ") {
		return
	}
	logger.Error("UI error: " + message)
}

// statusMessage carries feedback from input adapters back into the model.
type statusMessage struct {
	message string
	isError bool
}
