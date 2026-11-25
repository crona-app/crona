package layout

import (
	"strings"
	"testing"

	dialogstate "crona.work/tui/internal/tui/dialogs/controller"
	uistate "crona.work/tui/internal/tui/state"

	"github.com/charmbracelet/x/ansi"
)

func TestErrorsWrapWithoutHidingControls(t *testing.T) {
	for _, width := range []int{60, 120} {
		for _, overlay := range []bool{false, true} {
			state := shellStateForTest(uistate.ViewDaily, width, 20)
			state.StatusErr = true
			state.StatusMsg = "Could not save because the daemon is unavailable. Check the daemon and submit again."
			if overlay {
				state.DialogOpen = true
				state.DialogState = dialogstate.OpenCreateRepo(dialogstate.State{Width: width, Height: 20})
			}
			rendered := ansi.Strip(Render(state))
			for _, fragment := range []string{"Could not save", "submit again.", "esc"} {
				if !strings.Contains(rendered, fragment) {
					t.Errorf("width=%d overlay=%t missing %q: %q", width, overlay, fragment, rendered)
				}
			}
			lines := strings.Split(rendered, "\n")
			if len(lines) > 20 {
				t.Fatalf("notice exceeded viewport: %d", len(lines))
			}
			for _, line := range lines {
				if ansi.StringWidth(line) > width {
					t.Fatalf("line exceeds width %d: %q", width, line)
				}
			}
		}
	}
}

func TestLongErrorTokensRemainInsideViewport(t *testing.T) {
	state := shellStateForTest(uistate.ViewDaily, 60, 20)
	state.StatusErr = true
	state.StatusMsg = strings.Repeat("界", 200)
	rendered := ansi.Strip(Render(state))
	for _, line := range strings.Split(rendered, "\n") {
		if ansi.StringWidth(line) > 60 {
			t.Fatalf("error exceeded viewport: %q", line)
		}
	}
	if !strings.Contains(rendered, "details in logs") {
		t.Fatalf("long error did not indicate omitted details: %q", rendered)
	}
}

func TestLongInlineErrorKeepsSubmitAndCancelVisible(t *testing.T) {
	state := shellStateForTest(uistate.ViewDaily, 60, 20)
	state.DialogOpen = true
	state.DialogState = dialogstate.OpenCreateRepo(dialogstate.State{Width: 60, Height: 20})
	state.DialogState.ErrorMessage = strings.Repeat("save rejected ", 100)
	rendered := ansi.Strip(Render(state))
	for _, hint := range []string{"ctrl+s", "esc cancel", "details in logs"} {
		if !strings.Contains(rendered, hint) {
			t.Fatalf("long error hid %q: %q", hint, rendered)
		}
	}
}

func TestDialogNoticeDoesNotAdvertiseUnderlyingPaneKeys(t *testing.T) {
	state := shellStateForTest(uistate.ViewDaily, 60, 20)
	state.PaneActions = []string{"[a] new task"}
	state.ErrorNotice = "Could not refresh"
	state.DialogOpen = true
	state.DialogState = dialogstate.OpenCreateRepo(dialogstate.State{Width: 60, Height: 20})
	rendered := ansi.Strip(Render(state))
	if strings.Contains(rendered, "[a] new task") {
		t.Fatal("modal notice advertised inactive shell keys")
	}
}
