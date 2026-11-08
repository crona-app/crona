package controller

import (
	"fmt"
	"slices"
	"strings"
	"time"

	sharedtypes "crona.work/shared/types"

	tea "github.com/charmbracelet/bubbletea"
)

func updateDatePicker(
	state State,
	ctx UpdateContext,
	currentDate string,
	msg tea.KeyMsg,
) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		if state.Parent == "shared_date" {
			return Close(state), nil, ""
		}
		return closeDatePicker(state), nil, ""
	case "enter", " ":
		selected := state.DateCursorValue
		if isIssueDueDatePicker(state) {
			if err := ValidateDueDate(state, selected); err != nil {
				state.ErrorMessage = err.Error()
				return state, nil, ""
			}
			if state.Parent != "" {
				if idx, ok := dialogInputIndex(state, state.FocusIdx); ok {
					state.Inputs[idx].SetValue(selected)
				}
				return closeDatePicker(state), nil, ""
			}
		}
		if state.Parent == "edit_rest_protection" {
			state.ProtectionDates = normalizedDateList(append(state.ProtectionDates, selected))
			return closeDatePicker(state), nil, ""
		}
		if state.Parent == "shared_date" {
			return Close(
					state,
				), &Action{
					Kind:    "set_shared_date",
					DueDate: ValueToPointer(selected),
				}, ""
		}
		return Close(
				state,
			), &Action{
				Kind:     "set_issue_todo_date",
				IssueID:  state.IssueID,
				StreamID: ctx.ActiveIssueStream,
				DueDate:  ValueToPointer(selected),
			}, ""
	case "backspace", "delete", "c":
		if isIssueDueDatePicker(state) && state.Parent != "" {
			if idx, ok := dialogInputIndex(state, state.FocusIdx); ok {
				state.Inputs[idx].SetValue("")
			}
			return closeDatePicker(state), nil, ""
		}
		if state.Parent == "edit_rest_protection" {
			return closeDatePicker(state), nil, ""
		}
		if state.Parent == "shared_date" {
			return Close(state), nil, ""
		}
		return Close(
				state,
			), &Action{
				Kind:     "set_issue_todo_date",
				IssueID:  state.IssueID,
				StreamID: ctx.ActiveIssueStream,
				DueDate:  ValueToPointer(""),
			}, ""
	case "left":
		return shiftDatePicker(state, 0, 0, -1), nil, ""
	case "right":
		return shiftDatePicker(state, 0, 0, 1), nil, ""
	case "up":
		return shiftDatePicker(state, 0, 0, -7), nil, ""
	case "down":
		return shiftDatePicker(state, 0, 0, 7), nil, ""
	case ",":
		state = shiftDatePicker(state, 0, -1, 0)
		return state, summaryCalendarMonthAction(state), ""
	case ".":
		state = shiftDatePicker(state, 0, 1, 0)
		return state, summaryCalendarMonthAction(state), ""
	case "g":
		today := currentDate
		if isIssueDueDatePicker(state) {
			today = strings.TrimSpace(state.DueDateToday)
			if today == "" {
				today = currentDate
			}
			if err := ValidateDueDateAt(state, today, today); err != nil {
				state.ErrorMessage = err.Error()
				return state, nil, ""
			}
		}
		return OpenDatePicker(
			state,
			state.Parent,
			state.IssueID,
			state.FocusIdx,
			ValueToPointer(today),
			today,
		), nil, ""
	}
	return state, nil, ""
}

func summaryCalendarMonthAction(state State) *Action {
	month := DialogMonth(state, state.DateCursorValue)
	start := month.Format("2006-01-02")
	end := month.AddDate(0, 1, -1).Format("2006-01-02")
	return &Action{Kind: "load_summary_calendar_month", StartDate: start, EndDate: end}
}

