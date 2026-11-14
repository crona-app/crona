package model

import (
	"testing"

	"crona.work/tui/internal/api"
	"crona.work/tui/internal/tui/commands"
	tea "github.com/charmbracelet/bubbletea"
)

func previewKey(m Model, key tea.KeyType) Model {
	next, _ := m.Update(tea.KeyMsg{Type: key})
	return next.(Model)
}

func TestExpandedDailyDetailsBrowseAndKeepFocus(t *testing.T) {
	for _, width := range []int{120, 200} {
		m := daySummaryModel(width)
		m = previewKey(m, tea.KeyEnter)
		if !m.dailyDetailFocus || m.detailSessionsIssueID != 1 {
			t.Fatalf("width %d: Enter did not expand issue 1", width)
		}
		m.dailyDetailScroll = 8
		m = previewKey(m, tea.KeyRight)
		if !m.dailyDetailFocus || m.cursor[PaneIssues] != 1 || m.detailSessionsIssueID != 2 || m.dailyDetailScroll != 0 {
			t.Fatalf("width %d: browsing lost focus or selection: cursor=%d history=%d scroll=%d", width, m.cursor[PaneIssues], m.detailSessionsIssueID, m.dailyDetailScroll)
		}
		m = previewKey(m, tea.KeyRight)
		if m.cursor[PaneIssues] != 1 {
			t.Fatal("browsing wrapped past the last item")
		}
		m = previewKey(m, tea.KeyLeft)
		m = previewKey(m, tea.KeyLeft)
		if !m.dailyDetailFocus || m.cursor[PaneIssues] != 0 {
			t.Fatal("left browsing wrapped or closed details")
		}
		m = previewKey(m, tea.KeyEsc)
		if m.dailyDetailFocus || m.cursor[PaneIssues] != 0 || m.detailSessionsIssueID != 0 {
			t.Fatal("Esc did not restore the selected list and clear history target")
		}
	}
}

func TestDailySummaryKeyExpandsAndRestoresSplit(t *testing.T) {
	m := daySummaryModel(200)
	m = previewKey(m, tea.KeyEnter)
	m = pressKey(m, 'S')
	if !m.dailyDetailFocus || !m.dailyDetailSummary || m.detailSessionsIssueID != 0 {
		t.Fatal("S did not expand the summary and stop item history loading")
	}
	m = pressKey(m, 'S')
	if m.dailyDetailFocus || m.dailyDetailSummary {
		t.Fatal("second S did not restore the split")
	}
}

func TestCompactDailyPreviewDoesNotLoadHistoryAndIgnoresStaleReplies(t *testing.T) {
	m := daySummaryModel(200)
	m = previewKey(m, tea.KeyDown)
	if m.detailSessionsIssueID != 0 || m.detailHistoryHabitID != 0 {
		t.Fatal("compact selection requested detail history")
	}
	m = previewKey(m, tea.KeyEnter)
	m = previewKey(m, tea.KeyLeft)
	next, _ := m.Update(commands.DetailIssueHistoryLoadedMsg{IssueID: 2, Sessions: []api.SessionHistoryEntry{{}}})
	m = next.(Model)
	if len(m.detailSessions) != 0 || m.detailSessionsIssueID != 1 {
		t.Fatal("late history response populated the wrong item")
	}
}

func TestEmptyDailySelectionDoesNotExpand(t *testing.T) {
	m := daySummaryModel(200)
	m.allIssues = nil
	m = previewKey(m, tea.KeyEnter)
	if m.dailyDetailFocus {
		t.Fatal("empty task list expanded details")
	}
}

func TestExpandedDailySelectionClampsAfterRemoval(t *testing.T) {
	m := previewKey(daySummaryModel(200), tea.KeyEnter)
	m = previewKey(m, tea.KeyRight)
	m.allIssues = m.allIssues[:1]
	m.syncDetailSessions()
	if !m.dailyDetailFocus || m.cursor[PaneIssues] != 0 || m.detailSessionsIssueID != 1 {
		t.Fatal("removal did not keep details on the nearest remaining item")
	}
	m.allIssues = nil
	m.syncDetailSessions()
	if m.dailyDetailFocus || m.detailSessionsIssueID != 0 {
		t.Fatal("empty list did not close expanded details")
	}
}

func TestExpandedDailyHabitBrowsingAndHistory(t *testing.T) {
	m := daySummaryModel(200)
	m.pane = PaneHabits
	m.dueHabits = []api.HabitDailyItem{
		{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: 10, Name: "Walk"}}},
		{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: 20, Name: "Read"}}},
	}
	m = previewKey(m, tea.KeyEnter)
	if !m.dailyDetailFocus || m.detailHistoryHabitID != 10 {
		t.Fatal("Enter did not expand the selected habit")
	}
	m.dailyDetailScroll = 5
	m = previewKey(m, tea.KeyRight)
	if !m.dailyDetailFocus || m.cursor[PaneHabits] != 1 || m.detailHistoryHabitID != 20 || m.dailyDetailScroll != 0 {
		t.Fatal("habit browsing did not keep focus and reset scroll/history target")
	}
	next, _ := m.Update(commands.DetailHabitHistoryLoadedMsg{HabitID: 10, Completions: []api.HabitCompletion{{}}})
	m = next.(Model)
	if len(m.detailHabitHistory) != 0 {
		t.Fatal("late habit response populated another habit")
	}
	m = previewKey(m, tea.KeyEsc)
	if m.pane != PaneHabits || m.cursor[PaneHabits] != 1 || m.dailyDetailFocus {
		t.Fatal("Esc lost the selected habit")
	}
}

func TestExpandedBrowsingRespectsFilterAndResize(t *testing.T) {
	m := daySummaryModel(200)
	m.filters[PaneIssues] = "Other"
	m = previewKey(m, tea.KeyEnter)
	m = previewKey(m, tea.KeyRight)
	if !m.dailyDetailFocus || m.cursor[PaneIssues] != 0 || m.detailSessionsIssueID != 2 {
		t.Fatal("expanded browse escaped the filtered list")
	}
	for _, width := range []int{120, 200} {
		next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: 40})
		m = next.(Model)
		if !m.dailyDetailFocus || m.detailSessionsIssueID != 2 {
			t.Fatalf("resize to %d lost detail selection/focus", width)
		}
	}
}
