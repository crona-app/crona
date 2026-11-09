package itemdetail

import (
	"testing"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	types "crona.work/tui/internal/tui/views/types"
)

func plan(date string, entry api.DailyPlanEntry) *api.DailyPlan {
	return &api.DailyPlan{Date: date, Entries: []api.DailyPlanEntry{entry}}
}

// pushPlan has issue 2 late with a live projection of one more day, or settled.
func pushPlan(historical bool) *api.DailyPlan {
	penalty := &api.DailyPlanEntryPenalty{HighRisk: true, HighRiskThreshold: 2.5, Historical: historical}
	if !historical {
		penalty.IfPostponedOneDay = &api.DailyPlanPenaltyProjection{FailScore: 3.81, AccountabilityDelta: 0.43}
	}
	return plan("2026-10-24", api.DailyPlanEntry{IssueID: 2, CurrentDelayedDays: 5, FailScore: 3.38, Penalty: penalty})
}

func TestRescheduleCarriesOnlyWorkedIssuesToday(t *testing.T) {
	settings := &api.CoreSettings{CarryOverEnabled: true}
	plan := plan("2026-10-24", api.DailyPlanEntry{IssueID: 2, Status: sharedtypes.DailyPlanEntryStatusPlanned, WorkedSeconds: 60})
	if !RescheduleCarries(plan, 2, "2026-10-24", settings) {
		t.Fatal("expected a worked issue today to carry")
	}
	if RescheduleCarries(plan, 2, "2026-10-25", settings) {
		t.Fatal("expected no carry when the plan is not today's")
	}
	plan.Entries[0].WorkedSeconds = 0
	if RescheduleCarries(plan, 2, "2026-10-24", settings) {
		t.Fatal("expected an untouched issue to reschedule normally")
	}
	plan.Entries[0].WorkedSeconds = 60
	if RescheduleCarries(plan, 2, "2026-10-24", &api.CoreSettings{}) {
		t.Fatal("expected no carry when carry-over is off")
	}
}

func TestCarriedStatusShowsOnlyOnTheDayCarriedFrom(t *testing.T) {
	state := types.ContentState{}
	state.DailyPlan = plan("2026-10-24", api.DailyPlanEntry{IssueID: 2, Status: sharedtypes.DailyPlanEntryStatusCarried})
	if got := Status(state, 2, sharedtypes.IssueStatusInReview); got != "carried" {
		t.Fatalf("expected carried on the day carried from, got %q", got)
	}
	state.DailyPlan = plan("2026-10-24", api.DailyPlanEntry{IssueID: 2, Status: sharedtypes.DailyPlanEntryStatusPlanned, CarriedFromDate: new("2026-10-23")})
	if got := Status(state, 2, sharedtypes.IssueStatusInReview); got != "in_review" {
		t.Fatalf("expected the issue's own status on the day carried into, got %q", got)
	}
}

func TestRescheduleHintShowsPushCost(t *testing.T) {
	if got := RescheduleHint(pushPlan(false), 2); got != "Push a day: penalty rises to 3.81 (+0.43)" {
		t.Fatalf("unexpected hint %q", got)
	}
	if got := RescheduleHint(pushPlan(true), 2); got != "" {
		t.Fatalf("expected no hint for a resolved entry, got %q", got)
	}
	if got := RescheduleHint(nil, 2); got != "" {
		t.Fatalf("expected no hint without a plan, got %q", got)
	}
}
