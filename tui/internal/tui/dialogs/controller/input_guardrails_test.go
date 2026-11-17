package controller

import (
	"strings"
	"testing"

	"crona.work/tui/internal/api"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDailyDurationParsingBeforeRounding(t *testing.T) {
	for _, raw := range []string{"24h", "1440m", "23h59m59s"} {
		if _, err := ParseOptionalDurationMinutes(raw, "Screen time"); err != nil {
			t.Errorf("%q: %v", raw, err)
		}
	}
	for _, raw := range []string{"68h", "24h1s", "1440.0001m", "NaN", "Inf", "-1", "9999999999999999999999999"} {
		if _, err := ParseOptionalDurationMinutes(raw, "Screen time"); err == nil {
			t.Errorf("accepted screen time %q", raw)
		}
		if _, err := ParseOptionalDurationHours(raw, "Sleep"); err == nil {
			t.Errorf("accepted sleep %q", raw)
		}
	}
	if value, err := ParseEstimateInput("68h"); err != nil || value == nil || *value != 4080 {
		t.Fatalf("multi-day estimate rejected: %v %v", value, err)
	}
	if _, err := ParseDurationSecondsInput("0", true, "Required work"); err == nil {
		t.Fatal("accepted required zero")
	}
}

func TestCheckInInvalidDurationRetainsTextAndFocus(t *testing.T) {
	state := OpenCreateCheckIn(State{}, "2026-11-17")
	state.Inputs[0].SetValue("3")
	state.Inputs[1].SetValue("3")
	state.Inputs[2].SetValue("68h")
	state.FocusIdx = 5
	next, action, status := Update(state, UpdateContext{}, "2026-11-17", tea.KeyMsg{Type: tea.KeyCtrlS})
	if action != nil || next.FocusIdx != 2 || next.Inputs[2].Value() != "68h" || !strings.Contains(status, "24h") {
		t.Fatalf("invalid check-in: focus=%d action=%v status=%q", next.FocusIdx, action, status)
	}
}

func TestContextPickerListsFilterAndConfirmWithoutSaving(t *testing.T) {
	ctx := UpdateContext{Repos: []api.Repo{{ID: 1, Name: "Alpha"}, {ID: 2, Name: "Long Beta Repo"}}, Streams: []api.Stream{{ID: 11, RepoID: 1, Name: "main"}, {ID: 21, RepoID: 2, Name: "app"}}}
	state := OpenCreateIssueDefault(State{})
	state.FocusIdx = 0
	state, action, status := Update(state, ctx, "2026-11-17", tea.KeyMsg{Type: tea.KeyDown})
	if action != nil || status != "" || state.RepoIndex != 1 {
		t.Fatalf("down: %d %v %q", state.RepoIndex, action, status)
	}
	state, _, _ = Update(state, ctx, "2026-11-17", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Beta")})
	if got := DefaultRepoOptions(state.Inputs, ctx.Repos); len(got) != 2 || got[state.RepoIndex].ID != "2" {
		t.Fatalf("filter selection: %+v", got)
	}
	state, action, status = Update(state, ctx, "2026-11-17", tea.KeyMsg{Type: tea.KeyEnter})
	if action != nil || status != "" || state.FocusIdx != 1 || state.Inputs[0].Value() != "Long Beta Repo" {
		t.Fatalf("enter saved or lost choice: %+v %q", action, status)
	}
	state, action, status = Update(state, ctx, "2026-11-17", tea.KeyMsg{Type: tea.KeyEnter})
	if action != nil || status != "" || state.FocusIdx != 2 || state.Inputs[1].Value() != "app" {
		t.Fatalf("stream enter: %+v %q", action, status)
	}
}

func TestCreateRequiresExplicitPickerConfirmation(t *testing.T) {
	state := OpenCreateIssueDefault(State{})
	state.FocusIdx = 0
	state.Inputs[0].SetValue("New project")
	next, action, status := Update(state, UpdateContext{}, "2026-11-17", tea.KeyMsg{Type: tea.KeyCtrlS})
	if action != nil || !strings.Contains(status, "Create") {
		t.Fatalf("unconfirmed create: %v %q", action, status)
	}
	next, action, status = Update(next, UpdateContext{}, "2026-11-17", tea.KeyMsg{Type: tea.KeyEnter})
	if action != nil || status != "" || next.ConfirmedNewRepo != "New project" || next.FocusIdx != 1 {
		t.Fatalf("explicit create not confirmed: %v %q", action, status)
	}
	next.Inputs[0].SetValue("Alpha")
	options := DefaultRepoOptions(next.Inputs, []api.Repo{{ID: 1, Name: "alpha"}})
	for _, option := range options {
		if option.ID == "__new__" {
			t.Fatal("duplicate Create option for exact match")
		}
	}
}
