package dialogs

import (
	"strings"
	"testing"

	controllerpkg "crona.work/tui/internal/tui/dialogs/controller"
)

func TestFocusBlockedDialogExplainsStatusAndRecovery(t *testing.T) {
	state := controllerpkg.OpenFocusBlocked(
		controllerpkg.State{},
		22,
		"Ship onboarding",
		"backlog",
		"ready",
	)
	rendered := Render(testTheme(), state)
	for _, want := range []string{
		"Can't focus yet",
		"Ship onboarding is in backlog",
		"Focus sessions start from Planned, Ready, or In Progress",
		"Mark Ready and start",
		"Cancel",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("dialog missing %q in %q", want, rendered)
		}
	}
}
