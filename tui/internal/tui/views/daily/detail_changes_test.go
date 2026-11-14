package daily

import (
	"strings"
	"testing"

	"crona.work/tui/internal/api"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	"github.com/charmbracelet/x/ansi"
)

func TestFocusedDetailsKeepSourceSelectionMarker(t *testing.T) {
	state := detailState()
	state.DailyDetailFocus = true
	for _, pane := range []string{"issues", "habits"} {
		state.Pane = pane
		for _, cursor := range []int{0, 1} {
			state.Cursors[pane] = cursor
			var text, title string
			if pane == "issues" {
				text = ansi.Strip(renderIssues(testTheme(), state, 100, 25))
				title = state.DailyIssues[cursor].Title
			} else {
				text = ansi.Strip(renderHabits(testTheme(), state, 100, 25))
				title = state.DueHabits[cursor].Name
			}
			found := false
			for _, line := range strings.Split(text, "\n") {
				if strings.Contains(line, viewchrome.SelectionCursor) && strings.Contains(line, title) {
					found = true
				}
			}
			if !found {
				t.Fatalf("%s cursor %d: source marker missing for %q:\n%s", pane, cursor, title, text)
			}
		}
	}
}

func TestDetailActionGroupsMatchAvailableKeys(t *testing.T) {
	state := detailState()
	state.DailyDetailFocus = true
	text := ansi.Strip(renderDetail(testTheme(), state, 76, 40))
	for _, want := range []string{"[s] status", "[d] due", "[P] pin", "[+] tags", "[e] edit", "[D] delete", "[f] focus", "[m] log"} {
		if !strings.Contains(text, want) {
			t.Errorf("task details missing %q:\n%s", want, text)
		}
	}
	state.Pane = "habits"
	text = ansi.Strip(renderDetail(testTheme(), state, 76, 40))
	for _, want := range []string{"[x] done", "[F] fail", "[e] edit", "[+] tags", "[m] log"} {
		if !strings.Contains(text, want) {
			t.Errorf("habit details missing %q:\n%s", want, text)
		}
	}
	state.Timer = &api.TimerState{State: "running"}
	text = ansi.Strip(renderDetail(testTheme(), state, 76, 40))
	if strings.Contains(text, "[m] log") || strings.Contains(text, "[D] delete") {
		t.Fatalf("busy timer advertised unavailable actions:\n%s", text)
	}
}

func TestDetailsKeepContentFirstAndActionsByTheirData(t *testing.T) {
	state := detailState()
	state.DailyDetailFocus = true
	state.DetailSessionsIssueID = state.DailyIssues[1].ID
	text := ansi.Strip(renderDetail(testTheme(), state, 90, 30))
	lines := strings.Split(text, "\n")
	if !strings.Contains(lines[2], "Make the sync") {
		t.Fatalf("detail content was displaced by hints:\n%s", text)
	}
	for _, want := range [][2]string{{"in progress", "[s] status"}, {"Due 2026-05-27", "[d] due"}, {"Sessions", "[m] log"}} {
		found := false
		for _, line := range lines {
			found = found || (strings.Contains(line, want[0]) && strings.Contains(line, want[1]))
		}
		if !found {
			t.Errorf("expected %q alongside %q:\n%s", want[1], want[0], text)
		}
	}
	last := lines[len(lines)-1]
	if !strings.Contains(last, "[e] edit") || !strings.Contains(last, "[D] delete") {
		t.Fatalf("general actions are not pinned to the bottom: %q", last)
	}
	state.DailyDetailScroll = 3
	text = ansi.Strip(renderDetail(testTheme(), state, 90, 30))
	lines = strings.Split(text, "\n")
	if !strings.Contains(lines[len(lines)-1], "[D] delete") {
		t.Fatal("scrolling moved the general action footer")
	}
}
