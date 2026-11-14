package habitlist

import (
	"strings"
	"testing"

	"crona.work/tui/internal/api"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func TestHabitListCardsAndActions(t *testing.T) {
	target := 30
	state := types.ContentState{
		View:   "habit_list",
		Pane:   "habits",
		Width:  120,
		Height: 20,
		ScopedHabits: []api.HabitWithMeta{
			{
				Habit: api.Habit{
					ID:            1,
					Name:          "Walk",
					ScheduleType:  "weekdays",
					TargetMinutes: &target,
				},
				RepoName:   "crona",
				StreamName: "tui",
			},
		},
	}
	rendered := ansi.Strip(Render(types.Theme{}, state))
	for _, want := range []string{"[a] new habit", "Walk", "[enter] open  [e] edit  [D] delete  [m] log  [+] tags"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %q: %s", want, rendered)
		}
	}
	state.Pane = "scope"
	rendered = ansi.Strip(Render(types.Theme{}, state))
	if !strings.Contains(rendered, "crona › tui | weekdays | 30m") ||
		strings.Contains(rendered, "[a] new habit") {
		t.Fatal(rendered)
	}
	state.Pane = "habits"
	state.Timer = &api.TimerState{State: "running"}
	rendered = Render(types.Theme{}, state)
	if strings.Contains(rendered, "[D]") || strings.Contains(rendered, "[m]") {
		t.Fatal(rendered)
	}
}

func TestHabitLayoutDimensions(t *testing.T) {
	for _, width := range []int{60, 99, 100, 159, 160, 200} {
		state := types.ContentState{View: "habit_list", Pane: "habits", Width: width, Height: 20}
		for i := range 20 {
			state.ScopedHabits = append(
				state.ScopedHabits,
				api.HabitWithMeta{Habit: api.Habit{ID: int64(i), Name: "Walk"}},
			)
		}
		rows := strings.Split(Render(types.Theme{}, state), "\n")
		if len(rows) != state.Height {
			t.Fatalf("width %d: height %d", width, len(rows))
		}
		for _, row := range rows {
			if ansi.StringWidth(row) != width {
				t.Fatalf("width %d: got %d in %q", width, ansi.StringWidth(row), row)
			}
		}
	}
}

func TestHabitPagingKeepsBothIndicators(t *testing.T) {
	state := types.ContentState{Pane: "habits", Cursors: map[string]int{"habits": 9}}
	for i := range 20 {
		state.ScopedHabits = append(
			state.ScopedHabits,
			api.HabitWithMeta{Habit: api.Habit{ID: int64(i), Name: "Walk"}},
		)
	}
	rendered := renderHabits(types.Theme{}, state, 100, 20)
	if !strings.Contains(rendered, "↑") || !strings.Contains(rendered, "↓") ||
		strings.Contains(rendered, "↓ more") {
		t.Fatalf("paged cards lost indicator: %s", rendered)
	}
}

func TestHabitLayoutBeginsWithPanesAndKeepsScopedCount(t *testing.T) {
	for _, width := range []int{60, 120, 200} {
		state := types.ContentState{
			View: "habit_list", Pane: "habits", Width: width, Height: 40,
			ScopedHabits: []api.HabitWithMeta{
				{Habit: api.Habit{ID: 1, Name: "Walk", Active: true}},
				{Habit: api.Habit{ID: 2, Name: "Read", Active: false}},
			},
		}
		text := ansi.Strip(Render(types.Theme{}, state))
		if first := strings.Split(text, "\n")[0]; !strings.Contains(first, "Scope") {
			t.Errorf("width %d: expected panes at the top, got %q", width, first)
		}
		for _, removed := range []string{"Done today", "due today 0", "best streak"} {
			if strings.Contains(text, removed) {
				t.Errorf("width %d: removed aggregate %q remains", width, removed)
			}
		}
		found := false
		for _, line := range strings.Split(text, "\n") {
			if strings.Contains(line, "─ Habits") && strings.Contains(line, "2") {
				found = true
			}
		}
		if !found {
			t.Errorf("width %d: scoped count missing from habit heading", width)
		}
	}
}
