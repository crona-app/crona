package summary

import (
	"strings"
	"testing"
	"time"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func dayText(state types.ContentState, width int) string {
	return ansi.Strip(strings.Join(DayLines(types.Theme{}, state, width), "\n"))
}

func squash(s string) string { return strings.Join(strings.Fields(s), " ") }

func TestDayLinesIncludeStatsMetrics(t *testing.T) {
	state := summaryTestState()
	state.SummarySnapshot.Focus = &api.FocusScoreSummary{
		Score:               82,
		Level:               sharedtypes.FocusScoreLevelStrong,
		Reason:              sharedtypes.FocusScoreReasonBalanced,
		TargetWorkedSeconds: 3600,
	}
	state.SummarySnapshot.Metrics = &api.DailyMetricsDay{
		WorkedSeconds:       2400,
		RestSeconds:         600,
		SessionCount:        3,
		TotalIssues:         4,
		CompletedIssues:     2,
		AbandonedIssues:     1,
		HabitDueCount:       3,
		HabitCompletedCount: 2,
		HabitFailedCount:    1,
	}
	rendered := dayText(state, 42)
	for _, want := range []string{"82", "Score", "strong", "Worked", "of 1h", "to go 20m", "3 sessions, avg 13m", "rest 10m", "1/1", "all done", "2/3", "Issues", "Habits", "1 failed"} {
		if !strings.Contains(squash(rendered), want) {
			t.Fatalf("expected summary stats to contain %q, got %q", want, rendered)
		}
	}
	// Each figure once: no verdict banner, the target only under the arc, and no hint for a
	// balanced day.
	for _, gone := range []string{"Strong", "to target", "Worked/", "balanced", "Accountability", "▲"} {
		if strings.Contains(squash(rendered), gone) {
			t.Fatalf("expected summary stats without %q, got %q", gone, rendered)
		}
	}
}

func TestDayLinesHintLowFocusOnce(t *testing.T) {
	state := summaryTestState()
	state.SummarySnapshot.Focus = &api.FocusScoreSummary{
		Score:               23,
		Level:               sharedtypes.FocusScoreLevelLow,
		Reason:              sharedtypes.FocusScoreReasonNeedsBreaks,
		TargetWorkedSeconds: 19800,
	}
	state.SummarySnapshot.Metrics = &api.DailyMetricsDay{WorkedSeconds: 1140, SessionCount: 1}
	rendered := squash(dayText(state, 60))
	for _, want := range []string{"Score", "low", "of 5h30m", "to go 5h11m", "take a break"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected %q in the low-focus summary, got %q", want, rendered)
		}
	}
	if n := strings.Count(rendered, "23"); n != 1 {
		t.Fatalf("expected the score once, got %d times in %q", n, rendered)
	}
	if strings.Contains(rendered, "▲") || strings.Contains(rendered, "Low focus") {
		t.Fatalf("expected no low-focus banner, got %q", rendered)
	}
}

func TestDayLinesOneStatPerRowWithinWidth(t *testing.T) {
	state := summaryTestState()
	state.SummarySnapshot.Focus = &api.FocusScoreSummary{
		Score:               52,
		Level:               sharedtypes.FocusScoreLevelOverextended,
		Reason:              "overextended",
		TargetWorkedSeconds: 21600,
	}
	state.SummarySnapshot.Metrics = &api.DailyMetricsDay{
		WorkedSeconds:       6660,
		SessionCount:        2,
		TotalIssues:         4,
		CompletedIssues:     1,
		HabitDueCount:       5,
		HabitCompletedCount: 4,
	}
	for _, line := range DayLines(types.Theme{}, state, 42) {
		if w := ansi.StringWidth(line); w > 42 || strings.Contains(line, "…") {
			t.Fatalf("expected every line whole within 42 cells, got %d: %q", w, ansi.Strip(line))
		}
	}
	if n := strings.Count(strings.ToLower(dayText(state, 42)), "overextended"); n != 1 {
		t.Fatalf("expected the focus level once, got %d", n)
	}
}

func TestDayLinesCoverPlanWellbeingAndStreaks(t *testing.T) {
	rendered := squash(dayText(summaryTestState(), 42))
	for _, want := range []string{"Plan", "none high risk", "nothing late", "no misses yet", "Wellbeing", "Wellbeing", "4/5", "3/5", "Momentum", "Focus ↑3d", "Check-ins ↑4d"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected the Day summary to contain %q, got %q", want, rendered)
		}
	}
}

