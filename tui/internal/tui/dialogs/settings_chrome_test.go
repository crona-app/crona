package dialogs

import (
	"strings"
	"testing"

	sharedtypes "crona.work/shared/types"
	controllerpkg "crona.work/tui/internal/tui/dialogs/controller"

	"github.com/charmbracelet/lipgloss"
)

func restProtection(step int) controllerpkg.State {
	state := controllerpkg.OpenEditRestProtection(
		controllerpkg.State{},
		nil,
		[]int{0, 6},
		[]string{"2000-01-01", "2099-12-25"},
	)
	state.ProtectionStep = step
	return state
}

func dayBoundary() controllerpkg.State {
	return controllerpkg.OpenEditDayBoundary(
		controllerpkg.State{},
		sharedtypes.CoreSettingsKey("startOfDay"),
		sharedtypes.DayBoundarySchedule{Enabled: true, DefaultTime: "06:00", WeekdayOverrides: map[int]string{3: "07:30"}},
		"Asia/Kolkata",
	)
}

func telemetry(step int) controllerpkg.State {
	state := controllerpkg.OpenEditTelemetrySettings(controllerpkg.State{}, true, false)
	state.TelemetryStep = step
	return state
}

func TestSettingsDialogs(t *testing.T) {
	cases := []struct {
		c    formCase
		want []string
	}{
		{
			formCase{
				name:  "date format",
				state: controllerpkg.OpenEditDateDisplayFormat(controllerpkg.State{}, "YYYY-MM-DD"),
				title: "Custom Date Format",
			},
			[]string{"Pattern", "ctrl+s save"},
		},
		{
			formCase{
				name:    "protection",
				state:   restProtection(0),
				title:   "Rest & Streak Protection",
				context: "3 streaks · Sat Sun · 1 date",
			},
			[]string{"[x] Habit days", "Su", "Fri 2099-12-25", "+ add date", "1 date in the past", "ctrl+s save"},
		},
		{
			formCase{
				name:    "boundary",
				state:   dayBoundary(),
				title:   "Start of Day",
				context: "Asia/Kolkata",
			},
			[]string{"Enabled", "● on", "Default", "Mo", "07:30   custom", "Su", "ctrl+s save", "r use default"},
		},
		{
			formCase{
				name:    "telemetry usage",
				state:   telemetry(0),
				title:   "Privacy & Diagnostics",
				context: "1/3 Usage",
			},
			[]string{"[x] Share usage signals", "space toggle"},
		},
		{
			formCase{
				name:    "telemetry review",
				state:   telemetry(2),
				title:   "Privacy & Diagnostics",
				context: "3/3 Review",
			},
			[]string{"Usage signals", "Save and Restart Now", "enter choose"},
		},
	}
	for _, tc := range cases {
		out := assertListChrome(t, tc.c)
		for _, want := range tc.want {
			if !strings.Contains(out, want) {
				t.Errorf("%s lacks %q:\n%s", tc.c.name, want, out)
			}
		}
		if strings.Contains(out, "[tab]") || strings.Contains(out, "[space") ||
			strings.Contains(out, "[a] ") {
			t.Errorf("%s keeps bracketed keys:\n%s", tc.c.name, out)
		}
	}
}

func TestWizardsKeepOneWidthAcrossSteps(t *testing.T) {
	steps := map[string][]controllerpkg.State{
		"telemetry": {telemetry(0), telemetry(1), telemetry(2)},
		"momentum":  {editMomentum(1), editMomentum(2), editMomentum(3)},
	}
	for name, states := range steps {
		width := -1
		for i, state := range states {
			top := strings.Split(Render(testTheme(), sized(state, 120, 40)), "\n")[0]
			if got := lipgloss.Width(top); width >= 0 && got != width {
				t.Errorf("%s step %d width %d, want %d", name, i, got, width)
			} else {
				width = got
			}
		}
	}
}
