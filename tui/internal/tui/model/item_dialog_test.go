package model

import (
	"fmt"
	"strings"
	"testing"

	"crona.work/tui/internal/api"
	"crona.work/tui/internal/tui/commands"
	"crona.work/tui/internal/tui/dialogs"
	layoutpkg "crona.work/tui/internal/tui/layout"
	"crona.work/tui/internal/tui/views/habitlist"
	"crona.work/tui/internal/tui/views/issues"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func issueDialogModel(width int) Model {
	m := daySummaryModel(width)
	m.view, m.height = ViewDefault, 30
	m.allIssues[0].Title = "A Today"
	m.allIssues[1].Title = "B Other"
	m.allIssues[0].WorkedSeconds = 1800
	m.allIssues[0].EstimateMinutes = new(60)
	return m
}

func TestIssueModalSharesDetailsAndRoutesContextualActions(t *testing.T) {
	m := issueDialogModel(80)
	m = previewKey(m, tea.KeyEnter)
	if !m.isItemDetailDialog() || m.detailSessionsIssueID != 1 {
		t.Fatalf("Enter did not open issue modal/history target: %s %d", m.dialog, m.detailSessionsIssueID)
	}
	body := ansi.Strip(m.dialogViewBody)
	for _, want := range []string{"Effort/1h", "30m", "[s] status", "[d] due", "Sessions", "[m] log"} {
		if !strings.Contains(body, want) {
			t.Errorf("issue modal missing %q:\n%s", want, body)
		}
	}
	if !strings.Contains(strings.Join(m.dialogViewDetailHints, " "), "e edit") {
		t.Fatal("edit missing from modal footer")
	}
	m = pressKey(m, 's')
	if m.dialog != "issue_status" {
		t.Fatalf("status key opened %q, want issue_status", m.dialog)
	}
}

func TestItemModalBrowsingResizeAndStaleHistory(t *testing.T) {
	m := previewKey(issueDialogModel(80), tea.KeyEnter)
	m.dialogViewScroll = 5
	m = previewKey(m, tea.KeyRight)
	if !m.isItemDetailDialog() || m.detailSessionsIssueID != 2 || m.dialogViewName != "B Other" || m.dialogViewScroll != 0 {
		t.Fatalf("modal browse failed: %s name=%s target=%d scroll=%d", m.dialog, m.dialogViewName, m.detailSessionsIssueID, m.dialogViewScroll)
	}
	m = previewKey(m, tea.KeyRight)
	if m.cursor[PaneIssues] != 1 {
		t.Fatal("modal browsing wrapped at list boundary")
	}
	next, _ := m.Update(commands.DetailIssueHistoryLoadedMsg{IssueID: 1, Sessions: []api.SessionHistoryEntry{{}}})
	m = next.(Model)
	if len(m.detailSessions) != 0 {
		t.Fatal("modal accepted history for previous item")
	}
	next, _ = m.Update(tea.WindowSizeMsg{Width: 50, Height: 24})
	m = next.(Model)
	if !m.isItemDetailDialog() || m.dialogViewName != "B Other" {
		t.Fatal("resize lost the item modal")
	}
	m = previewKey(m, tea.KeyEsc)
	if m.dialog != "" || m.cursor[PaneIssues] != 1 || m.detailSessionsIssueID != 0 {
		t.Fatal("Esc lost the selection or kept modal history active")
	}
}

func TestDailyHabitModalSharesHitArcsAndContextualKeys(t *testing.T) {
	m := daySummaryModel(80)
	m.pane = PaneHabits
	m.dueHabits = []api.HabitDailyItem{{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: 10, Name: "Walk", TargetMinutes: new(30)}}, DurationMinutes: new(15)}}
	var ok bool
	m, ok = m.openSelectedViewDialog()
	if !ok || !m.isItemDetailDialog() {
		t.Fatal("habit modal did not open")
	}
	m.syncDetailSessions()
	next, _ := m.Update(commands.DetailHabitHistoryLoadedMsg{HabitID: 10, Generation: m.detailHabitHistoryGeneration, Completions: []api.HabitCompletion{
		{HabitID: 10, Date: m.currentDate(), Status: "completed"},
	}})
	m = next.(Model)
	body := ansi.Strip(m.dialogViewBody)
	for _, want := range []string{"Today/30m", "15m", "Hit/7d", "Hit/30d", "[x] done", "[F] fail", "Recent", "[m] log"} {
		if !strings.Contains(body, want) {
			t.Errorf("habit modal missing %q:\n%s", want, body)
		}
	}
	m = pressKey(m, 'e')
	if m.dialog != "edit_habit" {
		t.Fatalf("habit edit key opened %q", m.dialog)
	}
}