func TestDayLinesWaitForTheShownDate(t *testing.T) {
	state := summaryTestState()
	state.SummaryDate = "2026-07-29"
	if got := ansi.Strip(strings.Join(DayLines(types.Theme{}, state, 100), "\n")); got != "Loading summary…" {
		t.Fatalf("expected a loading line for another day's snapshot, got %q", got)
	}
}

func summaryTestState() types.ContentState {
	estimate := 30
	target := 20
	return types.ContentState{
		View:        "summary",
		Pane:        "summary",
		Width:       120,
		Height:      42,
		SummaryDate: "2026-07-28",
		Timer:       &api.TimerState{State: "running"},
		Elapsed:     900,
		SummarySnapshot: &api.SummarySnapshot{
			Date: "2026-07-28",
			Issues: &api.DailyIssueSummary{
				Date:                  "2026-07-28",
				TotalIssues:           1,
				CompletedIssues:       1,
				TotalEstimatedMinutes: 30,
				WorkedSeconds:         1200,
				Issues: []api.Issue{{
					Title:           "Ship summary",
					Status:          sharedtypes.IssueStatusDone,
					EstimateMinutes: &estimate,
					WorkedSeconds:   1200,
				}},
			},
			Habits: []api.HabitDailyItem{{
				HabitWithMeta: api.HabitWithMeta{
					Habit: api.Habit{Name: "Walk", TargetMinutes: &target},
				},
				Completed: true,
				Status:    sharedtypes.HabitCompletionStatusCompleted,
			}},
			Plan: &api.DailyPlan{
				Date: "2026-07-28",
				Summary: api.DailyPlanAccountabilitySummary{
					PlannedCount:        1,
					CompletedCount:      1,
					AccountabilityScore: 100,
				},
			},
			CheckIn: &api.DailyCheckIn{Date: "2026-07-28", Mood: 4, Energy: 3},
			Streaks: &api.StreakSummary{
				CurrentFocusDays:   3,
				CurrentCheckInDays: 4,
				CurrentHabitDays:   2,
			},
		},
	}
}

func TestDayLinesWrapLongNotes(t *testing.T) {
	state := summaryTestState()
	note := "Productive, but sleep debt is starting to show in the afternoon slump"
	state.SummarySnapshot.CheckIn.Notes = &note
	lines := DayLines(types.Theme{}, state, 42)
	for _, line := range lines {
		if ansi.StringWidth(line) > 42 {
			t.Fatalf("expected notes within 42 cells, got %q", ansi.Strip(line))
		}
	}
	if !strings.Contains(squash(dayText(state, 42)), note) {
		t.Fatalf("expected the whole note, got %q", dayText(state, 42))
	}
}

func TestDayLinesListTopPenalties(t *testing.T) {
	state := summaryTestState()
	entry := func(id int64, score float64, highRisk, historical bool) sharedtypes.DailyPlanEntry {
		return sharedtypes.DailyPlanEntry{
			IssueID:   id,
			FailScore: score,
			Penalty: &sharedtypes.DailyPlanEntryPenalty{
				HighRisk: highRisk, Historical: historical,
				Contribution: &sharedtypes.DailyPlanPenaltyShare{AccountabilityPct: score / 10},
			},
		}
	}
	state.SummarySnapshot.Plan = &sharedtypes.DailyPlan{
		Date: state.SummarySnapshot.Date,
		Entries: []sharedtypes.DailyPlanEntry{
			entry(1, 1, false, false), entry(2, 3.38, true, false), entry(3, 5, true, true),
		},
	}
	state.SummarySnapshot.Issues = &sharedtypes.DailyIssueSummary{Issues: []sharedtypes.Issue{
		{ID: 1, Title: "Small slip"}, {ID: 2, Title: "Fix sync conflict"},
	}}
	rendered := squash(dayText(state, 60))
	if !strings.Contains(rendered, "Top ▲ Fix sync conflict · due today · 34% Small slip · due today · 10%") ||
		strings.Contains(rendered, "5.00") {
		t.Fatalf("expected top penalties largest first without resolved entries, got %q", rendered)
	}
}

