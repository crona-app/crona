package review

import (
	"strings"
	"testing"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func reviewState(width, height int) types.ContentState {
	end := "2026-10-14"
	dates := []string{
		"2026-10-08",
		"2026-10-09",
		"2026-10-10",
		"2026-10-11",
		"2026-10-12",
		"2026-10-13",
		"2026-10-14",
	}
	metrics := make([]api.DailyMetricsDay, 0, len(dates))
	days := make([]sharedtypes.DashboardWindowDay, 0, len(dates))
	for i, date := range dates {
		day := api.DailyMetricsDay{
			Date:                  date,
			WorkedSeconds:         (i + 2) * 3600,
			TotalEstimatedMinutes: 300,
		}
		if i != 2 {
			day.CheckIn = &api.DailyCheckIn{
				Date:      date,
				Mood:      1 + i%5,
				Energy:    3,
				CreatedAt: date + "T09:00:00Z",
			}
		}
		metrics = append(metrics, day)
		days = append(
			days,
			sharedtypes.DashboardWindowDay{
				Date:           date,
				Status:         "done",
				PlannedCount:   5,
				CompletedCount: 4,
			},
		)
	}
	mood, energy, sleep := 3.8, 3.4, 7.1
	return types.ContentState{
		View: "review", Pane: "review_days", Width: width, Height: height,
		DashboardDate: end, SummaryDate: end,
		ReviewWindowDays: 7, ReviewWindowEnd: end,
		MetricsRange:       metrics,
		RollupMetricsRange: metrics,
		MetricsRollup: &api.MetricsRollup{
			Days: 7, CheckInDays: 6, WorkedSeconds: 17 * 3600,
			AverageMood: &mood, AverageEnergy: &energy, AverageSleepHours: &sleep,
			LatestBurnout: &api.BurnoutIndicator{
				Score:   20,
				Level:   "low",
				Factors: map[string]float64{"breakDebt": 0.3, "recoveryBreaks": -0.2},
			},
		},
		DashboardWindow: &api.DashboardWindowSummary{
			StartDate:      dates[0],
			EndDate:        end,
			PlannedCount:   27,
			CompletedCount: 22,
			FailedCount:    3,
			Days:           days,
		},
		WeeklyFocusScore: &api.FocusScoreSummary{Score: 72},
	}
}

func plain(s string) string { return ansi.Strip(s) }

func TestDashboardSections(t *testing.T) {
	state := reviewState(200, 50)
	state.WorkBreakdown = &api.TimeDistributionSummary{TotalSeconds: 7200, Rows: []api.TimeDistributionRow{
		{Key: "repo:1", Label: "OSS", WorkedSeconds: 3600, Percent: 50},
		{Key: "repo:2", Label: "Work", WorkedSeconds: 3600, Percent: 50},
	}}
	out := plain(Render(types.Theme{}, state))
	for _, want := range []string{
		"Thu 8 Oct – Wed 14 Oct · 7 days",
		"Worked 17h", "Focus 72", "Plan 22/27", "Mood 3.8", "Energy 3.4", "Sleep 7h06m",
		"─ Wellbeing ", "Mood · avg 2.5", "Energy · avg 3.0", "┄", "─ Work ", "● hours worked", "● focus",
		"Thu 8", "▸Wed 14", "▸ Wed 14 Oct · mood 2 · energy 3", "plan 4/5",
		"─ Risk ", "20 LOW", "easing ◀│▶ raising", "+30 Break debt",
		"[b] group", "OSS", "─ Plan ", "22/27 done",
		"[←/→] day", "[enter] open in Today",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q:\n%s", want, out)
		}
	}
	state.ReviewPrev = &types.ReviewPrev{Rollup: &api.MetricsRollup{Days: 7}, Focus: &api.FocusScoreSummary{}}
	if out := plain(Render(types.Theme{}, state)); strings.Contains(out, "Worked 17h ▼") || strings.Contains(out, "Focus 72 ▲") {
		t.Fatalf("change shown without previous data:\n%s", out)
	}
	// The charts draw in braille.
	if !strings.ContainsFunc(out, func(r rune) bool { return r > 0x2800 && r <= 0x28FF }) {
		t.Fatalf("expected braille lines:\n%s", out)
	}
}

func TestDashboardComparesWithPreviousWindow(t *testing.T) {
	state := reviewState(200, 50)
	mood := 3.0
	state.ReviewPrev = &types.ReviewPrev{
		Rollup: &api.MetricsRollup{Days: 7, CheckInDays: 6, WorkedSeconds: 20 * 3600, SessionCount: 9, AverageMood: &mood},
		Focus:  &api.FocusScoreSummary{Score: 60, SessionCount: 9},
	}
	out := plain(Render(types.Theme{}, state))
	for _, want := range []string{"Focus 72 ▲ 12", "Worked 17h ▼ 3h", "Mood 3.8 ▲ 0.8"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q:\n%s", want, out)
		}
	}
}

func TestDashboardStacksSideWhenNarrow(t *testing.T) {
	out := plain(Render(types.Theme{}, reviewState(90, 60)))
	wellbeing, risk := strings.Index(out, "─ Wellbeing "), strings.Index(out, "─ Risk ")
	if wellbeing < 0 || risk < wellbeing {
		t.Fatalf("expected the risk under the charts:\n%s", out)
	}
}

func TestLongWindowKeepsDailyPoints(t *testing.T) {
	state := reviewState(200, 50)
	state.ReviewWindowDays = 30
	out := plain(Render(types.Theme{}, state))
	if !strings.Contains(out, "30 days") || !strings.Contains(out, "▸Oct 14") {
		t.Fatalf("expected a daily axis with month labels:\n%s", out)
	}
}

func TestHistoricalAwayDate(t *testing.T) {
	state := reviewState(200, 50)
	state.HistoricalAwayDate = "2026-10-14"
	if out := plain(Render(types.Theme{}, state)); !strings.Contains(out, "rest and recover") {
		t.Fatalf("expected the away message:\n%s", out)
	}
}

func TestRenderExactSize(t *testing.T) {
	for _, size := range [][2]int{{60, 20}, {120, 36}, {200, 50}, {200, 22}} {
		lines := strings.Split(Render(types.Theme{}, reviewState(size[0], size[1])), "\n")
		if len(lines) != size[1] {
			t.Fatalf("%v: expected %d lines, got %d", size, size[1], len(lines))
		}
		for i, line := range lines {
			if w := ansi.StringWidth(line); w != size[0] {
				t.Fatalf("%v: line %d is %d cells: %q", size, i, w, plain(line))
			}
		}
	}
}

func TestEstimateDriftInWords(t *testing.T) {
	cases := map[string]string{"over": "ran 39% high", "under": "ran 39% low", "balanced": "held"}
	for bias, want := range cases {
		got := estimateDrift(&api.GoalProgressSummary{EstimatedItems: 3, AverageDeltaPercent: -39, EstimateBias: bias})
		if !strings.Contains(got, want) || !strings.Contains(got, "3 issues") {
			t.Fatalf("%s: %q lacks %q", bias, got, want)
		}
	}
}