func TestHabitDefinitionModalLoadsHistoryWithoutDailyCompletionKeys(t *testing.T) {
	m := scopeModel(1)
	m.view, m.pane, m.width, m.height = ViewHabitList, PaneHabits, 80, 30
	m.allHabits = []api.HabitWithMeta{{Habit: api.Habit{ID: 10, Name: "Walk", StreamID: 4, Active: true}, RepoID: 1}}
	m = previewKey(m, tea.KeyEnter)
	if !m.isItemDetailDialog() || m.detailHistoryHabitID != 10 {
		t.Fatal("habit definition modal did not request history")
	}
	body := ansi.Strip(m.dialogViewBody)
	if strings.Contains(body, "[x] done") || strings.Contains(body, "[F] fail") {
		t.Fatal("definition modal advertises unsupported daily completion keys")
	}
	m = pressKey(m, 'e')
	if m.dialog != "edit_habit" {
		t.Fatalf("definition edit key opened %q", m.dialog)
	}
}

func TestItemModalSmallScreenLayoutAndScrolling(t *testing.T) {
	for _, size := range [][2]int{{40, 20}, {60, 24}, {80, 30}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			m := issueDialogModel(size[0])
			m.height = size[1]
			m.allIssues[0].Description = new(strings.Repeat("Description content stays first. ", 50))
			m.allIssues[0].Notes = new("Unique tail note")
			m = previewKey(m, tea.KeyEnter)
			check := func(m Model) string {
				t.Helper()
				out := ansi.Strip(dialogs.Render(layoutpkg.DialogTheme(), m.dialogState()))
				lines := strings.Split(out, "\n")
				if len(lines) > size[1] {
					t.Errorf("modal height %d exceeds %d", len(lines), size[1])
				}
				for _, line := range lines {
					if ansi.StringWidth(line) > size[0] {
						t.Errorf("modal row exceeds width %d: %q", size[0], line)
					}
				}
				return out
			}
			top := check(m)
			if !strings.Contains(top, "Description content") || !strings.Contains(top, "[e] edit") {
				t.Errorf("missing content-first body or pinned actions:\n%s", top)
			}
			m = pressKey(m, 'G')
			bottom := check(m)
			if !strings.Contains(bottom, "Unique tail note") || !strings.Contains(bottom, "[e] edit") {
				t.Errorf("bottom content or fixed actions missing:\n%s", bottom)
			}
			t.Logf("top:\n%s\nbottom:\n%s", top, bottom)
		})
	}
}

func TestWorkPanesLoadAndRetainSelectedHistory(t *testing.T) {
	m := issueDialogModel(200)
	m.syncDetailSessions()
	if m.detailSessionsIssueID != 1 {
		t.Fatal("visible Work issue pane did not target selection history")
	}
	m = previewKey(m, tea.KeyDown)
	if m.detailSessionsIssueID != 2 {
		t.Fatal("Work history did not follow list selection")
	}
	m.pane = PaneScope
	m.syncDetailSessions()
	if m.detailSessionsIssueID != 2 {
		t.Fatal("scope focus discarded the visible issue pane's history")
	}
	m.width, m.height = 80, 20
	m.syncDetailSessions()
	if m.detailSessionsIssueID != 0 {
		t.Fatal("hidden Work detail pane kept requesting history")
	}
	m = scopeModel(1)
	m.view, m.pane, m.width, m.height = ViewHabitList, PaneHabits, 200, 40
	m.allHabits = []api.HabitWithMeta{{Habit: api.Habit{ID: 10, Name: "Walk", StreamID: 4, Active: true}, RepoID: 1}}
	m.syncDetailSessions()
	if m.detailHistoryHabitID != 10 {
		t.Fatal("visible Work habit pane did not target its history")
	}
	m.pane = PaneScope
	m.syncDetailSessions()
	if m.detailHistoryHabitID != 10 {
		t.Fatal("scope focus discarded visible habit history")
	}
}

