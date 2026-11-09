// Package itemdetail renders what an issue or a habit has to say about itself: the detail
// bodies Today, Issues and Habits show, the markers on their rows, and the plan's view of an
// issue (penalty, carry-over) that rescheduling consults.
package itemdetail

import (
	"fmt"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	types "crona.work/tui/internal/tui/views/types"
)

func planEntry(plan *api.DailyPlan, issueID int64) *api.DailyPlanEntry {
	if plan == nil {
		return nil
	}
	for i := range plan.Entries {
		if plan.Entries[i].IssueID == issueID {
			return &plan.Entries[i]
		}
	}
	return nil
}

// shownEntry is the shown day's plan entry for an issue, if it has one.
func shownEntry(state types.ContentState, issueID int64) *api.DailyPlanEntry {
	return planEntry(state.DailyPlan, issueID)
}

// RescheduleCarries reports whether rescheduling the issue to a later day will carry it over:
// the kernel carries an issue that was worked on its planned day when that day is today.
func RescheduleCarries(plan *api.DailyPlan, issueID int64, today string, settings *api.CoreSettings) bool {
	if plan == nil || plan.Date != today || settings == nil || !settings.CarryOverEnabled {
		return false
	}
	entry := planEntry(plan, issueID)
	return entry != nil && entry.Status == sharedtypes.DailyPlanEntryStatusPlanned && entry.WorkedSeconds > 0
}

// RescheduleHint is the cost of pushing an issue one more day, shown when rescheduling it,
// or "" when the issue has no live penalty projection in the plan.
func RescheduleHint(plan *api.DailyPlan, issueID int64) string {
	entry := planEntry(plan, issueID)
	if entry == nil || entry.Penalty == nil || entry.Penalty.IfPostponedOneDay == nil {
		return ""
	}
	push := entry.Penalty.IfPostponedOneDay
	hint := "Push a day: penalty " + pushedText(push)
	if push.BecomesHighRisk {
		hint += " ▲ high risk"
	}
	return hint
}

// Status is the status to show for the issue on the shown day: "carried" on a day the issue
// was carried away from, its own status everywhere else.
func Status(state types.ContentState, issueID int64, status sharedtypes.IssueStatus) string {
	if entry := shownEntry(state, issueID); entry != nil && entry.Status == sharedtypes.DailyPlanEntryStatusCarried {
		return string(sharedtypes.IssueStatusCarried)
	}
	return string(status)
}

func plural(count int, unit string) string {
	if count == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", count, unit)
}
