package daily

import (
	"fmt"
	"strings"
	"testing"

	"crona.work/tui/internal/api"
	uistate "crona.work/tui/internal/tui/state"
	"github.com/charmbracelet/x/ansi"
)

func TestWidePreviewFollowsSelectionWithoutOpeningDetails(t *testing.T) {
	state := detailState()
	state.DailyDetailFocus = false
	state.SummaryDate = state.DashboardDate
	state.SummarySnapshot = &api.SummarySnapshot{Date: state.SummaryDate, Metrics: &api.DailyMetricsDay{}, Focus: &api.FocusScoreSummary{Score: 72}}
	right := ansi.Strip(dailyRightColumn(testTheme(), state, 76, 40))
	if !strings.Contains(right, "Fix sync conflict") || !strings.Contains(right, "Make the sync") || !strings.Contains(right, "Plan") {
		t.Fatalf("expected preview description and summary together:\n%s", right)
	}
	calendar, preview, summary := rowOf(t, right, "Calendar"), rowOf(t, right, "Fix sync conflict"), rowOf(t, right, "Wed 27 May")
	if calendar >= preview || preview >= summary {
		t.Fatalf("wrong column order: calendar %d, preview %d, summary %d", calendar, preview, summary)
	}
	state.Pane = "habits"
	state.Cursors["habits"] = 1
	state.DueHabits[1].Notes = new("Read a chapter before bed")
	right = ansi.Strip(dailyRightColumn(testTheme(), state, 76, 40))
	if !strings.Contains(right, "Read a chapter") || strings.Contains(right, "Make the sync") {
		t.Fatalf("preview did not follow habit focus:\n%s", right)
	}
}

func TestWideDetailKeepsShortSummaryAndRestoresSavedCalendar(t *testing.T) {
	state := detailState()
	state.SummaryDate = state.DashboardDate
	state.SummarySnapshot = &api.SummarySnapshot{Date: state.SummaryDate, Focus: &api.FocusScoreSummary{Score: 72}, Metrics: &api.DailyMetricsDay{WorkedSeconds: 7200}}
	for _, saved := range [][]string{nil, {uistate.CollapseDailyCalendarWeek}, {uistate.CollapseDailyCalendar}} {
		state.Settings = &api.CoreSettings{CollapsedPanes: saved}
		state.DailyDetailFocus = true
		expanded := ansi.Strip(dailyRightColumn(testTheme(), state, 76, 40))
		if !strings.Contains(expanded, "─ Week") || !strings.Contains(expanded, "Score   72") || !strings.Contains(expanded, "[esc] back") {
			t.Fatalf("expanded details missing compact summary:\n%s", expanded)
		}
		hidden := uistate.PaneCollapsed(saved, uistate.CollapseDailyCalendar)
		if strings.Contains(expanded, "─ Week ▸") != hidden || (!hidden && !strings.Contains(expanded, "May 2026")) {
			t.Fatalf("focused calendar should show its week unless explicitly hidden:\n%s", expanded)
		}
		state.DailyDetailSummary = true
		summary := ansi.Strip(dailyRightColumn(testTheme(), state, 76, 40))
		if !strings.Contains(summary, "Plan") || strings.Contains(summary, "Make the sync") {
			t.Fatalf("summary focus should reduce item to its header:\n%s", summary)
		}
		state.DailyDetailFocus, state.DailyDetailSummary = false, false
		restored := ansi.Strip(dailyRightColumn(testTheme(), state, 76, 40))
		if strings.Contains(restored, "Calendar ▸") != uistate.PaneCollapsed(saved, uistate.CollapseDailyCalendar) {
			t.Fatalf("saved calendar mode not restored:\n%s", restored)
		}
	}
}

func TestPreviewColumnKeepsDimensionsAtEveryHeight(t *testing.T) {
	for _, width := range []int{160, 200, 240} {
		for _, height := range []int{3, 8, 12, 16, 24, 40, 60} {
			for _, mode := range []string{"compact", "details", "summary"} {
				t.Run(fmt.Sprintf("%dx%d/%s", width, height, mode), func(t *testing.T) {
					state := detailState()
					state.Width, state.Height = width, height
					state.DailyDetailFocus = mode != "compact"
					state.DailyDetailSummary = mode == "summary"
					columnW := layoutFor(state).detailW
					lines := strings.Split(dailyRightColumn(testTheme(), state, columnW, height), "\n")
					if len(lines) != height {
						t.Fatalf("column height = %d, want %d", len(lines), height)
					}
					for _, line := range lines {
						if got := ansi.StringWidth(line); got != columnW {
							t.Fatalf("column width = %d, want %d: %q", got, columnW, ansi.Strip(line))
						}
					}
				})
			}
		}
	}
}

func TestPreviewExcerptAndEmptySelection(t *testing.T) {
	state := dailyState(200, 40)
	state.DailyIssues = nil
	_, lines, selected := previewBody(testTheme(), state, 40, 7)
	if selected || !strings.Contains(ansi.Strip(strings.Join(lines, "\n")), "No task selected") {
		t.Fatalf("expected empty preview, got %v", lines)
	}
	lines = previewExcerpt(testTheme(), new(strings.Repeat("long description ", 30)), 30)
	if len(lines) != 2 || !strings.HasSuffix(ansi.Strip(lines[1]), "…") {
		t.Fatalf("expected two-line excerpt ending in ellipsis: %v", lines)
	}
}

func TestPreviewTerminalCaptures(t *testing.T) {
	state := detailState()
	state.SummaryDate = state.DashboardDate
	state.SummarySnapshot = &api.SummarySnapshot{Date: state.SummaryDate, Focus: &api.FocusScoreSummary{Score: 72}, Metrics: &api.DailyMetricsDay{WorkedSeconds: 7200}}
	for _, mode := range []string{"compact", "details", "summary"} {
		state.DailyDetailFocus, state.DailyDetailSummary = mode != "compact", mode == "summary"
		t.Logf("%s:\n%s", mode, ansi.Strip(dailyRightColumn(testTheme(), state, 76, 40)))
	}
}

func TestDailyPreviewModesFitFullView(t *testing.T) {
	for _, width := range []int{80, 120, 159, 160, 200} {
		for _, height := range []int{20, 32, 40} {
			for _, mode := range []string{"compact", "details", "summary"} {
				state := detailState()
				state.Width, state.Height = width, height
				state.DailyDetailFocus, state.DailyDetailSummary = mode != "compact", mode == "summary"
				lines := strings.Split(renderView(testTheme(), state), "\n")
				if len(lines) != height {
					t.Fatalf("%dx%d/%s: height = %d", width, height, mode, len(lines))
				}
				for _, line := range lines {
					if got := ansi.StringWidth(line); got != width {
						t.Fatalf("%dx%d/%s: width = %d", width, height, mode, got)
					}
				}
			}
		}
	}
}
