package dialogs

import (
	"strings"
	"testing"

	controllerpkg "crona.work/tui/internal/tui/dialogs/controller"
	uistate "crona.work/tui/internal/tui/state"

	"github.com/charmbracelet/x/ansi"
)

func TestViewJumpRendersGroupsAndASystemGrid(t *testing.T) {
	state := controllerpkg.OpenViewJumpFrom(controllerpkg.State{}, uistate.ViewOrder(), uistate.ViewDaily)
	state.Width, state.Height = 120, 50
	out := ansi.Strip(renderViewJump(Theme{}, state))
	lines := strings.Split(out, "\n")
	order := []string{"Today", "Work", "Insights", "Log", "System"}
	last := -1
	for _, head := range order {
		at := strings.Index(out, head)
		if at < 0 || at < last {
			t.Fatalf("expected %q in tab order, got\n%s", head, out)
		}
		last = at
	}
	for _, want := range []string{"▶ d  Daily ●", "planned issues and habits", "b  Habits"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q, got\n%s", want, out)
		}
	}
	grid := false
	for _, line := range lines {
		if strings.Contains(line, "s  Settings") && strings.Contains(line, "l  Alerts") && strings.Contains(line, "p  Reports") {
			grid = true
		}
	}
	if !grid {
		t.Fatalf("expected System three to a row, got\n%s", out)
	}
	for _, gone := range []string{"Press a mnemonic", "↑/↓ move", "key jump", "History"} {
		if strings.Contains(out, gone) {
			t.Fatalf("expected no %q, got\n%s", gone, out)
		}
	}
}
