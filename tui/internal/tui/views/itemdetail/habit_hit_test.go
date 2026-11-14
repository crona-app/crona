package itemdetail

import (
	"testing"

	types "crona.work/tui/internal/tui/views/types"
)

func TestHabitHitArcsUseIndependentWindows(t *testing.T) {
	days := map[string]habitDay{
		"2026-11-15": {status: "completed"},
		"2026-11-09": {status: "failed"},
		"2026-11-08": {status: "completed"},
		"2026-10-17": {status: "completed"},
		"2026-10-16": {status: "failed"},
		"2026-11-16": {status: "failed"},
	}
	for _, test := range []struct {
		window   int
		label    string
		fraction float64
	}{
		{7, "Hit/7d", 0.5},
		{30, "Hit/30d", 0.75},
	} {
		arc, ok := habitHitArc(types.Theme{}, days, "2026-11-15", test.window)
		if !ok || arc.Label != test.label || arc.Fraction != test.fraction {
			t.Errorf("%d-day arc = %+v, %v; want %s fraction %v", test.window, arc, ok, test.label, test.fraction)
		}
	}
	if _, ok := habitHitArc(types.Theme{}, nil, "2026-11-15", 7); ok {
		t.Error("empty history produced a hit-rate arc")
	}
}
