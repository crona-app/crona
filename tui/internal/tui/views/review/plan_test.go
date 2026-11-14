package review

import (
	"strings"
	"testing"

	"crona.work/tui/internal/api"
	types "crona.work/tui/internal/tui/views/types"
)

func TestPlanExplainsOutcomesAndEstimates(t *testing.T) {
	state := reviewState(200, 50)
	state.DashboardWindow.MissedCount = 2
	state.DashboardWindow.AbandonedCount = 1
	state.DashboardWindow.CarryOverCount = 4
	state.DashboardWindow.Days[0].CompletedCount = 5
	state.ReviewPrev = &types.ReviewPrev{Window: &api.DashboardWindowSummary{PlannedCount: 10, CompletedCount: 6}}
	state.GoalProgress = &api.GoalProgressSummary{
		GroupBy: "issue", TotalEstimateMinutes: 180, TotalActualSeconds: 7200,
		EstimatedItems: 3, AverageDeltaMinutes: -20, AverageDeltaPercent: -33, EstimateBias: "over",
	}
	text := plain(strings.Join(planLines(types.Theme{}, state, 64), "\n"))
	for _, want := range []string{
		"81% completed", "22/27 done", "3 failed · 2 missed their planned day", "1 abandoned",
		"4 carried from an earlier plan", "1/7 planned days completed in full",
		"Completion +21 pp vs previous window", "Counts are plan entries", "carryovers can",
		"Estimated 3h · tracked 2h", "Tracked time 1h below total estimate",
		"estimates ran 33% high across 3 compared issues", "Average gap 20m per compared item",
		"Tracked time is time spent",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("plan missing %q:\n%s", want, text)
		}
	}
	// The expanded pane must still be visible in the full dashboard at a normal size.
	if out := plain(Render(types.Theme{}, state)); !strings.Contains(out, "Average gap 20m") {
		t.Errorf("expanded estimate details clipped at 200x50:\n%s", out)
	}
	t.Log("Expanded Plan pane:\n" + text)
}

func TestPlanEstimateEmptyAndOverrunStates(t *testing.T) {
	for _, test := range []struct {
		name     string
		progress *api.GoalProgressSummary
		want     string
	}{
		{"unavailable", nil, "Estimate comparison unavailable."},
		{"unestimated", &api.GoalProgressSummary{}, "No estimates recorded"},
		{"untracked", &api.GoalProgressSummary{TotalEstimateMinutes: 60}, "accuracy is not available"},
		{"overrun", &api.GoalProgressSummary{TotalEstimateMinutes: 60, TotalActualSeconds: 7200, EstimatedItems: 1}, "Tracked time 1h above total estimate"},
		{"matched", &api.GoalProgressSummary{TotalEstimateMinutes: 60, TotalActualSeconds: 3600, EstimatedItems: 1}, "Tracked time matches total estimate"},
	} {
		t.Run(test.name, func(t *testing.T) {
			text := strings.Join(strings.Fields(plain(strings.Join(planEstimateLines(types.Theme{}, test.progress, 44), "\n"))), " ")
			if !strings.Contains(text, test.want) {
				t.Errorf("estimate lines = %q, want %q", text, test.want)
			}
		})
	}
}

func TestEstimateDriftNamesComparedGroups(t *testing.T) {
	text := estimateDrift(&api.GoalProgressSummary{GroupBy: "repo", EstimatedItems: 2, EstimateBias: "balanced"})
	if text != "estimates held across 2 compared repos" {
		t.Errorf("repo drift = %q, want repo sample label", text)
	}
}