func updateSummaryCalendar(
	state State,
	currentDate string,
	msg tea.KeyMsg,
) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "enter", " ":
		selected := state.DateCursorValue
		if current, err := time.Parse("2006-01-02", currentDate); err == nil {
			if candidate, err := time.Parse("2006-01-02", selected); err == nil &&
				candidate.After(current) {
				state.ErrorMessage = "Future dates are not available"
				return state, nil, ""
			}
		}
		return Close(
				state,
			), &Action{
				Kind:    "set_summary_date",
				DueDate: ValueToPointer(selected),
			}, ""
	case "left":
		return summaryCalendarMove(state, currentDate, -1), nil, ""
	case "right":
		return summaryCalendarMove(state, currentDate, 1), nil, ""
	case "up":
		return summaryCalendarMove(state, currentDate, -7), nil, ""
	case "down":
		return summaryCalendarMove(state, currentDate, 7), nil, ""
	case ",":
		state = summaryCalendarShiftMonth(state, currentDate, -1)
		return state, summaryCalendarMonthAction(state), ""
	case ".":
		state = summaryCalendarShiftMonth(state, currentDate, 1)
		return state, summaryCalendarMonthAction(state), ""
	case "g":
		return OpenSummaryCalendar(state, currentDate), nil, ""
	}
	return state, nil, ""
}

func summaryCalendarMove(state State, currentDate string, days int) State {
	month := DialogMonth(state, currentDate)
	candidate := DialogDate(state, currentDate).AddDate(0, 0, days)
	return summaryCalendarSetValidDate(state, candidate, month, currentDate)
}

func summaryCalendarShiftMonth(state State, currentDate string, months int) State {
	currentMonth := DialogMonth(state, currentDate)
	targetMonth := currentMonth.AddDate(0, months, 0)
	today, _ := time.Parse("2006-01-02", currentDate)
	if !today.IsZero() &&
		targetMonth.After(time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())) {
		return state
	}
	day := DialogDate(state, currentDate).Day()
	candidate := time.Date(
		targetMonth.Year(),
		targetMonth.Month(),
		day,
		0,
		0,
		0,
		0,
		targetMonth.Location(),
	)
	return summaryCalendarSetValidDate(state, candidate, targetMonth, currentDate)
}

func summaryCalendarSetValidDate(
	state State,
	candidate, month time.Time,
	currentDate string,
) State {
	monthStart := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
	monthEnd := monthStart.AddDate(0, 1, -1)
	today, _ := time.Parse("2006-01-02", currentDate)
	maxDate := monthEnd
	if !today.IsZero() && monthStart.Year() == today.Year() && monthStart.Month() == today.Month() {
		maxDate = today
	}
	if candidate.Before(monthStart) {
		candidate = monthStart
	}
	if candidate.After(maxDate) {
		candidate = maxDate
	}
	state.DateCursorValue = candidate.Format("2006-01-02")
	state.DateMonthValue = monthStart.Format("2006-01-02")
	return state
}

func ternaryDir(key string) int {
	if key == "shift+tab" {
		return -1
	}
	return 1
}

func closeDatePicker(state State) State {
	parent := state.Parent
	dateField := state.FocusIdx
	state.Kind = parent
	state.Parent = ""
	state.DateMonthValue = ""
	state.DateCursorValue = ""
	if parent == "create_issue_meta" || parent == "create_issue_default" || parent == "edit_issue" {
		state.FocusIdx = dateField
		state = SyncDialogFocus(state)
		return state
	}
	for i := range state.Inputs {
		if i == state.FocusIdx {
			state.Inputs[i].Focus()
		} else {
			state.Inputs[i].Blur()
		}
	}
	return state
}

func shiftDatePicker(state State, years, months, days int) State {
	selected := DialogDate(state, time.Now().Format("2006-01-02")).AddDate(years, months, days)
	monthStart := time.Date(selected.Year(), selected.Month(), 1, 0, 0, 0, 0, selected.Location())
	state.DateCursorValue = selected.Format("2006-01-02")
	state.DateMonthValue = monthStart.Format("2006-01-02")
	return state
}

// Rest & Streak Protection is one screen of three rows: ProtectionStep is the row (Protect,
// Rest on, Dates) and ProtectionCursor the item within it. Dates lists only today and later;
// the last slot adds a date.
const (
	restRowStreaks = iota
	restRowWeekdays
	restRowDates
)

