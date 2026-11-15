package momentumview

import (
	"strings"
	"testing"
	"time"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func TestAwayStripUsesSharedDateNotSelectedBucket(t *testing.T) {
	state := types.ContentState{MomentumDate: "2026-10-14", Settings: &api.CoreSettings{AwayDates: []string{"2026-10-14"}}}
	normal := StripRows(testTheme(), types.ContentState{}, 100)
	for _, bucket := range []int{0, 3} {
		state.MomentumBucket = bucket
		rows := StripRows(testTheme(), state, 100)
		if len(rows) != len(normal) || rows[0] != normal[0] || !strings.Contains(ansi.Strip(rows[1]), "Away day") {
			t.Fatalf("bucket %d: incorrect strip %q", bucket, rows)
		}
	}
	state.MomentumDate = "2026-10-15"
	if got := strings.Join(StripRows(testTheme(), state, 100), "\n"); strings.Contains(got, "Away day") {
		t.Fatalf("strip followed bucket instead of shared date: %q", got)
	}
}

func TestAwayDayCountsUseExplicitDatesAndInclusiveBoundaries(t *testing.T) {
	dates := awayDateSet(&api.CoreSettings{
		AwayDates:         []string{"2026-10-01", "2026-10-07", "2026-10-07", "2026-10-31", "invalid"},
		RestSpecificDates: []string{"2026-10-03"},
	})
	for _, tc := range []struct {
		start, end string
		want       int
	}{
		{"2026-10-01", "2026-10-07", 2}, {"2026-10-01", "2026-10-31", 3}, {"2026-10-02", "2026-10-06", 0},
	} {
		point := sharedtypes.MomentumSeriesPoint{StartDate: tc.start, EndDate: tc.end, ProtectedDayCount: 7, Skipped: true}
		if got := awayDayCount(point, dates); got != tc.want {
			t.Fatalf("%s..%s: got %d away days, want %d", tc.start, tc.end, got, tc.want)
		}
	}
}

func TestAwayCalendarRetainsSelectionAndHits(t *testing.T) {
	series := dailyMomentumSeries(time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), []int{1, 0, 1, 1, 1, 1, 1}, 1)
	normal := calendarLines(testTheme(), series, 70, 15, true, "2026-10-05", nil)
	marked := calendarLines(testTheme(), series, 70, 15, true, "2026-10-05", map[string]bool{"2026-10-05": true})
	out := ansi.Strip(strings.Join(marked, "\n"))
	if !strings.Contains(out, "A") || !strings.Contains(out, "•") || !strings.Contains(out, "█") {
		t.Fatalf("away calendar lost marker, selection, or status: %s", out)
	}
	if len(marked) != len(normal) || lineContaining(out, "hit") != lineContaining(ansi.Strip(strings.Join(normal, "\n")), "hit") {
		t.Fatal("away metadata changed calendar size or hit totals")
	}
}

func TestAwayBarsKeepDataWhenColumnsAreCompressed(t *testing.T) {
	series := dailyMomentumSeries(time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), make([]int, 90), 4)
	for i := range series {
		series[i].Count = 3
	}
	normal := renderMomentumBars(testTheme(), barDef(), series, 30, 10, true, -1, nil)
	marked := renderMomentumBars(testTheme(), barDef(), series, 30, 10, true, -1, map[string]bool{"2026-10-02": true})
	if !strings.Contains(ansi.Strip(strings.Join(marked, "\n")), "A") {
		t.Fatal("compressed away column has no marker")
	}
	if ansi.Strip(strings.Join(normal[:8], "\n")) != ansi.Strip(strings.Join(marked[:8], "\n")) {
		t.Fatal("away annotation changed chart data glyphs")
	}
}

func TestAwayAnnotationsReachMomentumPaneAndDialog(t *testing.T) {
	card := momentumTestCard("Focus", "October")
	card.Series[0].StartDate, card.Series[0].EndDate = "2026-10-01", "2026-10-31"
	state := types.ContentState{Width: 200, Height: 45, MomentumCards: []sharedtypes.MomentumCard{card}, Settings: &api.CoreSettings{AwayDates: []string{"2026-10-01", "2026-10-31"}}}
	for _, width := range []int{90, 200} {
		state.Width = width
		out := ansi.Strip(Render(testTheme(), state))
		if !strings.Contains(out, "2 away days") || !strings.Contains(out, "Focus") {
			t.Fatalf("width %d: missing pane annotation: %s", width, out)
		}
	}
	_, lines, ok := Detail(testTheme(), state, 100)
	if !ok || !strings.Contains(ansi.Strip(strings.Join(lines, "\n")), "2 away days") {
		t.Fatalf("dialog lost away annotation: %q", lines)
	}
}

func TestAwayViewKeepsMomentumPanePositions(t *testing.T) {
	card := momentumTestCard("Focus", "October")
	for _, width := range []int{60, 120, 200} {
		for _, height := range []int{20, 30, 45} {
			state := types.ContentState{Width: width, Height: height, MomentumDate: "2026-10-14", MomentumCards: []sharedtypes.MomentumCard{card}}
			normal := strings.Split(ansi.Strip(Render(testTheme(), state)), "\n")
			state.Settings = &api.CoreSettings{AwayDates: []string{state.MomentumDate}}
			away := strings.Split(ansi.Strip(Render(testTheme(), state)), "\n")
			if len(normal) != len(away) {
				t.Fatalf("%dx%d: away state changed view height", width, height)
			}
			for i := 2; i < len(normal); i++ {
				if normal[i] != away[i] {
					t.Fatalf("%dx%d: pane moved or changed at row %d", width, height, i)
				}
			}
		}
	}
}
