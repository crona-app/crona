package controller

import (
	"slices"
	"testing"

	sharedtypes "crona.work/shared/types"

	tea "github.com/charmbracelet/bubbletea"
)

func restKey(t *testing.T, state State, keys ...tea.KeyMsg) State {
	t.Helper()
	for _, key := range keys {
		state, _, _ = Update(state, UpdateContext{}, "2026-11-08", key)
	}
	return state
}

var (
	kDown  = tea.KeyMsg{Type: tea.KeyDown}
	kRight = tea.KeyMsg{Type: tea.KeyRight}
	kSpace = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
	kEsc   = tea.KeyMsg{Type: tea.KeyEsc}
	kRune  = func(r string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(r)} }
)

func openRest() State {
	return OpenEditRestProtection(State{}, nil, nil, []string{"2026-01-01", "2026-12-25"})
}

func TestRestWeekdaysFollowWeekStart(t *testing.T) {
	state := restKey(t, openRest(), kDown, kSpace)
	if !slices.Equal(state.ProtectionWeekdays, []int{1}) {
		t.Fatalf("first slot on a Monday week should be Monday, got %v", state.ProtectionWeekdays)
	}
	sunday := openRest()
	sunday.DueDateWeekStart = sharedtypes.WeekStartSunday
	sunday = restKey(t, sunday, kDown, kSpace)
	if !slices.Equal(sunday.ProtectionWeekdays, []int{0}) {
		t.Fatalf("first slot on a Sunday week should be Sunday, got %v", sunday.ProtectionWeekdays)
	}
}

func TestRestRemovesOnlyUpcomingDates(t *testing.T) {
	state := restKey(t, openRest(), kDown, kDown, kRune("d"))
	if !slices.Equal(state.ProtectionDates, []string{"2026-01-01"}) {
		t.Fatalf("d should remove the shown upcoming date and keep the past one, got %v", state.ProtectionDates)
	}
	if got := RestUpcomingDates([]string{"2026-01-01", "2026-12-25"}, "2026-11-08"); len(got) != 1 {
		t.Fatalf("upcoming = %v", got)
	}
}

func TestRestEscAsksWhenChanged(t *testing.T) {
	if state := restKey(t, openRest(), kEsc); state.Kind != "" {
		t.Fatalf("unchanged esc should close")
	}
	state := restKey(t, openRest(), kRight, kSpace, kEsc)
	if state.Kind == "" || !state.ConfirmDiscard {
		t.Fatalf("changed esc should ask")
	}
	if kept := restKey(t, state, kRune("n")); kept.Kind == "" || kept.ConfirmDiscard {
		t.Fatalf("n should keep editing")
	}
	if gone := restKey(t, state, kRune("y")); gone.Kind != "" {
		t.Fatalf("y should discard")
	}
}

func TestRestSaveKeepsPastDates(t *testing.T) {
	_, action, _ := Update(openRest(), UpdateContext{}, "2026-11-08", tea.KeyMsg{Type: tea.KeyCtrlS})
	if action == nil || action.Kind != "patch_rest_protection" || len(action.RestDates) != 2 {
		t.Fatalf("save should patch with every date, got %+v", action)
	}
}
