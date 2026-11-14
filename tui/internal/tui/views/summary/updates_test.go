package summary

import (
	"fmt"
	"strings"
	"testing"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

func TestRecentUpdatesAppearAfterPlanBeforeWork(t *testing.T) {
	t.Parallel()
	state := summaryTestState()
	state.SummarySnapshot.UpdateTitles = map[string]string{"3": "Removed task"}
	state.SummarySnapshot.RecentUpdates = []api.Op{
		{EntityID: "3", Action: sharedtypes.OpActionDelete, Timestamp: "2026-11-14T12:30:00+05:30"},
		{EntityID: "4", Action: sharedtypes.OpActionUpdate, Timestamp: "2026-11-14T12:29:00+05:30", Payload: map[string]any{"from": "ready", "to": "active"}},
		{EntityID: "5", Action: sharedtypes.OpActionCreate, Timestamp: "2026-11-14T12:28:00+05:30", Payload: map[string]any{"title": "New task"}},
	}
	rendered := dayText(state, 80)
	for _, want := range []string{"12:30  Deleted · #3 Removed task", "12:29  ready → active · #4", "12:28  Added · #5"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("summary missing %q:\n%s", want, rendered)
		}
	}
	plan, updates, wellbeing := strings.Index(rendered, "Plan"), strings.Index(rendered, "Recent updates"), strings.Index(rendered, "Wellbeing")
	if plan < 0 || updates <= plan || wellbeing <= updates {
		t.Errorf("section order = plan %d, updates %d, wellbeing %d", plan, updates, wellbeing)
	}
}

func TestRecentUpdatesLimitWidthAndEmptyStates(t *testing.T) {
	t.Parallel()
	state := summaryTestState()
	for i := range 10 {
		state.SummarySnapshot.RecentUpdates = append(state.SummarySnapshot.RecentUpdates, api.Op{EntityID: fmt.Sprint(i), Action: sharedtypes.OpActionCreate})
	}
	for _, width := range []int{20, 40, 80} {
		lines := renderRecentUpdates(types.Theme{}, state, width)
		if len(lines) != 8 {
			t.Fatalf("recent updates lines = %d, want title plus seven actions", len(lines))
		}
		for _, line := range lines {
			if ansi.StringWidth(line) > width {
				t.Errorf("line %q exceeds width %d", line, width)
			}
		}
	}
	state.SummarySnapshot.RecentUpdates = nil
	if text := dayText(state, 80); !strings.Contains(text, "No issue updates this day.") {
		t.Errorf("empty summary = %q, want explicit empty state", text)
	}
	state.SummarySnapshot.RecentUpdatesUnavailable = true
	if text := dayText(state, 80); !strings.Contains(text, "Updates unavailable.") || strings.Contains(text, "No issue updates") {
		t.Errorf("unavailable summary = %q, want distinct failure state", text)
	}
}

func TestIssueUpdateActions(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		op   api.Op
		want string
	}{
		{"status", api.Op{Action: sharedtypes.OpActionUpdate, Payload: map[string]any{"from": "ready", "to": "active"}}, "ready → active"},
		{"scheduled", api.Op{Action: sharedtypes.OpActionUpdate, Payload: map[string]any{"todoForDate": "2026-11-15", "status": "planned"}}, "Scheduled 2026-11-15"},
		{"cleared", api.Op{Action: sharedtypes.OpActionUpdate, Payload: map[string]any{"todoForDate": nil}}, "Unscheduled"},
		{"edit", api.Op{Action: sharedtypes.OpActionUpdate, Payload: map[string]any{"Title": map[string]any{"Set": true, "Value": "Renamed"}}}, "Updated"},
		{"restored", api.Op{Action: sharedtypes.OpActionRestore}, "Restored"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := issueUpdateAction(test.op); got != test.want {
				t.Errorf("issueUpdateAction(%+v) = %q, want %q", test.op, got, test.want)
			}
		})
	}
}
