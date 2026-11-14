package habitlist

import (
	"strings"
	"testing"

	"crona.work/tui/internal/api"
	types "crona.work/tui/internal/tui/views/types"
	"github.com/charmbracelet/x/ansi"
)

func TestWorkHabitDetailsShareArcsAndActions(t *testing.T) {
	state := types.ContentState{
		View: "habit_list", Pane: "habits", DashboardDate: "2026-11-15",
		Cursors:              map[string]int{"habits": 0},
		ScopedHabits:         []api.HabitWithMeta{{Habit: api.Habit{ID: 10, Name: "Walk", Active: true, TargetMinutes: new(30)}}},
		DetailHistoryHabitID: 10,
		DetailHabitHistory:   []api.HabitCompletion{{HabitID: 10, Date: "2026-11-15", Status: "completed"}},
	}
	text := ansi.Strip(renderDetail(types.Theme{}, state, 76, 40))
	for _, want := range []string{"Hit/7d", "Hit/30d", "Recent", "[m] log", "[e] edit", "[D] delete", "[enter] expand"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "[x] done") || strings.Contains(text, "[F] fail") {
		t.Fatal("Work habit definition advertised Daily-only completion actions")
	}
	lines := strings.Split(text, "\n")
	if !strings.Contains(lines[len(lines)-1], "[D] delete") {
		t.Fatal("general habit actions are not pinned to the bottom")
	}
}