func TestDayLineFitsTheDayInOneLine(t *testing.T) {
	state := summaryTestState()
	state.SummarySnapshot.Focus = &api.FocusScoreSummary{
		Score: 82, Level: sharedtypes.FocusScoreLevelStrong, TargetWorkedSeconds: 3600,
	}
	state.SummarySnapshot.Metrics = &api.DailyMetricsDay{WorkedSeconds: 2400, RestSeconds: 600, SessionCount: 3}
	wide := ansi.Strip(DayLine(types.Theme{}, state, 120))
	for _, want := range []string{"Focus 82 strong", "Worked 40m/1h", "Rest 10m", "Sessions 3"} {
		if !strings.Contains(wide, want) {
			t.Fatalf("expected the day line to contain %q, got %q", want, wide)
		}
	}
	narrow := ansi.Strip(DayLine(types.Theme{}, state, 40))
	if ansi.StringWidth(narrow) > 40 || !strings.Contains(narrow, "Focus") || strings.Contains(narrow, "Sessions") {
		t.Fatalf("expected parts to drop from the right at 40, got %q", narrow)
	}
	state.SummarySnapshot.Focus = nil
	if line := DayLine(types.Theme{}, state, 120); line != "" {
		t.Fatalf("expected no day line before the summary loads, got %q", line)
	}
}

func TestDayLinesCarryTheStripsBarsAndBestStreaks(t *testing.T) {
	state := summaryTestState()
	state.SummarySnapshot.Focus = &api.FocusScoreSummary{Score: 60, TargetWorkedSeconds: 3600}
	state.SummarySnapshot.Metrics = &api.DailyMetricsDay{
		TotalIssues: 2, CompletedIssues: 1, HabitDueCount: 2, HabitCompletedCount: 1,
	}
	state.DailyPlannedIssues = []api.Issue{{Status: "done"}, {Status: "planned"}}
	lines := strings.Split(dayText(state, 42), "\n")
	text := strings.Join(lines, "\n")
	for _, want := range []string{"1/2", "Issues    ● 1 planned", "1 left"} {
		if !strings.Contains(text, want) || strings.Contains(text, "█") {
			t.Fatalf("expected arcs with %q and no bars, got %q", want, lines)
		}
	}
	if rendered := squash(strings.Join(lines, "\n")); !strings.Contains(rendered, "best") {
		t.Fatalf("expected best streaks under Momentum, got %q", rendered)
	}
}

func TestWellbeingShowsArcGaugesAndTheWeekTrend(t *testing.T) {
	state := summaryTestState()
	sleep, screen := 7.5, 300
	state.SummarySnapshot.CheckIn.SleepHours = &sleep
	state.SummarySnapshot.CheckIn.ScreenTimeMinutes = &screen
	avgMood, avgSleep, avgScreen := 3.0, 7.4, 120.0
	state.SummarySnapshot.Rollup = &sharedtypes.MetricsRollup{
		AverageMood: &avgMood, AverageSleepHours: &avgSleep, AverageScreenTimeMins: &avgScreen,
	}
	state.Settings = &sharedtypes.CoreSettings{SleepTargetMinutes: 480, ScreenTimeLimitMins: 360}
	rendered := squash(dayText(state, 78))
	// 5h of a 6h limit leaves 1h; the week averaged 2h, so 4h left, and today is worse.
	for _, want := range []string{"4/5", "3/5", "7h30m", "1h", "Mood Energy Sleep Screen left", "▲ 7d 3.0", "= 7d 7h24m", "▼ 7d 4h", "⣿"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected %q in the wellbeing gauges, got %q", want, rendered)
		}
	}
	screen = 400
	if rendered := squash(dayText(state, 78)); !strings.Contains(rendered, "+40m") {
		t.Fatalf("expected screen time over the limit as +40m, got %q", rendered)
	}
	for _, line := range DayLines(types.Theme{}, state, 30) {
		if w := ansi.StringWidth(line); w > 30 {
			t.Fatalf("expected gauges to wrap within 30 cells, got %d: %q", w, ansi.Strip(line))
		}
	}
}

