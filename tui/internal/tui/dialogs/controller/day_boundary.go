package controller

import (
	"fmt"
	"strings"

	sharedtypes "crona.work/shared/types"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// The Start/End of Day schedule is one screen. FocusIdx is the row: Enabled, Default, then
// one row per weekday in week-start order. Inputs[0] holds the default time and Inputs[d]
// weekday d's own time (1 = Monday … 7 = Sunday); an empty day input follows the default.
const (
	DayBoundaryRowEnabled = iota
	DayBoundaryRowDefault
	dayBoundaryRowFirstDay
	dayBoundaryRowCount = dayBoundaryRowFirstDay + 7
)

func OpenEditDayBoundary(
	state State,
	key sharedtypes.CoreSettingsKey,
	schedule sharedtypes.DayBoundarySchedule,
	timezone string,
) State {
	weekStart := state.DueDateWeekStart
	state = Close(state)
	state.DueDateWeekStart = weekStart
	schedule = schedule.Normalized()
	state.Kind = "edit_day_boundary"
	state.SettingKey = key
	state.DayBoundaryEnabled = schedule.Enabled
	state.DayBoundaryTimezone = timezone
	state.DayBoundarySchedule = schedule
	state.Inputs = make([]textinput.Model, 8)
	for day := range 8 {
		value := schedule.DefaultTime
		if day > 0 {
			value = schedule.WeekdayOverrides[day]
		}
		input := textinput.New()
		input.Prompt = ""
		input.CharLimit = 5
		input.Width = 5
		input.Placeholder = "HH:mm"
		input.SetValue(value)
		state.Inputs[day] = input
	}
	state.FocusIdx = DayBoundaryRowDefault
	state.DiscardBaseline = dayBoundaryFingerprint(state)
	return SyncDialogFocus(state)
}

// DayBoundaryWeekdays lists weekday numbers (1 = Monday … 7 = Sunday) in week-start order.
func DayBoundaryWeekdays(start sharedtypes.WeekStart) []int {
	if sharedtypes.NormalizeWeekStart(start) == sharedtypes.WeekStartSunday {
		return []int{7, 1, 2, 3, 4, 5, 6}
	}
	return []int{1, 2, 3, 4, 5, 6, 7}
}

// DayBoundaryInputForRow maps a row to its time input.
func DayBoundaryInputForRow(state State, row int) (int, bool) {
	idx := -1
	switch {
	case row == DayBoundaryRowDefault:
		idx = 0
	case row >= dayBoundaryRowFirstDay && row < dayBoundaryRowCount:
		idx = DayBoundaryWeekdays(state.DueDateWeekStart)[row-dayBoundaryRowFirstDay]
	}
	return idx, idx >= 0 && idx < len(state.Inputs)
}

func dayBoundaryFingerprint(state State) string {
	parts := []string{fmt.Sprint(state.DayBoundaryEnabled)}
	for _, input := range state.Inputs {
		parts = append(parts, strings.TrimSpace(input.Value()))
	}
	return strings.Join(parts, "|")
}

func updateDayBoundary(state State, msg tea.KeyMsg) (State, *Action, string) {
	key := msg.String()
	if state.ConfirmDiscard {
		switch key {
		case "y", "enter":
			return Close(state), nil, ""
		case "n", "esc":
			state.ConfirmDiscard = false
		}
		return state, nil, ""
	}
	switch key {
	case "esc":
		if dayBoundaryFingerprint(state) != state.DiscardBaseline {
			state.ConfirmDiscard = true
			return state, nil, ""
		}
		return Close(state), nil, ""
	case "up", "shift+tab":
		state.FocusIdx = (state.FocusIdx + dayBoundaryRowCount - 1) % dayBoundaryRowCount
		return SyncDialogFocus(blurDayBoundary(state)), nil, ""
	case "down", "tab", "enter":
		state.FocusIdx = (state.FocusIdx + 1) % dayBoundaryRowCount
		return SyncDialogFocus(blurDayBoundary(state)), nil, ""
	case "ctrl+s":
		schedule, err := dayBoundaryScheduleFromState(state)
		if err != nil {
			state.ErrorMessage = err.Error()
			return state, nil, ""
		}
		return Close(state), &Action{
			Kind:                "patch_setting",
			SettingKey:          state.SettingKey,
			DayBoundarySchedule: schedule,
		}, ""
	}
	if state.FocusIdx == DayBoundaryRowEnabled {
		if key == " " || key == "left" || key == "right" {
			state.DayBoundaryEnabled = !state.DayBoundaryEnabled
		}
		return clearDialogError(state), nil, ""
	}
	idx, ok := DayBoundaryInputForRow(state, state.FocusIdx)
	if !ok {
		return state, nil, ""
	}
	if key == "r" && idx > 0 {
		state.Inputs[idx].SetValue("")
		return clearDialogError(state), nil, ""
	}
	if msg.Type == tea.KeyRunes && strings.Trim(string(msg.Runes), "0123456789:") != "" {
		return state, nil, ""
	}
	// A full time is replaced by typing, so a row can be retyped without clearing it first.
	if msg.Type == tea.KeyRunes && len(state.Inputs[idx].Value()) >= 5 {
		state.Inputs[idx].SetValue("")
	}
	state.Inputs[idx], _ = state.Inputs[idx].Update(msg)
	return clearDialogError(state), nil, ""
}

func blurDayBoundary(state State) State {
	for i := range state.Inputs {
		state.Inputs[i].Blur()
	}
	return state
}

var dayBoundaryNames = []string{"", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

func dayBoundaryScheduleFromState(state State) (sharedtypes.DayBoundarySchedule, error) {
	if len(state.Inputs) < 8 {
		return sharedtypes.DayBoundarySchedule{}, fmt.Errorf("schedule editor is unavailable")
	}
	schedule := sharedtypes.DayBoundarySchedule{
		Enabled:          state.DayBoundaryEnabled,
		DefaultTime:      strings.TrimSpace(state.Inputs[0].Value()),
		WeekdayOverrides: map[int]string{},
	}
	if !validDayBoundaryTime(schedule.DefaultTime) {
		return schedule, fmt.Errorf("default time: use HH:mm, e.g. 06:00")
	}
	for day := 1; day <= 7; day++ {
		value := strings.TrimSpace(state.Inputs[day].Value())
		if value == "" || value == schedule.DefaultTime {
			continue
		}
		if !validDayBoundaryTime(value) {
			return schedule, fmt.Errorf("%s: use HH:mm, or r to use the default", dayBoundaryNames[day])
		}
		schedule.WeekdayOverrides[day] = value
	}
	if err := schedule.Validate(); err != nil {
		return sharedtypes.DayBoundarySchedule{}, err
	}
	return schedule.Normalized(), nil
}

func validDayBoundaryTime(value string) bool {
	return sharedtypes.DayBoundarySchedule{DefaultTime: value}.Validate() == nil
}
