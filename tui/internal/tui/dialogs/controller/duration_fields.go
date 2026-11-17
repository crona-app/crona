package controller

import "fmt"

func durationField(state State) (index int, label string, hours, required bool) {
	index = -1
	input, ok := dialogInputIndex(state, state.FocusIdx)
	if !ok {
		return
	}
	switch state.Kind {
	case "create_checkin", "edit_checkin":
		if input == 2 {
			index, label, hours = 2, "Sleep", true
		}
		if input == 4 {
			index, label = 4, "Screen time"
		}
	case "create_habit":
		if input == 4 {
			index, label = 4, "Target"
		}
	case "edit_habit":
		if input == 2 {
			index, label = 2, "Target"
		}
	case "complete_habit":
		if input == 0 {
			index, label = 0, "Duration"
		}
	case "manual_session":
		if input == 2 || input == 3 {
			index, label, required = input, "Session duration", input == 2
		}
	case "timer_countdown_start", "timer_countdown_extend":
		if input == 0 {
			index, label, required = 0, "Timer duration", true
		}
	case "pomodoro_start", "hard_limit_extend":
		if custom, ok := pomodoroDialogInputIndex(state, state.FocusIdx); ok && custom < 3 {
			index, label, required = custom, "Timer duration", true
		}
	}
	return
}

func validateFocusedDuration(state State) error {
	index, label, hours, required := durationField(state)
	if index < 0 || index >= len(state.Inputs) {
		return nil
	}
	if hours {
		_, err := ParseOptionalDurationHours(state.Inputs[index].Value(), label)
		return err
	}
	if required {
		_, err := ParseDurationInput(state.Inputs[index].Value(), true, label)
		return err
	}
	_, err := ParseOptionalDurationMinutes(state.Inputs[index].Value(), label)
	return err
}

// DurationFieldHint describes the focused duration without changing the entered text.
func DurationFieldHint(state State) string {
	index, label, hours, _ := durationField(state)
	if index < 0 || index >= len(state.Inputs) {
		return ""
	}
	return DurationInputHint(state.Inputs[index].Value(), label, hours)
}

// DurationInputHint explains units, normalization, and the physical daily ceiling.
func DurationInputHint(raw, label string, hours bool) string {
	normalized := ""
	if hours {
		if value, err := ParseOptionalDurationHours(raw, label); err == nil && value != nil {
			normalized = FormatDurationHoursInput(value)
		}
	} else {
		if value, err := ParseOptionalDurationMinutes(raw, label); err == nil && value != nil {
			normalized = FormatDurationMinutesInput(value)
		}
	}
	units := "bare numbers are minutes"
	if hours {
		units = "bare numbers are hours"
	}
	if normalized != "" {
		return fmt.Sprintf("%s · %s · max 24h", normalized, units)
	}
	return units + " · e.g. 1h30m · max 24h"
}
