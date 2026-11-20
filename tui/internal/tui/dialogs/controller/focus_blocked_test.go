package controller

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFocusBlockedConfirmationReturnsRecoveryAction(t *testing.T) {
	state := OpenFocusBlocked(State{}, 22, "Ship onboarding", "backlog", "ready")
	if state.Kind != "focus_blocked" {
		t.Fatalf("dialog kind = %q, want focus_blocked", state.Kind)
	}
	if len(state.ChoiceValues) != 2 || state.ChoiceValues[0] != "recover" || state.ChoiceValues[1] != "cancel" {
		t.Fatalf("choices = %v, want recover and cancel", state.ChoiceValues)
	}

	_, action, message := updateFocusBlocked(state, tea.KeyMsg{Type: tea.KeyEnter})
	if message != "" || action == nil || action.Kind != "prepare_focus_status" {
		t.Fatalf("confirmation action = %+v, message %q", action, message)
	}
	if action.IssueID != 22 || action.Status != "ready" {
		t.Fatalf("recovery action = %+v, want issue 22 -> ready", action)
	}
}

func TestFocusBlockedCancelClosesWithoutAction(t *testing.T) {
	state := OpenFocusBlocked(State{}, 22, "Ship onboarding", "backlog", "ready")
	state.ChoiceCursor = 1

	next, action, message := updateFocusBlocked(state, tea.KeyMsg{Type: tea.KeyEnter})
	if message != "" || action != nil || next.Kind != "" {
		t.Fatalf("cancel result = state %q, action %+v, message %q", next.Kind, action, message)
	}
}
