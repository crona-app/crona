package dialogs

import (
	"strings"
	"testing"

	sharedtypes "crona.work/shared/types"
	controllerpkg "crona.work/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/x/ansi"
)

func editMomentum(step int) controllerpkg.State {
	state := controllerpkg.OpenEditMomentumDirect(controllerpkg.State{}, nil, nil, nil, nil, nil,
		sharedtypes.HabitStreakDefinition{
			ID: "momentum-1", Name: "Recovery Mix", Enabled: true,
			TargetKind: sharedtypes.MomentumTargetKindHabit, Period: sharedtypes.HabitStreakPeriodWeek, RequiredCount: 2,
		})
	state.HabitStreakStep = step
	return state
}

func TestMomentumWizardChrome(t *testing.T) {
	create := controllerpkg.OpenCreateMomentumDirect(controllerpkg.State{}, nil, nil, nil, nil, nil)
	cases := []struct {
		c    formCase
		want []string
	}{
		{
			formCase{name: "create", state: create, title: "New Momentum", context: "1/2 What counts"},
			[]string{"Based on", "Habit completions", "Time in contexts", "Search", "Match", "ctrl+s next"},
		},
		{
			formCase{name: "targets", state: editMomentum(1), title: "Edit Momentum", context: "1/2 What counts"},
			[]string{"Search", "Match", "ctrl+s next"},
		},
		{
			formCase{name: "goal", state: editMomentum(2), title: "Edit Momentum", context: "2/2 Goal"},
			[]string{"Name", "Period", "Required", "a week", "Recovery Mix: 2 completions a week", "ctrl+s save"},
		},
	}
	for _, tc := range cases {
		out := assertListChrome(t, tc.c)
		for _, want := range tc.want {
			if !strings.Contains(out, want) {
				t.Errorf("%s lacks %q:\n%s", tc.c.name, want, out)
			}
		}
		if strings.Contains(out, "(Optional)") || strings.Contains(out, "[ > ") {
			t.Errorf("%s keeps old chrome:\n%s", tc.c.name, out)
		}
	}
}

func TestMomentumGoalReadsAWeek(t *testing.T) {
	out := ansi.Strip(Render(testTheme(), sized(editMomentum(2), 120, 36)))
	if !strings.Contains(out, "a week") || strings.Contains(out, "a weekly") {
		t.Fatalf("goal should read a week:\n%s", out)
	}
}