func TestWorkPaneAndModalUseIdenticalDetailContent(t *testing.T) {
	issue := issueDialogModel(200)
	habit := scopeModel(1)
	habit.view, habit.pane, habit.width, habit.height = ViewHabitList, PaneHabits, 200, 40
	habit.allHabits = []api.HabitWithMeta{{Habit: api.Habit{ID: 10, Name: "Walk", StreamID: 4, Active: true}, RepoID: 1}}
	for _, m := range []Model{issue, habit} {
		m.syncDetailSessions()
		content, ok := m.itemDialogContent()
		if !ok {
			t.Fatal("modal content was unavailable")
		}
		state, width := m.detailContentState(m.selectionSnapshot()), viewEntityBodyWidth(m.width)
		var title string
		var lines []string
		if m.pane == PaneIssues {
			title, lines, ok = issues.Detail(layoutpkg.ViewTheme(), state, width)
		} else {
			title, lines, ok = habitlist.Detail(layoutpkg.ViewTheme(), state, width)
		}
		if !ok || title != content.title || ansi.Strip(strings.Join(lines, "\n")) != ansi.Strip(strings.Join(content.body, "\n")) {
			t.Errorf("%s pane and modal differ", content.kind)
		}
	}
}

func TestDetailHistoryRejectsOlderSameItemGenerations(t *testing.T) {
	m := issueDialogModel(200)
	m.syncDetailSessions()
	oldGeneration := m.detailSessionsGeneration
	m.detailSessionsIssueID = 0 // A session change invalidates the in-flight request.
	m.syncDetailSessions()
	if m.detailSessionsGeneration == oldGeneration {
		t.Fatal("same-item history refresh did not advance request generation")
	}
	next, _ := m.Update(commands.DetailIssueHistoryLoadedMsg{
		IssueID: 1, Generation: m.detailSessionsGeneration,
		Sessions: []api.SessionHistoryEntry{{Session: api.Session{ID: "new"}}},
	})
	m = next.(Model)
	next, _ = m.Update(commands.DetailIssueHistoryLoadedMsg{
		IssueID: 1, Generation: oldGeneration,
		Sessions: []api.SessionHistoryEntry{{Session: api.Session{ID: "old"}}},
	})
	m = next.(Model)
	if len(m.detailSessions) != 1 || m.detailSessions[0].ID != "new" {
		t.Fatal("older same-issue response overwrote refreshed sessions")
	}
	m = daySummaryModel(200)
	m.pane, m.dailyDetailFocus = PaneHabits, true
	m.dueHabits = []api.HabitDailyItem{{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: 10, Name: "Walk"}}}}
	m.syncDetailSessions()
	oldGeneration = m.detailHabitHistoryGeneration
	m.detailHistoryHabitID = 0 // A completion change invalidates the in-flight request.
	m.syncDetailSessions()
	if m.detailHabitHistoryGeneration == oldGeneration {
		t.Fatal("same-habit refresh did not advance request generation")
	}
	next, _ = m.Update(commands.DetailHabitHistoryLoadedMsg{
		HabitID: 10, Generation: m.detailHabitHistoryGeneration,
		Completions: []api.HabitCompletion{{HabitID: 10, Date: m.currentDate(), Status: "completed"}},
	})
	m = next.(Model)
	next, _ = m.Update(commands.DetailHabitHistoryLoadedMsg{
		HabitID: 10, Generation: oldGeneration,
		Completions: []api.HabitCompletion{{HabitID: 10, Date: m.currentDate(), Status: "failed"}},
	})
	m = next.(Model)
	if len(m.detailHabitHistory) != 1 || m.detailHabitHistory[0].Status != "completed" {
		t.Fatal("older same-habit response overwrote refreshed completions")
	}
}
