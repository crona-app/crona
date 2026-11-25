package daily

import (
	"strings"
	"testing"

	"crona.work/tui/internal/api"
	"github.com/charmbracelet/x/ansi"
)

func TestHistoricalAwayKeepsDailyLayout(t *testing.T) {
	for _, width := range []int{60, 120, 200} {
		for _, focus := range []int{0, 1, 2} {
			state := dailyState(width, 40)
			state.DailyDetailFocus = focus > 0
			state.DailyDetailSummary = focus == 2
			normal := renderView(testTheme(), state)
			monthFits := CalendarMonthFits(testTheme(), state)
			state.HistoricalAwayDate = state.DashboardDate
			away := renderView(testTheme(), state)
			plain := ansi.Strip(away)
			if strings.Count(plain, "You chose to rest and recover.") != 1 {
				t.Fatalf("width %d focus %d: expected one recovery message:\n%s", width, focus, plain)
			}
			for _, hidden := range []string{"Ship momentum block", "Walk", "Read", "Loading summary", "[enter]", "[f] focus"} {
				if strings.Contains(plain, hidden) {
					t.Fatalf("width %d focus %d: away content contains %q:\n%s", width, focus, hidden, plain)
				}
			}
			headings := []string{"─ Issues"}
			if width >= 160 || focus == 0 {
				headings = append(headings, "─ Habits Due")
			}
			if width >= 160 || focus == 2 {
				headings = append(headings, "─ Wed 27 May")
			}
			for _, heading := range headings {
				if rowOf(t, away, heading) != rowOf(t, normal, heading) {
					t.Fatalf("width %d focus %d: %s moved", width, focus, heading)
				}
			}
			if CalendarMonthFits(testTheme(), state) != monthFits {
				t.Fatalf("width %d: away date changed calendar sizing", width)
			}
			normalLines, awayLines := strings.Split(normal, "\n"), strings.Split(away, "\n")
			if len(normalLines) != len(awayLines) {
				t.Fatalf("width %d focus %d: height changed from %d to %d", width, focus, len(normalLines), len(awayLines))
			}
			if width < 160 {
				end := rowOf(t, normal, "─ Issues")
				if strings.Join(normalLines[:end], "\n") != strings.Join(awayLines[:end], "\n") {
					t.Fatalf("width %d: calendar changed on away date", width)
				}
			} else {
				l := layoutFor(state)
				_, calendarH := calendarPane(testTheme(), state, l.detailW, detailCalendarFits(state.Height))
				if state.DailyDetailFocus {
					_, calendarH = focusedCalendar(testTheme(), state, l.detailW, state.Height)
				}
				for i := range calendarH {
					if ansi.Strip(ansi.Cut(normalLines[i], l.listW+3, width)) != ansi.Strip(ansi.Cut(awayLines[i], l.listW+3, width)) {
						t.Fatalf("focus %d: calendar row %d changed: normal %q away %q", focus, i, ansi.Strip(normalLines[i]), ansi.Strip(awayLines[i]))
					}
				}
			}
		}
	}
}

func TestHistoricalAwayDateLabelFitsWhileLoading(t *testing.T) {
	state := dailyState(120, 40)
	state.HistoricalAwayDate = state.DashboardDate
	for _, room := range []int{24, 30, 40, 80} {
		got := SubTabMeta(testTheme(), state, room)
		plain := ansi.Strip(got)
		if !strings.Contains(plain, "2026-05-27 · Away day") || ansi.StringWidth(got) > room {
			t.Fatalf("room %d: expected date and away label within room, got %q", room, plain)
		}
	}
	state.HistoricalAwayDate = "2026-05-26"
	if got := ansi.Strip(SubTabMeta(testTheme(), state, 80)); strings.Contains(got, "Away day") {
		t.Fatalf("stale date status leaked into navigation: %q", got)
	}
}

func TestHistoricalAwayHidesLoadedSummary(t *testing.T) {
	for _, width := range []int{60, 200} {
		state := dailyState(width, 40)
		state.HistoricalAwayDate = state.DashboardDate
		state.SummaryDate = state.DashboardDate
		state.DailyDetailFocus, state.DailyDetailSummary = true, true
		state.SummarySnapshot = &api.SummarySnapshot{
			Date:    state.SummaryDate,
			Metrics: &api.DailyMetricsDay{WorkedSeconds: 7200},
			Focus:   &api.FocusScoreSummary{Score: 72},
		}
		out := ansi.Strip(renderView(testTheme(), state))
		for _, hidden := range []string{"Score", "Worked", "2h", "[S]", "[esc]"} {
			if strings.Contains(out, hidden) {
				t.Fatalf("width %d: away summary contains %q:\n%s", width, hidden, out)
			}
		}
		if !strings.Contains(out, "You chose to rest and recover.") {
			t.Fatalf("width %d: missing recovery message", width)
		}
	}
}
