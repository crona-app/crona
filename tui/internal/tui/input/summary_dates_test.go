package input

import "testing"

func TestDayKeysWorkFromTheFocusedSummary(t *testing.T) {
	for _, selecting := range []bool{false, true} {
		reloads := 0
		deps := dateDeps(&reloads)
		state := planSummaryState()
		state.DailyPlanSelect = selecting
		state.Date = "2026-07-30"
		next, _ := Handle(state, runeKey(','), deps)
		if next.Date != "2026-07-29" {
			t.Fatalf("selecting=%v: expected , to go back a day, got %q", selecting, next.Date)
		}
		next, _ = Handle(next, runeKey('.'), deps)
		next, _ = Handle(next, runeKey('.'), deps)
		if next.Date != "2026-07-31" {
			t.Fatalf("selecting=%v: expected . to go forward a day, got %q", selecting, next.Date)
		}
		next, _ = Handle(next, runeKey('g'), deps)
		if next.Date != "2026-07-30" {
			t.Fatalf("selecting=%v: expected g to return to today, got %q", selecting, next.Date)
		}
		if !next.DailyDetailFocus {
			t.Fatalf("selecting=%v: expected the summary to keep focus", selecting)
		}
	}
}
