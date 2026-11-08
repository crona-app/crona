package controller

import (
	"strings"
	"testing"

	uistate "crona.work/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

func TestViewJumpGroupsViewsByTab(t *testing.T) {
	state := OpenViewJumpFrom(State{}, uistate.ViewOrder(), uistate.ViewReview)
	var heads []string
	for i, group := range state.ChoiceGroups {
		if group != "" {
			heads = append(heads, strings.ReplaceAll(group, "\t", " ")+":"+state.ChoiceItems[i])
		}
	}
	want := "Today 1:[d] Daily|Work 2:[i] Issues|Insights 3:[r] Review|Log 4:[y] Sessions|System 5:[s] Settings"
	if got := strings.Join(heads, "|"); got != want {
		t.Fatalf("expected tab groups in tab order\n want %s\n got  %s", want, got)
	}
	if !strings.Contains(strings.Join(state.ChoiceItems, ","), "[b] Habits") {
		t.Fatalf("expected habit history listed, got %v", state.ChoiceItems)
	}
	if state.ChoiceValues[state.ChoiceCursor] != "review" || state.ChoiceCurrent != "review" {
		t.Fatalf("expected the cursor on the current view, got %q", state.ChoiceValues[state.ChoiceCursor])
	}
}

func TestViewJumpMovesWithJK(t *testing.T) {
	state := OpenViewJumpFrom(State{}, []uistate.View{uistate.ViewDaily, uistate.ViewDefault, uistate.ViewReview}, uistate.ViewDaily)
	state, _, _ = Update(state, UpdateContext{CurrentDate: "2026-11-08"}, "2026-11-08", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if state.ChoiceValues[state.ChoiceCursor] != "default" {
		t.Fatalf("expected j to move down, got %q", state.ChoiceValues[state.ChoiceCursor])
	}
	state, _, _ = Update(state, UpdateContext{CurrentDate: "2026-11-08"}, "2026-11-08", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if state.ChoiceValues[state.ChoiceCursor] != "daily" {
		t.Fatalf("expected k to move up, got %q", state.ChoiceValues[state.ChoiceCursor])
	}
}
