package commands

import (
	"fmt"
	"testing"
	"time"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
)

func TestSummaryUpdatesFiltersDayAndLimitsNewestActions(t *testing.T) {
	t.Parallel()
	location := time.FixedZone("test", 19800)
	start := time.Date(2026, 11, 14, 0, 0, 0, 0, location)
	end := start.AddDate(0, 0, 1)
	ops := []api.Op{
		{ID: "previous", Entity: sharedtypes.OpEntityIssue, Timestamp: start.Add(-time.Second).Format(time.RFC3339)},
		{ID: "next", Entity: sharedtypes.OpEntityIssue, Timestamp: end.Format(time.RFC3339)},
		{ID: "invalid", Entity: sharedtypes.OpEntityIssue, Timestamp: "invalid"},
		{ID: "noise", Entity: sharedtypes.OpEntityActiveContext, Timestamp: start.Add(time.Hour).Format(time.RFC3339)},
		{ID: "midnight", Entity: sharedtypes.OpEntityIssue, EntityID: "12", Timestamp: start.UTC().Format(time.RFC3339), Payload: map[string]any{"title": "Deleted task"}},
	}
	for i := range 10 {
		ops = append(ops, api.Op{ID: fmt.Sprint(i), Entity: sharedtypes.OpEntityIssue, EntityID: "12", Action: sharedtypes.OpActionDelete, Timestamp: start.Add(time.Duration(i+1) * time.Minute).UTC().Format(time.RFC3339)})
	}
	updates, titles := summaryUpdates(ops, start, end)
	if len(updates) != 7 || updates[0].ID != "9" || updates[6].ID != "3" {
		t.Fatalf("summaryUpdates = %v, want seven actions descending from 9 to 3", updates)
	}
	if updates[0].Timestamp != "2026-11-14T00:10:00+05:30" {
		t.Errorf("newest timestamp = %q, want display timezone preserved", updates[0].Timestamp)
	}
	if titles["12"] != "Deleted task" {
		t.Errorf("titles = %v, want deleted title retained outside seven-action window", titles)
	}
	midnight, _ := summaryUpdates(ops[:5], start, end)
	if len(midnight) != 1 || midnight[0].ID != "midnight" {
		t.Errorf("day boundary updates = %v, want midnight included and next midnight excluded", midnight)
	}
}
