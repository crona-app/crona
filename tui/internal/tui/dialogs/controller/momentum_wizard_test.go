package controller

import (
	"testing"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"

	tea "github.com/charmbracelet/bubbletea"
)

func momentumHabits() []api.HabitWithMeta {
	habit := func(id int64, name, repo, stream string, tags ...string) api.HabitWithMeta {
		return api.HabitWithMeta{
			Habit:  sharedtypes.Habit{ID: id, Name: name, ScheduleType: sharedtypes.HabitScheduleDaily, Tags: tags},
			RepoID: 1, RepoName: repo, StreamName: stream,
		}
	}
	return []api.HabitWithMeta{
		habit(1, "Journal", "Life", "Mind"),
		habit(2, "Run", "Life", "Body", "cardio"),
		habit(3, "Read", "Life", "Mind"),
	}
}

func momentumKeys(t *testing.T, state State, keys ...string) (State, *Action) {
	t.Helper()
	var action *Action
	for _, key := range keys {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		switch key {
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "tab":
			msg = tea.KeyMsg{Type: tea.KeyTab}
		case "shift+tab":
			msg = tea.KeyMsg{Type: tea.KeyShiftTab}
		case "up":
			msg = tea.KeyMsg{Type: tea.KeyUp}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		case "right":
			msg = tea.KeyMsg{Type: tea.KeyRight}
		case "space":
			msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
		case "ctrl+s":
			msg = tea.KeyMsg{Type: tea.KeyCtrlS}
		}
		state, action, _ = Update(state, UpdateContext{}, "2026-05-26", msg)
	}
	return state, action
}

func newMomentum() State {
	return OpenCreateMomentumDirect(State{}, nil, momentumHabits(), nil, nil, nil)
}

func TestMomentumEscClosesWhenUnchanged(t *testing.T) {
	next, _ := momentumKeys(t, newMomentum(), "esc")
	if next.Kind != "" {
		t.Fatalf("esc on an untouched wizard should close, kind %q", next.Kind)
	}
}

func TestMomentumEscAsksBeforeDiscarding(t *testing.T) {
	state, _ := momentumKeys(t, newMomentum(), "down", "space")
	state, _ = momentumKeys(t, state, "esc")
	if state.Kind == "" || !state.Momentum.ConfirmDiscard {
		t.Fatalf("esc with a pick should ask first")
	}
	kept, _ := momentumKeys(t, state, "n")
	if kept.Kind == "" || kept.Momentum.ConfirmDiscard || len(kept.HabitStreakDraft.HabitIDs) != 1 {
		t.Fatalf("n should keep editing with the pick")
	}
	gone, _ := momentumKeys(t, state, "y")
	if gone.Kind != "" {
		t.Fatalf("y should discard")
	}
}

func TestMomentumGroupsHabitsByContext(t *testing.T) {
	rows := MomentumHabitRows(newMomentum())
	want := []string{"Life › Mind", "Journal", "Read", "Life › Body", "Run"}
	if len(rows) != len(want) {
		t.Fatalf("rows %+v", rows)
	}
	for i, label := range want {
		if rows[i].Label != label || rows[i].Heading != (i == 0 || i == 3) {
			t.Fatalf("row %d = %+v, want %q", i, rows[i], label)
		}
	}
}

func TestMomentumHeadingTogglesItsGroup(t *testing.T) {
	state, _ := momentumKeys(t, newMomentum(), "space")
	if len(state.HabitStreakDraft.HabitIDs) != 2 {
		t.Fatalf("heading should pick its group, got %v", state.HabitStreakDraft.HabitIDs)
	}
	state, _ = momentumKeys(t, state, "space")
	if len(state.HabitStreakDraft.HabitIDs) != 0 {
		t.Fatalf("heading on a full group should clear it, got %v", state.HabitStreakDraft.HabitIDs)
	}
}

func TestMomentumSearchMatchesContextAndTags(t *testing.T) {
	state, _ := momentumKeys(t, newMomentum(), "/", "c", "a", "r", "d")
	if visible, total, _ := MomentumHabitCounts(state); visible != 1 || total != 3 {
		t.Fatalf("tag search shows %d of %d", visible, total)
	}
	state, _ = momentumKeys(t, newMomentum(), "/", "m", "i", "n", "d", "enter", "a")
	if len(state.HabitStreakDraft.HabitIDs) != 2 {
		t.Fatalf("a should pick only the shown habits, got %v", state.HabitStreakDraft.HabitIDs)
	}
	state.HabitStreakDraft.HabitIDs = append(state.HabitStreakDraft.HabitIDs, 2)
	if _, _, hidden := MomentumHabitCounts(state); hidden != 1 {
		t.Fatalf("hidden picks = %d, want 1", hidden)
	}
	state, _ = momentumKeys(t, state, "c")
	if ids := state.HabitStreakDraft.HabitIDs; len(ids) != 1 || ids[0] != 2 {
		t.Fatalf("c should clear only the shown habits, got %v", ids)
	}
	state, _ = momentumKeys(t, state, "esc")
	if state.Momentum.Filter.Value() != "" || state.Kind == "" {
		t.Fatalf("esc should clear the search before anything else")
	}
}

func TestMomentumArrowsNeverEnterSearch(t *testing.T) {
	state, _ := momentumKeys(t, newMomentum(), "up")
	if state.Momentum.FilterFocused || !state.Momentum.KindFocused {
		t.Fatalf("up from the list top should reach the kind row, not search")
	}
	state, _ = momentumKeys(t, state, "up", "down", "down")
	if state.Momentum.FilterFocused {
		t.Fatalf("arrows should never focus search")
	}
	state, _ = momentumKeys(t, state, "/", "up")
	if state.Momentum.FilterFocused {
		t.Fatalf("up should leave search")
	}
}

func TestMomentumKindSwitchKeepsPicks(t *testing.T) {
	state, _ := momentumKeys(t, newMomentum(), "down", "space", "shift+tab", "right")
	if !momentumContextKind(state) || len(state.HabitStreakDraft.HabitIDs) != 1 {
		t.Fatalf("switching to contexts should keep habit picks")
	}
	if state.HabitStreakDraft.RequiredCount != 3600 {
		t.Fatalf("required should reset to the context default, got %d", state.HabitStreakDraft.RequiredCount)
	}
	state, _ = momentumKeys(t, state, "right")
	if momentumContextKind(state) || len(state.HabitStreakDraft.HabitIDs) != 1 {
		t.Fatalf("switching back should still have the pick")
	}
}

func TestMomentumCreateSavesFromGoal(t *testing.T) {
	state, _ := momentumKeys(t, newMomentum(), "down", "space", "ctrl+s")
	if state.HabitStreakStep != 2 {
		t.Fatalf("ctrl+s should reach the goal, step %d (%q)", state.HabitStreakStep, state.ErrorMessage)
	}
	state, action := momentumKeys(t, state, "ctrl+s")
	if action != nil || state.ErrorMessage != "Name is required" {
		t.Fatalf("save without a name should fail, got %+v %q", action, state.ErrorMessage)
	}
	state, _ = momentumKeys(t, state, "D", "e", "e", "p")
	state, action = momentumKeys(t, state, "ctrl+s")
	if action == nil || action.Kind != "create_momentum" || action.HabitStreakDefs[0].Name != "Deep" {
		t.Fatalf("expected create_momentum, got %+v", action)
	}
	if state.Kind != "" {
		t.Fatalf("save should close")
	}
}
