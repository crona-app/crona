package model

import (
	"errors"
	"strings"
	"testing"

	"crona.work/shared/protocol"
	"crona.work/tui/internal/tui/commands"
	dialogstate "crona.work/tui/internal/tui/dialogs/controller"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestErrorSurvivesSuccessAndStatusClear(t *testing.T) {
	m := NewDailyRenderModel(60, 20)
	m = m.withStatus("Could not save your issue", true)
	_ = m.setStatus("Refreshed", false)
	result, _ := m.update(commands.ClearStatusMsg{Seq: m.statusSeq})
	m = result.(Model)
	if !strings.Contains(ansi.Strip(m.View()), "Could not save your issue") {
		t.Fatal("success and status clear hid the unacknowledged error")
	}
}

func TestEscapeDismissesErrorWithoutClosingForm(t *testing.T) {
	m := NewDailyRenderModel(60, 20)
	m = m.withDialogState(dialogstate.OpenCreateRepo(m.dialogState()))
	m = m.withStatus("Could not refresh", true)
	result, _ := m.update(tea.KeyMsg{Type: tea.KeyEsc})
	next := result.(Model)
	if next.dialog == "" {
		t.Fatal("dismissing the notice closed the form")
	}
	if strings.Contains(ansi.Strip(next.View()), "Could not refresh") {
		t.Fatal("Escape did not dismiss notice")
	}
}

func TestBackgroundErrorDoesNotOverwriteDialogValidation(t *testing.T) {
	m := NewDailyRenderModel(60, 20)
	state := dialogstate.OpenCreateRepo(m.dialogState())
	state.ErrorMessage = "Name is required"
	m = m.withDialogState(state)
	result, _ := m.update(commands.ErrMsg{Err: &protocol.RPCError{Code: "request_failed", Message: "Could not refresh repositories"}})
	next := result.(Model)
	if next.dialogErrorMessage != "Name is required" {
		t.Fatalf("background error replaced validation: %q", next.dialogErrorMessage)
	}
	rendered := ansi.Strip(next.View())
	if !strings.Contains(rendered, "Could not refresh repositories") || strings.Contains(rendered, "request_failed") {
		t.Fatalf("unexpected error presentation: %q", rendered)
	}
}

func TestErrorNoticeShowsWhileSessionOverlayIsOpen(t *testing.T) {
	m := NewDailyRenderModel(60, 20)
	m.sessionContextOpen = true
	result, _ := m.update(commands.ErrMsg{Err: errors.New("Connection lost; check the daemon")})
	if !strings.Contains(ansi.Strip(result.(Model).View()), "Connection lost") {
		t.Fatal("session overlay hid connection error")
	}
}
