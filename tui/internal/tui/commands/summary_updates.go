package commands

import (
	"slices"
	"time"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
)

func loadSummaryUpdates(c *api.Client, date string) ([]api.Op, map[string]string, error) {
	health, err := c.GetHealth()
	if err != nil {
		return nil, nil, err
	}
	location := time.Local
	if health.Timezone != "" {
		location, err = time.LoadLocation(health.Timezone)
		if err != nil {
			return nil, nil, err
		}
	}
	start, err := time.ParseInLocation(time.DateOnly, date, location)
	if err != nil {
		return nil, nil, err
	}
	ops, err := c.ListOpsSince(start.Add(-time.Second).UTC().Format(time.RFC3339))
	if err != nil {
		return nil, nil, err
	}
	updates, titles := summaryUpdates(ops, start, start.AddDate(0, 0, 1))
	return updates, titles, nil
}

func summaryUpdates(ops []api.Op, start, end time.Time) ([]api.Op, map[string]string) {
	updates := []api.Op{}
	titles := map[string]string{}
	for _, op := range ops {
		if op.Entity != sharedtypes.OpEntityIssue {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, op.Timestamp)
		if err != nil || at.Before(start) || !at.Before(end) {
			continue
		}
		// Preserve the display timezone along with the operation in the snapshot.
		op.Timestamp = at.In(start.Location()).Format(time.RFC3339Nano)
		updates = append(updates, op)
	}
	slices.SortStableFunc(updates, func(a, b api.Op) int {
		aTime, _ := time.Parse(time.RFC3339Nano, a.Timestamp)
		bTime, _ := time.Parse(time.RFC3339Nano, b.Timestamp)
		return bTime.Compare(aTime)
	})
	for _, op := range updates {
		payload, _ := op.Payload.(map[string]any)
		if title, ok := payload["title"].(string); ok && titles[op.EntityID] == "" {
			titles[op.EntityID] = title
		}
	}
	return updates[:min(7, len(updates))], titles
}
