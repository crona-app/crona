package checkins

import (
	"strings"
	"testing"

	"crona.work/tui/internal/api"
	types "crona.work/tui/internal/tui/views/types"
	"github.com/charmbracelet/x/ansi"
)

func history() []api.DailyCheckIn {
	return []api.DailyCheckIn{
		{Date: "2026-11-15", Mood: 4, Energy: 3, CreatedAt: "2026-11-15T09:15:00+05:30", SleepHours: new(7.5), SleepScore: new(85), ScreenTimeMinutes: new(120), Notes: new("Quiet morning")},
		{Date: "2026-11-14", Mood: 2, Energy: 4, CreatedAt: "2026-11-14T10:30:00Z", Notes: new("Busy afternoon")},
	}
}

func TestCheckInHistoryResponsiveLayoutAndDetails(t *testing.T) {
	for _, width := range []int{40, 60, 99, 100, 160, 200} {
		state := types.ContentState{Width: width, Height: 30, CheckInHistory: history()}
		text := ansi.Strip(Render(types.Theme{}, state))
		if width == 60 || width == 160 {
			t.Logf("%d-column check-ins:\n%s", width, text)
		}
		lines := strings.Split(text, "\n")
		if len(lines) != 30 {
			t.Fatalf("width %d: got %d lines", width, len(lines))
		}
		for _, line := range lines {
			if ansi.StringWidth(line) != width {
				t.Errorf("width %d: row has %d cells: %q", width, ansi.StringWidth(line), line)
			}
		}
		if !strings.Contains(text, "Check-ins") || !strings.Contains(text, "SUN 15 NOV") || !strings.Contains(text, "[e] edit") {
			t.Errorf("history missing headings/actions at width %d:\n%s", width, text)
		}
		if width >= 100 && !strings.Contains(text, "Wellbeing") {
			t.Errorf("detail pane missing at width %d", width)
		}
	}
	state := types.ContentState{}
	entry := history()[0]
	text := ansi.Strip(strings.Join(DetailLines(types.Theme{}, state, &entry, 80), "\n"))
	for _, want := range []string{"Created", "2026-11-15", "Mood", "Energy", "Sleep", "Quiet morning"} {
		if !strings.Contains(text, want) {
			t.Errorf("detail missing %q:\n%s", want, text)
		}
	}
}

func TestCheckInHistoryFilteringAndStates(t *testing.T) {
	for _, filter := range []string{"2026-11-14", "busy", "MOOD 2"} {
		indices := FilteredIndices(history(), filter, nil)
		if len(indices) != 1 || indices[0] != 1 {
			t.Errorf("filter %q = %v, want [1]", filter, indices)
		}
	}
	for _, test := range []struct {
		loading           bool
		err, filter, want string
	}{
		{false, "", "", "No check-ins recorded"},
		{true, "", "", "Loading check-ins"},
		{false, "failed", "", "Unable to load"},
		{false, "", "missing", "No check-ins match"},
	} {
		state := types.ContentState{Width: 80, Height: 20, CheckInHistoryLoading: test.loading, CheckInHistoryError: test.err, Filters: map[string]string{"checkins": test.filter}}
		if text := ansi.Strip(Render(types.Theme{}, state)); !strings.Contains(text, test.want) {
			t.Errorf("state missing %q:\n%s", test.want, text)
		}
	}
	state := types.ContentState{Width: 80, Height: 20, CheckInHistory: history(), CheckInHistoryReadOnly: true}
	if text := ansi.Strip(Render(types.Theme{}, state)); strings.Contains(text, "[e] edit") || strings.Contains(text, "[D] delete") {
		t.Error("read-only history advertises mutations")
	}
}
