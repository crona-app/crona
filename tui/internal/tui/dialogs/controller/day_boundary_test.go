package controller

import (
	"testing"

	sharedtypes "crona.work/shared/types"

	tea "github.com/charmbracelet/bubbletea"
)

func dbKeys(t *testing.T, state State, keys ...string) (State, *Action) {
	t.Helper()
	var action *Action
	for _, key := range keys {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		switch key {
		case "up":
			msg = tea.KeyMsg{Type: tea.KeyUp}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "space":
			msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
		case "bksp":
			msg = tea.KeyMsg{Type: tea.KeyBackspace}
		case "ctrl+s":
			msg = tea.KeyMsg{Type: tea.KeyCtrlS}
		}
		state, action, _ = Update(state, UpdateContext{}, "2026-11-08", msg)
	}
	return state, action
}

func openSchedule() State {
	return OpenEditDayBoundary(State{}, sharedtypes.CoreSettingsKeyStartOfDay, sharedtypes.DayBoundarySchedule{
		Enabled: true, DefaultTime: "06:00", WeekdayOverrides: map[int]string{3: "07:30"},
	}, "Local")
}

func TestDayBoundaryLoadsDaysIntoRows(t *testing.T) {
	state := openSchedule()
	if state.Inputs[0].Value() != "06:00" || state.Inputs[3].Value() != "07:30" || state.Inputs[1].Value() != "" {
		t.Fatalf("inputs = %q %q %q", state.Inputs[0].Value(), state.Inputs[3].Value(), state.Inputs[1].Value())
	}
	if idx, _ := DayBoundaryInputForRow(state, DayBoundaryRowDefault+1); idx != 1 {
		t.Fatalf("first day row on a Monday week = %d", idx)
	}
	state.DueDateWeekStart = sharedtypes.WeekStartSunday
	if idx, _ := DayBoundaryInputForRow(state, DayBoundaryRowDefault+1); idx != 7 {
		t.Fatalf("first day row on a Sunday week = %d", idx)
	}
}

func TestDayBoundaryTypingADaySavesAnOverride(t *testing.T) {
	// Default row → Monday; type 09:00.
	state, action := dbKeys(t, openSchedule(), "down", "0", "9", ":", "0", "0", "ctrl+s")
	if action == nil || action.Kind != "patch_setting" {
		t.Fatalf("expected save, got %+v (%q)", action, state.ErrorMessage)
	}
	got := action.DayBoundarySchedule.WeekdayOverrides
	if got[1] != "09:00" || got[3] != "07:30" || len(got) != 2 {
		t.Fatalf("overrides = %v", got)
	}
}

func TestDayBoundaryResetDropsOverride(t *testing.T) {
	// Default → Mon → Tue → Wed; r resets it.
	_, action := dbKeys(t, openSchedule(), "down", "down", "down", "r", "ctrl+s")
	if action == nil || len(action.DayBoundarySchedule.WeekdayOverrides) != 0 {
		t.Fatalf("r should drop Wednesday's own time, got %+v", action)
	}
}

func TestDayBoundaryBadTimeStaysWithError(t *testing.T) {
	state, action := dbKeys(t, openSchedule(), "down", "9", "9", "ctrl+s")
	if action != nil || state.ErrorMessage == "" || state.Kind == "" {
		t.Fatalf("bad time should stay open with an error, got %+v %q", action, state.ErrorMessage)
	}
	state, _ = dbKeys(t, openSchedule(), "down", "x")
	if state.Inputs[1].Value() != "" {
		t.Fatalf("letters should not type into a time")
	}
}

func TestDayBoundaryToggleAndDiscard(t *testing.T) {
	state, _ := dbKeys(t, openSchedule(), "up", "space")
	if state.DayBoundaryEnabled {
		t.Fatalf("space on Enabled should toggle off")
	}
	state, _ = dbKeys(t, state, "esc")
	if !state.ConfirmDiscard {
		t.Fatalf("esc with changes should ask")
	}
	if gone, _ := dbKeys(t, state, "y"); gone.Kind != "" {
		t.Fatalf("y should discard")
	}
	if closed, _ := dbKeys(t, openSchedule(), "esc"); closed.Kind != "" {
		t.Fatalf("esc unchanged should close")
	}
}

func TestDayBoundaryTypingReplacesAFullTime(t *testing.T) {
	_, action := dbKeys(t, openSchedule(), "0", "5", ":", "3", "0", "ctrl+s")
	if action == nil || action.DayBoundarySchedule.DefaultTime != "05:30" {
		t.Fatalf("typing on the default should replace it, got %+v", action)
	}
}
