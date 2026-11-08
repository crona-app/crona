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

func TestTimelineSections(t *testing.T) {
	state := reviewState(200, 50)
	state.WorkBreakdown = &api.TimeDistributionSummary{TotalSeconds: 7200, Rows: []api.TimeDistributionRow{
		{Key: "repo:1", Label: "OSS", WorkedSeconds: 3600, Percent: 50},
		{Key: "repo:2", Label: "Work", WorkedSeconds: 3600, Percent: 50},
	}}
	state.WorkColumns = make([]*api.TimeDistributionSummary, 7)
	state.WorkColumns[6] = &api.TimeDistributionSummary{Rows: []api.TimeDistributionRow{{Key: "repo:1", WorkedSeconds: 3600}}}
	out := plain(Render(types.Theme{}, state))
	for _, want := range []string{
		"Thu 8 Oct – Wed 14 Oct · 7 days", "▸Wed 14", "Thu 8", "Window",
		"─ Wellbeing ", "6/7 check-ins", "Mood", "Energy", "Sleep", "Focus", "avg 3.8", "avg 72",
		"─ Work by repo ", "[b] group", "OSS", "████", "50% ██░░", "Total", "17h",
		"─ Plan ", "Done", "22/27", "Estimates",
		"─ Risk ", "20 LOW", "easing ◀│▶ raising", "+30 Break debt", "Break recovery",
		"[←/→] day", "[enter] open in Today",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "vs prev") {
		t.Fatalf("no comparison without a previous window:\n%s", out)
	}
}

func TestTimelineComparesWithPreviousWindow(t *testing.T) {
	state := reviewState(200, 50)
	mood := 3.0
	state.ReviewPrev = &types.ReviewPrev{
		Rollup: &api.MetricsRollup{Days: 7, CheckInDays: 6, WorkedSeconds: 20 * 3600, AverageMood: &mood},
		Focus:  &api.FocusScoreSummary{Score: 60},
	}
	out := plain(Render(types.Theme{}, state))
	for _, want := range []string{"vs prev", "▲ 12", "▼ 3h", "▲ 0.8"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q:\n%s", want, out)
		}
	}
}

func TestShortScreenDropsRiskChartFirst(t *testing.T) {
	out := plain(Render(types.Theme{}, reviewState(200, 20)))
	if !strings.Contains(out, "─ Risk ") || strings.Contains(out, "easing ◀") || !strings.Contains(out, "Total") {
		t.Fatalf("expected the risk heading without its chart, and Work kept:\n%s", out)
	}
}

func TestLongWindowGoesByWeek(t *testing.T) {
	state := reviewState(200, 50)
	state.ReviewWindowDays = 30
	out := plain(Render(types.Theme{}, state))
	if !strings.Contains(out, "by week") || !strings.Contains(out, "▸W42") || strings.Contains(out, "Thu 8") {
		t.Fatalf("expected week columns:\n%s", out)
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