func updateRestProtection(
	state State,
	currentDate string,
	msg tea.KeyMsg,
) (State, *Action, string) {
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
	if isDialogSubmitKey(state, key) {
		return Close(state), &Action{
			Kind:        "patch_rest_protection",
			StreakKinds: streakKindStrings(state.ProtectionStreaks),
			IntList:     normalizedWeekdays(state.ProtectionWeekdays),
			RestDates:   normalizedDateList(state.ProtectionDates),
		}, ""
	}
	today := restToday(currentDate)
	switch key {
	case "esc":
		if restProtectionFingerprint(state) != state.DiscardBaseline {
			state.ConfirmDiscard = true
			return state, nil, ""
		}
		return Close(state), nil, ""
	case "down", "tab":
		state.ProtectionStep = (state.ProtectionStep + 1) % 3
		state.ProtectionCursor = 0
	case "up", "shift+tab":
		state.ProtectionStep = (state.ProtectionStep + 2) % 3
		state.ProtectionCursor = 0
	case "left", "right":
		dir := 1
		if key == "left" {
			dir = -1
		}
		state.ProtectionCursor = ShiftSelection(state.ProtectionCursor, restRowLen(state, today), dir)
	case " ", "x", "enter":
		return restProtectionActivate(state, currentDate, today)
	case "a":
		switch state.ProtectionStep {
		case restRowStreaks:
			state.ProtectionStreaks = append([]sharedtypes.StreakKind(nil), sharedtypes.AvailableStreakKinds()...)
		case restRowDates:
			return openRestDatePicker(state, currentDate), nil, ""
		}
	case "c":
		switch state.ProtectionStep {
		case restRowStreaks:
			state.ProtectionStreaks = nil
		case restRowWeekdays:
			state.ProtectionWeekdays = nil
		}
	case "d", "D", "backspace", "delete":
		if state.ProtectionStep == restRowDates {
			upcoming := RestUpcomingDates(state.ProtectionDates, today)
			if state.ProtectionCursor < len(upcoming) {
				state.ProtectionDates = slices.DeleteFunc(slices.Clone(state.ProtectionDates),
					func(date string) bool { return date == upcoming[state.ProtectionCursor] })
				state.ProtectionCursor = min(state.ProtectionCursor, max(0, restRowLen(state, today)-1))
			}
		}
	}
	return clearDialogError(state), nil, ""
}

func restProtectionActivate(state State, currentDate, today string) (State, *Action, string) {
	switch state.ProtectionStep {
	case restRowStreaks:
		kinds := sharedtypes.AvailableStreakKinds()
		if state.ProtectionCursor < len(kinds) {
			state.ProtectionStreaks = toggleStreakKind(state.ProtectionStreaks, kinds[state.ProtectionCursor])
		}
	case restRowWeekdays:
		day := RestWeekdayOrder(state.DueDateWeekStart)[min(state.ProtectionCursor, 6)]
		state.ProtectionWeekdays = toggleWeekday(state.ProtectionWeekdays, day)
	case restRowDates:
		if state.ProtectionCursor >= len(RestUpcomingDates(state.ProtectionDates, today)) {
			return openRestDatePicker(state, currentDate), nil, ""
		}
	}
	return clearDialogError(state), nil, ""
}

func openRestDatePicker(state State, currentDate string) State {
	return OpenDatePicker(state, "edit_rest_protection", 0, 0,
		ValueToPointer(restToday(currentDate)), currentDate)
}

func restToday(currentDate string) string {
	if strings.TrimSpace(currentDate) == "" {
		return time.Now().Format("2006-01-02")
	}
	return currentDate
}

func restRowLen(state State, today string) int {
	switch state.ProtectionStep {
	case restRowStreaks:
		return len(sharedtypes.AvailableStreakKinds())
	case restRowWeekdays:
		return 7
	default:
		return len(RestUpcomingDates(state.ProtectionDates, today)) + 1
	}
}

// RestWeekdayOrder lists weekday numbers (0 = Sunday) starting on the week start.
func RestWeekdayOrder(start sharedtypes.WeekStart) []int {
	if sharedtypes.NormalizeWeekStart(start) == sharedtypes.WeekStartSunday {
		return []int{0, 1, 2, 3, 4, 5, 6}
	}
	return []int{1, 2, 3, 4, 5, 6, 0}
}

// RestUpcomingDates keeps the rest dates from today on; earlier ones stay saved but unshown.
func RestUpcomingDates(dates []string, today string) []string {
	out := []string{}
	for _, date := range normalizedDateList(dates) {
		if date >= today {
			out = append(out, date)
		}
	}
	return out
}

func restProtectionFingerprint(state State) string {
	return fmt.Sprint(streakKindStrings(state.ProtectionStreaks), normalizedWeekdays(state.ProtectionWeekdays),
		normalizedDateList(state.ProtectionDates))
}
