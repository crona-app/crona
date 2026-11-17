package types

import (
	"math"
	"testing"
)

func TestSleepHoursBounds(t *testing.T) {
	for _, hours := range []float64{0, 7.5, 24} {
		if err := ValidateSleepHours(hours); err != nil {
			t.Errorf("%v: %v", hours, err)
		}
	}
	for _, hours := range []float64{-1, 24.0001, 68, math.NaN(), math.Inf(1)} {
		if err := ValidateSleepHours(hours); err == nil {
			t.Errorf("accepted %v", hours)
		}
	}
}

func TestMomentumDurationLimits(t *testing.T) {
	for _, tc := range []struct {
		period HabitStreakPeriod
		hours  int
	}{{HabitStreakPeriodDay, 24}, {HabitStreakPeriodWeek, 168}, {HabitStreakPeriodMonth, 744}} {
		limit := MomentumDurationLimit(tc.period)
		if limit != tc.hours*3600 {
			t.Fatalf("%s: limit %d", tc.period, limit)
		}
		if err := ValidateDurationSeconds("Required work", limit, limit, true); err != nil {
			t.Fatal(err)
		}
		if err := ValidateDurationSeconds("Required work", limit+1, limit, true); err == nil {
			t.Fatalf("%s: accepted over-limit value", tc.period)
		}
	}
}