func TestMomentumListsTheDaysMomentumsUnmetFirst(t *testing.T) {
	state := summaryTestState()
	state.SummaryDate = "2026-07-28"
	state.DueHabits = []api.HabitDailyItem{{HabitWithMeta: api.HabitWithMeta{Habit: api.Habit{ID: 7}}}}
	state.AllIssues = []api.IssueWithMeta{{Issue: api.Issue{ID: 3, StreamID: 9}, RepoID: 4}}
	state.DailyPlannedIssues = []api.Issue{{ID: 3, StreamID: 9}}
	point := func(count, target int, met bool) []sharedtypes.MomentumSeriesPoint {
		return []sharedtypes.MomentumSeriesPoint{{StartDate: "2026-07-27", EndDate: "2026-08-02", Count: count, Target: target, MetTarget: met}}
	}
	state.MomentumCards = []api.MomentumCard{
		{Definition: api.HabitStreakDefinition{ID: "fit", Name: "Fitness", Enabled: true, Period: "week", RequiredCount: 3, HabitIDs: []int64{7}},
			Current: 4, Longest: 9, Series: point(3, 3, true)},
		{Definition: api.HabitStreakDefinition{ID: "core", Name: "Core work", Enabled: true, TargetKind: sharedtypes.MomentumTargetKindContext,
			Period: "week", RequiredCount: 1, Contexts: []sharedtypes.MomentumContext{{RepoID: 4}}},
			Series: point(3600, 7200, false)},
		{Definition: api.HabitStreakDefinition{ID: "other", Name: "Unrelated", Enabled: true, Period: "week", HabitIDs: []int64{99}}},
	}
	text := dayText(state, 80)
	core, fit := strings.Index(text, "Core work"), strings.Index(text, "Fitness")
	if core < 0 || fit < 0 || core > fit {
		t.Fatalf("expected the unmet Core work before the met Fitness, got %q", text)
	}
	for _, want := range []string{"3/3 ✓ ↑4w best 9w", "1h/2h", "▰▰▰▱▱▱"} {
		if !strings.Contains(squash(text), want) {
			t.Fatalf("expected %q in the momentum rows, got %q", want, text)
		}
	}
	if strings.Contains(text, "Unrelated") {
		t.Fatalf("expected momentums the day does not feed left out, got %q", text)
	}
}

func TestPlanVerdictAndResolvedChips(t *testing.T) {
	theme := types.Theme{}
	s := sharedtypes.DailyPlanAccountabilitySummary{
		PlannedCount: 9, HighRiskIssueCount: 2, DelayedIssueCount: 4, AvgDelayDays: 2.3, MaxDelayDays: 5,
		FailedCount: 1, CarriedCount: 2,
	}
	got := ansi.Strip(planVerdict(theme, s))
	if got != "at risk" {
		t.Fatalf("verdict = %q", got)
	}
	if got := ansi.Strip(resolvedChips(theme, s)); got != "✗ 1 failed  ↷ 2 carried" {
		t.Fatalf("chips = %q", got)
	}
	if got := ansi.Strip(delayText(theme, s)); got != "4 · avg 2.3d · max 5d" {
		t.Fatalf("delay = %q", got)
	}
}

func planSelectState() types.ContentState {
	state := summaryTestState()
	risky := &sharedtypes.DailyPlanEntryPenalty{HighRisk: true}
	state.SummarySnapshot.Plan = &sharedtypes.DailyPlan{
		Date: state.SummarySnapshot.Date,
		Entries: []sharedtypes.DailyPlanEntry{
			// On the day's plan but adding nothing to the penalty: not listed.
			{IssueID: 5, Status: sharedtypes.DailyPlanEntryStatusPlanned},
			{IssueID: 1, Status: sharedtypes.DailyPlanEntryStatusPlanned, FailScore: 0.2},
		},
		Contributing: []sharedtypes.DailyPlanEntry{
			{IssueID: 1, Status: sharedtypes.DailyPlanEntryStatusPlanned, FailScore: 0.2,
				Penalty: &sharedtypes.DailyPlanEntryPenalty{}},
			// Planned on an earlier day and still open: listed though the day holds no entry for it.
			{IssueID: 2, Date: "2026-07-20", Status: sharedtypes.DailyPlanEntryStatusPlanned, FailScore: 0.1,
				CurrentDelayedDays: 2, Penalty: &sharedtypes.DailyPlanEntryPenalty{}},
			{IssueID: 3, Date: "2026-07-21", Status: sharedtypes.DailyPlanEntryStatusPlanned, FailScore: 0.9,
				Penalty: risky},
			{IssueID: 4, Status: sharedtypes.DailyPlanEntryStatusPlanned, FailScore: 0.5, Penalty: risky},
			{IssueID: 6, Status: sharedtypes.DailyPlanEntryStatusPlanned, FailScore: 0.05,
				Penalty: &sharedtypes.DailyPlanEntryPenalty{}},
		},
	}
	state.AllIssues = []api.IssueWithMeta{
		{Issue: api.Issue{ID: 1, Title: "Small task", Status: "ready"}},
		{Issue: api.Issue{ID: 2, Title: "Older task", Status: "in_progress"}},
		{Issue: api.Issue{ID: 3, Title: "Risky task", Status: "ready"}},
		// Done since the plan loaded: dropped.
		{Issue: api.Issue{ID: 4, Title: "Done task", Status: "done"}},
		{Issue: api.Issue{ID: 5, Title: "Calm task", Status: "ready"}},
		{Issue: api.Issue{ID: 6, Title: "Tiny task", Status: "ready"}},
	}
	return state
}

