package dialogs

import (
	"fmt"
	"strings"
	"testing"

	sharedtypes "crona.work/shared/types"
	controller "crona.work/tui/internal/tui/dialogs/controller"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestFormHintsFitAndMatchKeyboardActions(t *testing.T) {
	for _, size := range [][2]int{{60, 20}, {120, 36}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			base := controller.State{Width: size[0], Height: size[1]}
			check := func(state controller.State, want ...string) string {
				t.Helper()
				frame := Render(Theme{}, state)
				if width, height := lipgloss.Size(frame); width > size[0] || height > size[1] {
					t.Errorf("%s renders at %dx%d", state.Kind, width, height)
				}
				plain := ansi.Strip(frame)
				for _, fragment := range want {
					if !strings.Contains(plain, fragment) {
						t.Errorf("%s missing %q:\n%s", state.Kind, fragment, plain)
					}
				}
				return plain
			}
			press := func(state controller.State, key tea.KeyType) (controller.State, *controller.Action, string) {
				return controller.Update(state, controller.UpdateContext{}, "2026-11-25", tea.KeyMsg{Type: key})
			}
			defaultIssue := controller.OpenCreateIssueDefault(base)
			defaultIssue, _, _ = press(defaultIssue, tea.KeyShiftTab)
			defaultIssue.StreamOptionLabels = []string{"One", "Two", "Three", "Four", "Five"}
			defaultIssue.StreamIndex = 4
			check(defaultIssue, "Stream", "▶ Five", "enter select", "ctrl+s save & close", "esc cancel")

			issue := controller.OpenCreateIssueMeta(base, 1, "Core", "Crona")
			check(issue, "ctrl+s save & close", "esc cancel", "tab next", "shift+tab back")
			issue.Inputs[0].SetValue("Keep my issue title")
			issue.Inputs[1].SetValue("bad")
			issue, action, status := press(issue, tea.KeyCtrlS)
			if action != nil || status == "" || issue.Inputs[0].Value() != "Keep my issue title" {
				t.Fatal("invalid issue estimate should preserve the entered title")
			}
			issue.Inputs[1].SetValue("30m")
			if _, action, _ := press(issue, tea.KeyCtrlS); action == nil || action.Title != "Keep my issue title" {
				t.Fatal("Ctrl+S should save the corrected issue")
			}

			countdown := controller.OpenSingleTimerStart(base, 1, 1, 1, "Audit issue", nil, 0)
			check(countdown, "ctrl+s start", "esc back")
			if next, _, _ := press(countdown, tea.KeyEsc); next.Kind != "timer_start_type" {
				t.Fatalf("Escape from countdown opened %q", next.Kind)
			}
			countdown.Inputs[0].SetValue("bad")
			countdown, action, status = press(countdown, tea.KeyCtrlS)
			if action != nil || status == "" || countdown.Inputs[0].Value() != "bad" {
				t.Fatal("invalid countdown should stay open and preserve the entered time")
			}
			countdown.Inputs[0].SetValue("10m")
			if _, action, _ := press(countdown, tea.KeyCtrlS); action == nil || action.TimerStart == nil ||
				action.TimerStart.HardLimitTotalSeconds == nil || *action.TimerStart.HardLimitTotalSeconds != 600 {
				t.Fatal("Ctrl+S should start the corrected countdown")
			}
			pomodoro := controller.OpenPomodoroStart(base, 1, 1, 1, "Audit issue", nil, 0)
			check(pomodoro, "ctrl+s start", "esc back", "tab next", "shift+tab back")
			if next, _, _ := press(pomodoro, tea.KeyTab); next.FocusIdx == pomodoro.FocusIdx {
				t.Fatal("Tab did not move Pomodoro focus")
			}

			checkin := controller.OpenCreateCheckIn(base, "2026-11-25")
			check(checkin, "ctrl+s save", "esc cancel", "tab next", "shift+tab back")
			checkin.Inputs[0].SetValue("4")
			checkin.Inputs[1].SetValue("9")
			checkin, action, status = press(checkin, tea.KeyCtrlS)
			if action != nil || status == "" || checkin.Inputs[0].Value() != "4" {
				t.Fatal("invalid check-in energy should preserve the entered mood")
			}
			checkin.ErrorMessage = status
			check(checkin, status, "ctrl+s save", "esc cancel")
			checkin.Inputs[1].SetValue("3")
			if _, action, _ := press(checkin, tea.KeyCtrlS); action == nil || action.Mood != 4 || action.Energy != 3 {
				t.Fatal("Ctrl+S should save the corrected check-in")
			}

			day := controller.OpenEditDayBoundary(base, sharedtypes.CoreSettingsKeyStartOfDay,
				sharedtypes.DayBoundarySchedule{Enabled: true, DefaultTime: "09:00"}, "UTC")
			if frame := check(day, "ctrl+s save", "esc cancel", "tab next"); strings.Contains(frame, "r use default") {
				t.Error("default row should not offer weekday reset")
			}
			weekday, _, _ := press(day, tea.KeyTab)
			check(weekday, "r use default", "ctrl+s save", "esc cancel")
			day.Inputs[0].SetValue("10:00")
			confirm, _, _ := press(day, tea.KeyEsc)
			check(confirm, "Discard changes?", "y discard", "esc keep editing")
			if next, _, _ := press(confirm, tea.KeyEsc); next.ConfirmDiscard || next.Inputs[0].Value() != "10:00" {
				t.Fatal("Escape from discard confirmation should keep the edited time")
			}
			privacy := controller.OpenEditTelemetrySettings(base, false, false)
			privacy, _, _ = press(privacy, tea.KeyTab)
			privacy, _, _ = press(privacy, tea.KeyTab)
			check(privacy, "enter save", "esc cancel")
			if _, action, _ := press(privacy, tea.KeyEnter); action == nil || action.Kind != "patch_telemetry_settings" {
				t.Fatal("Enter should save the selected privacy choices")
			}
		})
	}
}
