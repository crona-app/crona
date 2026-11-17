package controller

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	shareddto "crona.work/shared/dto"
	sharedtypes "crona.work/shared/types"

	"github.com/charmbracelet/bubbles/textinput"
)

func ParseEstimateInput(raw string) (*int, error) {
	return ParseOptionalDurationMinutes(raw, "Estimate")
}

func ParseDueDateInput(raw string) (*string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", raw); err != nil {
		return nil, fmt.Errorf("due date must be YYYY-MM-DD")
	}
	return &raw, nil
}

func ValueToPointer(raw string) *string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return &raw
}

func ValueOrEmpty(raw *string) string {
	if raw == nil {
		return ""
	}
	return *raw
}

func ParseNumericID(raw string) int64 {
	value, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return value
}

func EndSessionRequest(inputs []textinput.Model) shareddto.EndSessionRequest {
	if len(inputs) == 0 {
		return shareddto.EndSessionRequest{}
	}
	req := shareddto.EndSessionRequest{
		CommitMessage: ValueToPointer(inputs[0].Value()),
	}
	if len(inputs) > 1 {
		req.WorkedOn = ValueToPointer(inputs[1].Value())
	}
	if len(inputs) > 2 {
		req.Outcome = ValueToPointer(inputs[2].Value())
	}
	if len(inputs) > 3 {
		req.NextStep = ValueToPointer(inputs[3].Value())
	}
	if len(inputs) > 4 {
		req.Blockers = ValueToPointer(inputs[4].Value())
	}
	if len(inputs) > 5 {
		req.Links = ValueToPointer(inputs[5].Value())
	}
	return req
}

func ParseDurationInput(raw string, required bool, label string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if required {
			return 0, fmt.Errorf("%s is required", label)
		}
		return 0, nil
	}
	minutes, err := parseFlexibleDurationMinutes(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be like 90, 90m, 1h30m, or 1.5h", label)
	}
	if err := validateRawDurationLimit(raw, sharedtypes.MaxDailyDurationSeconds/60, label); err != nil {
		return 0, err
	}
	seconds := minutes * 60
	if seconds < 0 {
		return 0, fmt.Errorf("%s must be non-negative", label)
	}
	if required && seconds == 0 {
		return 0, fmt.Errorf("%s must be positive", label)
	}
	return seconds, nil
}

func ParseDurationSecondsInput(raw string, required bool, label string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if required {
			return 0, fmt.Errorf("%s is required", label)
		}
		return 0, nil
	}
	var seconds float64
	if minutes, err := strconv.ParseFloat(raw, 64); err == nil {
		seconds = minutes * 60
	} else {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return 0, fmt.Errorf("%s must be like 90m, 1h30m, or 1h34m23s", label)
		}
		seconds = d.Seconds()
	}
	if math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds < 0 || seconds > float64(sharedtypes.MomentumDurationLimit(sharedtypes.HabitStreakPeriodMonth)) {
		return 0, fmt.Errorf("%s must be between 0 and 744h", label)
	}
	value := int(math.Round(seconds))
	if required && value == 0 {
		return 0, fmt.Errorf("%s must be positive", label)
	}
	return value, nil
}

func ParseClockInput(raw string) (*string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if _, err := time.Parse("15:04", raw); err != nil {
		return nil, fmt.Errorf("time must be HH:MM")
	}
	return &raw, nil
}

func ParseOptionalDurationMinutes(raw string, label string) (*int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if label != "Estimate" {
		if err := validateRawDurationLimit(raw, sharedtypes.MaxDailyDurationMinutes, label); err != nil {
			return nil, err
		}
	}
	value, err := parseFlexibleDurationMinutes(raw)
	if err != nil {
		return nil, fmt.Errorf("%s must be like 90, 90m, 1h30m, or 1.5h", strings.ToLower(label))
	}
	return &value, nil
}

func ParsePositiveIntInput(raw string, required bool, label string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if required {
			return 0, fmt.Errorf("%s is required", label)
		}
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a whole number", label)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s must be non-negative", label)
	}
	if required && value == 0 {
		return 0, fmt.Errorf("%s must be positive", label)
	}
	return value, nil
}

