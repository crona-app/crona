package model

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"crona.work/tui/internal/api"
	"crona.work/tui/internal/tui/commands"
	viewtypes "crona.work/tui/internal/tui/views/types"
)

func reviewModel() Model {
	return Model{
		view:    ViewReview,
		pane:    PaneReviewDays,
		date:    "2026-10-14",
		width:   200,
		height:  40,
		cursor:  map[Pane]int{},
		filters: map[Pane]string{},
	}
}

func reviewContent(m Model) viewtypes.ContentState {
	snapshot := m.selectionSnapshot()
	return m.viewContentState(m.width, m.height, snapshot, nil)
}

func TestReviewWindowCarriedWhenDateLeaves(t *testing.T) {
	m := reviewModel()
	m.reviewWindowEnd = "2026-10-14"
	s := m.inputState()
	s.Date = "2026-10-20"
	m = m.applyInputState(s)
	if m.reviewWindowEnd != "2026-10-20" {
		t.Fatalf("expected the window to end on the new date, got %q", m.reviewWindowEnd)
	}
	s = m.inputState()
	s.Date = "2026-10-19"
	m = m.applyInputState(s)
	if m.reviewWindowEnd != "2026-10-20" {
		t.Fatalf("expected a date inside the window to keep it, got %q", m.reviewWindowEnd)
	}
}

func TestReviewContentStateUsesOneRange(t *testing.T) {
	m := reviewModel()
	m.reviewWindowDays, m.reviewWindowEnd = 30, "2026-10-20"
	state := reviewContent(m)
	if state.ReviewWindowDays != 30 || state.ReviewWindowEnd != "2026-10-20" ||
		state.RollupStartDate != "2026-09-21" ||
		state.RollupEndDate != "2026-10-20" {
		t.Fatalf(
			"expected one 30-day window ending 2026-10-20, got %d %s %s..%s",
			state.ReviewWindowDays,
			state.ReviewWindowEnd,
			state.RollupStartDate,
			state.RollupEndDate,
		)
	}
}

func TestReviewBottomRowGlobalsOnly(t *testing.T) {
	row := bottomRow(reviewModel())
	for _, want := range []string{"[w] check-in", "[W] away", "[?] all keys"} {
		if !strings.Contains(row, want) {
			t.Fatalf("expected %q on the bottom row, got %q", want, row)
		}
	}
	for _, unwanted := range []string{"[enter]", "[,/.]", "[+/-/0]", "[S]"} {
		if strings.Contains(row, unwanted) {
			t.Fatalf("expected %q off the bottom row, got %q", unwanted, row)
		}
	}
}

func pressReview(m Model, key tea.KeyMsg) Model {
	next, _ := m.Update(key)
	return next.(Model)
}

func TestReviewEnterOpensDayInToday(t *testing.T) {
	m := reviewModel()
	m.width = 80
	m = pressReview(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.view != ViewDaily || m.dialogState().Kind != "" {
		t.Fatalf("expected enter to open Today, got view %s dialog %q", m.view, m.dialogState().Kind)
	}
}

func TestReviewDialogSurvivesResize(t *testing.T) {
	m := reviewModel()
	m.width = 80
	m = pressReview(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 40})
	m = next.(Model)
	if lines := strings.Split(m.View(), "\n"); len(lines) > 40 {
		t.Fatalf("expected the screen to fit after resize, got %d lines", len(lines))
	}
	m = pressReview(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.dialogState().Kind != "" {
		t.Fatalf("expected esc to close the dialog after resize")
	}
}

func TestHelpListsReviewKeys(t *testing.T) {
	body := ansi.Strip(reviewModel().openHelpDialog().dialogState().ViewBody)
	for _, want := range []string{"[S] trends/time", "[w] check-in", "[+/-/0] window", "[↑/↓] day"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected the help sheet to list %q", want)
		}
	}
}

func TestReviewComparisonFollowsRenderedRange(t *testing.T) {
	m := reviewModel()
	start, end := m.reviewRange()
	next, _ := m.Update(commands.ReviewCompareLoadedMsg{Start: start, End: end, Scores: []api.FocusScoreRangeDay{{Date: start}}})
	m = next.(Model)
	if state := reviewContent(m); state.ReviewPrev == nil || len(state.ReviewScores) != 1 {
		t.Fatal("matching comparison was hidden")
	}
	m.reviewWindowDays = 90
	if state := reviewContent(m); state.ReviewPrev != nil || state.ReviewScores != nil {
		t.Fatal("stale comparison exposed after window change")
	}
	next, _ = m.Update(commands.ReviewCompareLoadedMsg{Start: start, End: end})
	m = next.(Model)
	if reviewContent(m).ReviewPrev != nil {
		t.Fatal("obsolete response exposed")
	}
	start, end = m.reviewRange()
	next, _ = m.Update(commands.ReviewCompareLoadedMsg{Start: start, End: end})
	if reviewContent(next.(Model)).ReviewPrev == nil {
		t.Fatal("new comparison was hidden")
	}
}

func TestMomentumRefreshRejectsOldBucket(t *testing.T) {
	m := reviewModel()
	m.momentumBucketWant, m.momentumBucketKey = "old", "old"
	m.momentumBucketDetail = &api.MomentumDetail{}
	oldKey := m.momentumBucketKeyFor(api.MomentumCard{})
	next, _ := m.Update(commands.MomentumRangeLoadedMsg{})
	m = next.(Model)
	if m.momentumBucketDetail != nil || m.momentumBucketWant != "" || m.momentumBucketKey != "" {
		t.Fatal("refresh retained contributor cache outside Momentum")
	}
	if m.momentumBucketKeyFor(api.MomentumCard{}) == oldKey {
		t.Fatal("refresh did not change bucket identity")
	}
	next, _ = m.Update(commands.MomentumBucketLoadedMsg{Key: "old", Detail: &api.MomentumDetail{}})
	if next.(Model).momentumBucketDetail != nil {
		t.Fatal("old bucket response accepted")
	}
	m.momentumBucketWant = "new"
	detail := &api.MomentumDetail{}
	next, _ = m.Update(commands.MomentumBucketLoadedMsg{Key: "new", Detail: detail})
	if next.(Model).momentumBucketDetail != detail {
		t.Fatal("current bucket response rejected")
	}
}
