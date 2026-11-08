package input

import (
	"testing"

	uistate "crona.work/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func workSummaryState() State {
	return State{
		ActiveView: uistate.ViewDaily, DailyDetailFocus: true, DailyDetailSummary: true,
		DailyWorkRows: []uistate.WorkRow{{Key: "repo:1", Label: "OSS"}, {Key: "repo:2", Label: "Work"}},
	}
}

func workKey(s State, keys ...string) State {
	for _, key := range keys {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		switch key {
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "enter":
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		case "down":
			msg = tea.KeyMsg{Type: tea.KeyDown}
		}
		if next, ok := handleWorkSelect(s, msg); ok {
			s = next
			continue
		}
		s, _ = handleFocusedDailyDetail(s, msg)
	}
	return s
}

func TestWorkBreakdownKeys(t *testing.T) {
	s := workKey(workSummaryState(), "b")
	if s.DailyWorkGroup != "stream" {
		t.Fatalf("b should cycle to stream, got %q", s.DailyWorkGroup)
	}
	s = workKey(workSummaryState(), "B", "down", "enter")
	if !s.DailyWorkSelect || len(s.DailyWorkPath) != 1 || s.DailyWorkPath[0].ID != 2 {
		t.Fatalf("enter should open the second repo, got %+v", s.DailyWorkPath)
	}
	if s = workKey(s, "b"); s.DailyWorkGroup != "" {
		t.Fatalf("b should not regroup inside a drill")
	}
	s = workKey(s, "esc")
	if !s.DailyWorkSelect || len(s.DailyWorkPath) != 0 {
		t.Fatalf("esc should step back to the top, still selecting")
	}
	if s = workKey(s, "esc"); s.DailyWorkSelect {
		t.Fatalf("esc at the top should leave the list")
	}
}

func TestWorkSelectOnlyWhereRowsOpen(t *testing.T) {
	s := workSummaryState()
	s.DailyWorkGroup = "segment_type"
	if s = workKey(s, "B"); s.DailyWorkSelect {
		t.Fatalf("B on segments should do nothing")
	}
	s.DailyWorkGroup = "tag"
	s.DailyWorkRows = []uistate.WorkRow{{Key: "tag:urgent", Label: "+urgent"}}
	s = workKey(s, "B", "enter")
	if len(s.DailyWorkPath) != 1 || s.DailyWorkPath[0].Tag != "urgent" {
		t.Fatalf("enter on a tag should open its issues, got %+v", s.DailyWorkPath)
	}
	s = workKey(s, "esc", "b")
	if s.DailyWorkGroup != "segment_type" || s.DailyWorkSelect {
		t.Fatalf("regrouping to segments should leave the selection")
	}
}
