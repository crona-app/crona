package input

import (
	"testing"

	uistate "crona.work/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func planSummaryState() State {
	return State{
		ActiveView:         uistate.ViewDaily,
		ActivePane:         uistate.PaneIssues,
		DailyDetailFocus:   true,
		DailyDetailSummary: true,
		DailyPlanCount:     3,
	}
}

func TestPOnTheFocusedSummarySelectsThePlan(t *testing.T) {
	pinned := 0
	deps := dateDeps(new(int))
	deps.ToggleSelectedIssuePinnedDaily = func(*State) tea.Cmd { pinned++; return nil }
	next, _ := Handle(planSummaryState(), runeKey('P'), deps)
	if !next.DailyPlanSelect || next.DailyPlanCursor != 0 || pinned != 0 {
		t.Fatalf("expected P to select the plan list without pinning, got select=%v pinned=%d", next.DailyPlanSelect, pinned)
	}
	empty := planSummaryState()
	empty.DailyPlanCount = 0
	if next, _ := Handle(empty, runeKey('P'), deps); next.DailyPlanSelect || pinned != 0 {
		t.Fatal("expected P to do nothing without plan issues")
	}
	// Focused on a task's details rather than the summary, P still pins.
	detail := planSummaryState()
	detail.DailyDetailSummary = false
	if next, _ := Handle(detail, runeKey('P'), deps); next.DailyPlanSelect || pinned != 1 {
		t.Fatalf("expected P to pin from a task's details, got select=%v pinned=%d", next.DailyPlanSelect, pinned)
	}
}

func TestPlanSelectTakesOnlyItsKeys(t *testing.T) {
	statusOpened, dueOpened, other := 0, 0, 0
	deps := dateDeps(new(int))
	deps.OpenPlanIssueStatus = func(*State) bool { statusOpened++; return true }
	deps.OpenPlanIssueTodoDate = func(*State) bool { dueOpened++; return true }
	deps.ToggleSelectedIssuePinnedDaily = func(*State) tea.Cmd { other++; return nil }
	deps.OpenIssueStatusFromSelection = func(*State) bool { other++; return true }
	deps.OpenSelectedIssueTodoDateDialog = func(*State) bool { other++; return true }
	state := planSummaryState()
	state.DailyPlanSelect = true

	state, _ = Handle(state, runeKey('j'), deps)
	state, _ = Handle(state, runeKey('j'), deps)
	state, _ = Handle(state, runeKey('j'), deps)
	if state.DailyPlanCursor != 2 {
		t.Fatalf("expected the cursor held at the last row, got %d", state.DailyPlanCursor)
	}
	state, _ = Handle(state, runeKey('k'), deps)
	if state.DailyPlanCursor != 1 {
		t.Fatalf("expected k to move up, got %d", state.DailyPlanCursor)
	}
	Handle(state, runeKey('s'), deps)
	Handle(state, runeKey('d'), deps)
	if statusOpened != 1 || dueOpened != 1 {
		t.Fatalf("expected s and d on the plan issue, got status=%d due=%d", statusOpened, dueOpened)
	}
	for _, r := range []rune{'P', 'D', 'e', 'f', 'm', 'x'} {
		if next, _ := Handle(state, runeKey(r), deps); !next.DailyPlanSelect || next.Dialog != "" {
			t.Fatalf("expected %q held back in the plan list", r)
		}
	}
	if other != 0 {
		t.Fatalf("expected no Tasks action from the plan list, got %d", other)
	}
	state, _ = Handle(state, tea.KeyMsg{Type: tea.KeyEsc}, deps)
	if state.DailyPlanSelect || !state.DailyDetailFocus {
		t.Fatal("expected esc to leave the plan list but keep the summary focused")
	}
}

func TestPlanSelectEndsWhenTabLeavesThePane(t *testing.T) {
	state := planSummaryState()
	state.DailyPlanSelect = true
	next, _ := Handle(state, tea.KeyMsg{Type: tea.KeyTab}, dateDeps(new(int)))
	if next.DailyPlanSelect || next.DailyDetailFocus {
		t.Fatalf("expected tab to leave the pane and the plan list, got select=%v focus=%v", next.DailyPlanSelect, next.DailyDetailFocus)
	}
}