func ParseOptionalDurationHours(raw string, label string) (*float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := parseFlexibleDurationHours(raw)
	if err != nil {
		return nil, fmt.Errorf("%s must be like 7.5h, 7h30m, or 450m", strings.ToLower(label))
	}
	if err := sharedtypes.ValidateSleepHours(value); err != nil {
		return nil, err
	}
	return &value, nil
}

func FormatDurationMinutesInput(minutes *int) string {
	if minutes == nil {
		return ""
	}
	return formatFlexibleDurationMinutes(*minutes)
}

func FormatDurationHoursInput(hours *float64) string {
	if hours == nil {
		return ""
	}
	return formatFlexibleDurationHours(*hours)
}

func FormatDurationSecondsInput(seconds int) string {
	if seconds <= 0 {
		return "0s"
	}
	duration := time.Duration(seconds) * time.Second
	hours := int(duration / time.Hour)
	minutes := int(duration%time.Hour) / int(time.Minute)
	secondsPart := int(duration%time.Minute) / int(time.Second)
	switch {
	case hours > 0 && minutes > 0 && secondsPart > 0:
		return fmt.Sprintf("%dh%02dm%02ds", hours, minutes, secondsPart)
	case hours > 0 && minutes > 0:
		return fmt.Sprintf("%dh%02dm", hours, minutes)
	case hours > 0 && secondsPart > 0:
		return fmt.Sprintf("%dh%02ds", hours, secondsPart)
	case hours > 0:
		return fmt.Sprintf("%dh", hours)
	case minutes > 0 && secondsPart > 0:
		return fmt.Sprintf("%dm%02ds", minutes, secondsPart)
	case minutes > 0:
		return fmt.Sprintf("%dm", minutes)
	default:
		return fmt.Sprintf("%ds", secondsPart)
	}
}

func durationMinutesValue(raw string) (float64, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return 0, nil
	}
	var minutes float64
	if value, err := strconv.Atoi(raw); err == nil {
		minutes = float64(value)
	} else if hours, ok, err := parseBareHours(raw); err != nil {
		return 0, err
	} else if ok {
		minutes = hours * 60
	} else {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return 0, err
		}
		minutes = d.Minutes()
	}
	if math.IsNaN(minutes) || math.IsInf(minutes, 0) || minutes < 0 || minutes >= float64(math.MaxInt/60) {
		return 0, fmt.Errorf("duration must be finite, non-negative, and within supported range")
	}
	return minutes, nil
}

func validateRawDurationLimit(raw string, maxMinutes int, label string) error {
	minutes, err := durationMinutesValue(raw)
	if err != nil {
		return fmt.Errorf("%s must be a valid duration like 90m or 1h30m", label)
	}
	if minutes > float64(maxMinutes) {
		return fmt.Errorf("%s must be between 0 and %dh", label, maxMinutes/60)
	}
	return nil
}

func parseFlexibleDurationMinutes(raw string) (int, error) {
	minutes, err := durationMinutesValue(raw)
	if err != nil {
		return 0, err
	}
	return int(math.Round(minutes)), nil
}

func parseFlexibleDurationHours(raw string) (float64, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return 0, nil
	}
	if hours, ok, err := parseBareHours(raw); err != nil {
		return 0, err
	} else if ok {
		if hours < 0 {
			return 0, fmt.Errorf("duration must be non-negative")
		}
		return hours, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, err
	}
	hours := d.Hours()
	if hours < 0 {
		return 0, fmt.Errorf("duration must be non-negative")
	}
	return hours, nil
}

func parseBareHours(raw string) (float64, bool, error) {
	if strings.ContainsAny(raw, "hm") {
		return 0, false, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, false, nil
	}
	return value, true, nil
}

func formatFlexibleDurationMinutes(minutes int) string {
	if minutes <= 0 {
		return "0m"
	}
	hours := minutes / 60
	remainder := minutes % 60
	switch {
	case hours > 0 && remainder > 0:
		return fmt.Sprintf("%dh%02dm", hours, remainder)
	case hours > 0:
		return fmt.Sprintf("%dh", hours)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}

func formatFlexibleDurationHours(hours float64) string {
	if hours <= 0 {
		return "0h"
	}
	return formatFlexibleDurationMinutes(int(math.Round(hours * 60)))
}
