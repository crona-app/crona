package layout

import (
	"strings"
	"testing"

	dialogstate "crona.work/tui/internal/tui/dialogs/controller"
	uistate "crona.work/tui/internal/tui/state"
)

func TestOnboardingAtMinimumSize(t *testing.T) {
	for step, page := range []struct {
		name string
		want []string
	}{
		{"welcome", []string{"What to expect", "privacy preferences", "[←/→]"}},
		{"workspace", []string{"Your workspace", "Everything remains local", "[←/→]"}},
		{"privacy", []string{"Privacy choices", "Share usage signals", "Share diagnostics", "[space]", "[←/→]"}},
		{"review", []string{"Review your choices", "Start Crona", "Start and Restart Now", "[enter]", "[←/→]"}},
	} {
		t.Run(page.name, func(t *testing.T) {
			state := shellStateForTest(uistate.ViewDaily, MinWidth, MinHeight)
			state.DialogOpen = true
			state.DialogState = dialogstate.OpenOnboarding(dialogstate.State{Width: MinWidth, Height: MinHeight}, false, false)
			state.DialogState.TelemetryStep = step
			frame := Render(state)
			for _, fragment := range append(page.want, "Let's get things set up.", "4.Review") {
				if !strings.Contains(frame, fragment) {
					t.Errorf("missing %q in frame:\n%s", fragment, frame)
				}
			}
		})
	}
}