func TestPlanItemsAreTheIssuesBehindThePenalty(t *testing.T) {
	state := planSelectState()
	var titles []string
	for _, item := range PlanItems(state.SummarySnapshot, state.AllIssues) {
		titles = append(titles, item.Title)
	}
	if strings.Join(titles, ",") != "Risky task,Small task,Older task,Tiny task" {
		t.Fatalf("expected the contributing issues riskiest first, got %v", titles)
	}
	// An older daemon sends no contributing list: the day's penalised entries stand in.
	state.SummarySnapshot.Plan.Contributing = nil
	state.SummarySnapshot.Plan.Entries[1].Penalty = &sharedtypes.DailyPlanEntryPenalty{}
	if items := PlanItems(state.SummarySnapshot, state.AllIssues); len(items) != 1 || items[0].Title != "Small task" {
		t.Fatalf("expected the day's penalised entry without a contributing list, got %+v", items)
	}
}

func TestPlanListSelectsEveryContributingIssue(t *testing.T) {
	state := planSelectState()
	rest := squash(dayText(state, 80))
	if !strings.Contains(rest, "Top ▲ Risky task") || !strings.Contains(rest, "Older task · late 2d") ||
		strings.Contains(rest, "Tiny task") || strings.Contains(rest, "[P]") {
		t.Fatalf("expected the top three behind the penalty and no key at rest, got %q", rest)
	}
	state.DailyDetailFocus, state.DailyDetailSummary = true, true
	if focused := squash(dayText(state, 80)); !strings.Contains(focused, "Plan [P] select") {
		t.Fatalf("expected [P] select on the focused summary, got %q", focused)
	}
	state.DailyPlanSelect, state.DailyPlanCursor = true, 1
	selecting := squash(dayText(state, 80))
	for _, want := range []string{"Open ▲ Risky task", "▶ Small task", "[s] status [d] due", "Older task", "Tiny task"} {
		if !strings.Contains(selecting, want) {
			t.Fatalf("expected %q in the selected plan list, got %q", want, selecting)
		}
	}
	for _, gone := range []string{"Done task", "Calm task", "[P]"} {
		if strings.Contains(selecting, gone) {
			t.Fatalf("expected no %q while selecting, got %q", gone, selecting)
		}
	}
}

func TestIssueLegendCountsTheWholePlan(t *testing.T) {
	state := planSelectState()
	counts := plannedStatusCounts(state)
	if counts["ready"] != 2 || len(counts) != 1 {
		t.Fatalf("expected the day's plan issues counted, got %v", counts)
	}
}

func TestLastCheckInLineSaysHowLongAgo(t *testing.T) {
	now := time.Date(2026, 11, 8, 15, 0, 0, 0, time.Local)
	cases := map[string]string{"2026-11-08": "(today)", "2026-11-07": "(yesterday)", "2026-11-04": "(4 days ago)"}
	for date, want := range cases {
		got := ansi.Strip(strings.Join(lastCheckInLine(types.Theme{}, date, nil, now, 60), "\n"))
		if !strings.Contains(got, want) || !strings.Contains(got, date) {
			t.Errorf("%s: %q lacks %q", date, got, want)
		}
	}
	if lines := lastCheckInLine(types.Theme{}, "", nil, now, 60); lines != nil {
		t.Errorf("no check-in should add nothing, got %q", lines)
	}
}
