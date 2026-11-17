package types

import (
	"fmt"
	"math"
)

// MaxDailyDurationSeconds is the ceiling for an explicitly entered daily duration.
const MaxDailyDurationSeconds = 24 * 60 * 60

// MaxDailyDurationMinutes is the daily ceiling in persisted minute fields.
const MaxDailyDurationMinutes = MaxDailyDurationSeconds / 60

// ValidateDurationSeconds validates a duration without performing an unsafe conversion.
func ValidateDurationSeconds(label string, seconds, maximum int, positive bool) error {
	if seconds < 0 || seconds > maximum {
		return fmt.Errorf("%s must be between 0 and %dh", label, maximum/3600)
	}
	if positive && seconds == 0 {
		return fmt.Errorf("%s must be positive", label)
	}
	return nil
}

// ValidateDailyMinutes validates an optional daily value supplied in minutes.
func ValidateDailyMinutes(label string, minutes int) error {
	if minutes < 0 || minutes > MaxDailyDurationMinutes {
		return fmt.Errorf("%s must be between 0 and 24h", label)
	}
	return nil
}

// ValidateSleepHours rejects invalid and physically impossible daily sleep values.
func ValidateSleepHours(hours float64) error {
	if math.IsNaN(hours) || math.IsInf(hours, 0) || hours < 0 || hours > 24 {
		return fmt.Errorf("sleep must be between 0 and 24h")
	}
	return nil
}

// MomentumDurationLimit is the physical ceiling of a recurring work target.
func MomentumDurationLimit(period HabitStreakPeriod) int {
	switch NormalizeHabitStreakPeriod(period) {
	case HabitStreakPeriodWeek:
		return 7 * MaxDailyDurationSeconds
	case HabitStreakPeriodMonth:
		return 31 * MaxDailyDurationSeconds
	default:
		return MaxDailyDurationSeconds
	}
}
