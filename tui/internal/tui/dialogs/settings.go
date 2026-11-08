package dialogs

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"slices"
	"strings"
	"time"

	sharedtypes "crona.work/shared/types"
	controllerpkg "crona.work/tui/internal/tui/dialogs/controller"
)

// stepStrip draws a wizard's numbered steps with the active one selected.
func stepStrip(theme Theme, steps []string, active int) string {
	cells := make([]string, 0, len(steps))
	for i, step := range steps {
		label := fmt.Sprintf(" %d %s ", i+1, step)
		if i == active {
			cells = append(cells, theme.StyleSelected.Render(label))
		} else {
			cells = append(cells, theme.StyleDim.Render(label))
		}
	}
	return strings.Join(cells, " ")
}

// stepContext is the `n/total Step` label a wizard shows in its border.
func stepContext(steps []string, active int) string {
	return fmt.Sprintf("%d/%d %s", active+1, len(steps), steps[active])
}

func renderDateFormatForm(theme Theme, state controllerpkg.State) string {
	inner := formInner(state.Width)
	full, _ := formColumns(inner, len("Pattern"))
	return renderForm(theme, state, FormSpec{
		Title: "Custom Date Format", FocusCount: 1, Primary: plainHint(dialogSubmitHint(state, "save")),
		Rows: []FormRow{
			fieldRow(
				0,
				Field(
					theme,
					"Pattern",
					len("Pattern"),
					bareInput(state.Inputs[0], full, "YYYY-MM-DD"),
					true,
				),
			),
			staticRow(
				"",
				theme.StyleDim.Render("Examples: YYYY-MM-DD   Do MMM YYYY   MM/DD/YYYY   [Week] W"),
			),
		},
	})
}

func checkbox(on bool, label string) string {
	if on {
		return "[x] " + label
	}
	return "[ ] " + label
}

// renderRestProtectionDialog is one screen: which streaks rest protects, the weekly rest
// days in week-start order, and the rest dates from today on (earlier ones are counted).
func renderRestProtectionDialog(theme Theme, state controllerpkg.State) string {
	today := time.Now().Format("2006-01-02")
	upcoming := controllerpkg.RestUpcomingDates(state.ProtectionDates, today)
	past := len(state.ProtectionDates) - len(upcoming)
	column := len("Protect")
	item := func(row, i int, text string) string {
		if state.ProtectionStep == row && state.ProtectionCursor == i {
			return theme.StyleCursor.Render(text)
		}
		return text
	}
	streaks := []string{}
	for i, kind := range sharedtypes.AvailableStreakKinds() {
		streaks = append(streaks, item(0, i, checkbox(hasStreakKind(state.ProtectionStreaks, kind), streakKindLabel(kind))))
	}
	names := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	days := []string{}
	for i, day := range controllerpkg.RestWeekdayOrder(state.DueDateWeekStart) {
		label := " " + names[day] + " "
		if slices.Contains(state.ProtectionWeekdays, day) {
			label = theme.StyleSelected.Render(label)
		} else {
			label = theme.StyleDim.Render(label)
		}
		if state.ProtectionStep == 1 && state.ProtectionCursor == i {
			label = theme.StyleCursor.Render("[") + label + theme.StyleCursor.Render("]")
		} else {
			label = " " + label + " "
		}
		days = append(days, label)
	}
	dates := []string{}
	for i, date := range upcoming {
		dates = append(dates, item(2, i, restDateLabel(date)))
	}
	dates = append(dates, item(2, len(upcoming), "+ add date"))
	body := []string{
		Field(theme, "Protect", column, strings.Join(streaks, "   "), state.ProtectionStep == 0),
		Field(theme, "Rest on", column, strings.Join(days, ""), state.ProtectionStep == 1),
	}
	body = append(body, wrapField(theme, "Dates", column, dates, state.ProtectionStep == 2, formInner(state.Width))...)
	if past > 0 {
		body = append(body, strings.Repeat(" ", column+2)+theme.StyleDim.Render(
			fmt.Sprintf("%d %s in the past", past, map[bool]string{true: "date", false: "dates"}[past == 1])))
	}
	body = append(body, "", theme.StyleDim.Render("Rest days pause the protected streaks instead of breaking them."))
	d := listDialog{
		wide:    true,
		title:   "Rest & Streak Protection",
		context: restProtectionContext(state, len(upcoming)),
		focus:   -1,
		body:    body,
		primary: plainHint(dialogSubmitHint(state, "save")),
		hints:   []string{"↑/↓ row", "←/→ move", "space toggle"},
	}
	switch state.ProtectionStep {
	case 0:
		d.hints = append(d.hints, "a all", "c none")
	case 1:
		d.hints = append(d.hints, "c clear")
	default:
		d.hints = append(d.hints, "a add", "d remove")
	}
	if state.ConfirmDiscard {
		d.body = append(d.body, "", theme.StyleError.Render("Discard changes?"))
		d.primary, d.hints, d.esc = "y discard", nil, "n keep editing"
	}
	return renderListDialog(theme, state, d)
}

func streakKindLabel(kind sharedtypes.StreakKind) string {
	switch kind {
	case sharedtypes.StreakKindCheckInDays:
		return "Check-in days"
	case sharedtypes.StreakKindHabitDays:
		return "Habit days"
	default:
		return "Focus days"
	}
}

func restDateLabel(date string) string {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return parsed.Format("Mon ") + date
}

// wrapField lays items out after a label, joined with " · ", wrapping to the frame width.
func wrapField(theme Theme, label string, column int, items []string, focused bool, inner int) []string {
	width := max(20, inner-column-2)
	lines, line := []string{}, ""
	for _, part := range items {
		next := part
		if line != "" {
			next = line + theme.StyleDim.Render(" · ") + part
		}
		if line != "" && lipgloss.Width(next) > width {
			lines = append(lines, line)
			next = part
		}
		line = next
	}
	lines = append(lines, line)
	out := []string{Field(theme, label, column, lines[0], focused)}
	for _, rest := range lines[1:] {
		out = append(out, strings.Repeat(" ", column+2)+rest)
	}
	return out
}

func restProtectionContext(state controllerpkg.State, upcoming int) string {
	streaks := fmt.Sprintf("%d %s", len(state.ProtectionStreaks),
		map[bool]string{true: "streak", false: "streaks"}[len(state.ProtectionStreaks) == 1])
	days := "no rest days"
	if len(state.ProtectionWeekdays) > 0 {
		names := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
		parts := []string{}
		for _, day := range controllerpkg.RestWeekdayOrder(state.DueDateWeekStart) {
			if slices.Contains(state.ProtectionWeekdays, day) {
				parts = append(parts, names[day])
			}
		}
		days = strings.Join(parts, " ")
	}
	return fmt.Sprintf("%s · %s · %d %s", streaks, days, upcoming,
		map[bool]string{true: "date", false: "dates"}[upcoming == 1])
}

func renderDayBoundaryDialog(theme Theme, state controllerpkg.State) string {
	title := "Start of Day Schedule"
	if state.SettingKey == "endOfDay" {
		title = "End of Day Schedule"
	}
	steps := []string{"Default", "Add Override", "Time"}
	active := min(max(0, state.DayBoundaryStep), len(steps)-1)
	body := []string{stepStrip(theme, steps, active), ""}
	d := listDialog{
		wide:    true,
		title:   title,
		context: fallbackTimezone(state.DayBoundaryTimezone),
		focus:   -1,
	}
	column := len("Enabled")
	switch state.DayBoundaryStep {
	case controllerpkg.DayBoundaryStepDays:
		body = append(body, theme.StyleDim.Render("Select available days"))
		labels, cursor := []string{}, -1
		for day, label := range []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"} {
			day++
			line := checkbox(containsInt(state.DayBoundarySelectedDays, day), label)
			if !controllerpkg.DayBoundaryDayAvailable(state, day) {
				line = theme.StyleDim.Render(line + " (configured)")
			} else if day == state.FocusIdx {
				cursor = day - 1
			}
			labels = append(labels, line)
		}
		if cursor >= 0 {
			d.focus = len(body) + cursor
		}
		body = append(body, selectList(theme, labels, cursor)...)
		d.primary, d.hints, d.esc = "enter next", []string{"↑/↓ move", "space toggle"}, "esc back"
	case controllerpkg.DayBoundaryStepTime:
		full, _ := formColumns(formInner(state.Width), column)
		body = append(
			body,
			Field(
				theme,
				"Days",
				column,
				dayBoundaryDaysLabel(state.DayBoundarySelectedDays),
				false,
			),
			Field(theme, "Time", column, bareInput(state.Inputs[1], full, "HH:mm"), true),
		)
		d.focus = len(body) - 1
		d.primary, d.esc = "ctrl+s apply", "esc back"
	default:
		allOverridden := controllerpkg.AllDayBoundaryDaysOverridden(state.DayBoundarySchedule)
		full, _ := formColumns(formInner(state.Width), column)
		value := bareInput(state.Inputs[0], full, "HH:mm")
		if allOverridden {
			value = theme.StyleDim.Render(state.Inputs[0].Value() + "  (all days have overrides)")
		}
		body = append(body,
			Field(theme, "Enabled", column, enabledLabel(state.DayBoundaryEnabled), false),
			Field(theme, "Default", column, value, state.DayBoundaryOverrideCursor == 0),
			"", theme.StyleDim.Render("Configured overrides"),
		)
		if state.DayBoundaryOverrideCursor == 0 {
			d.focus = len(body) - 3
		}
		if len(state.DayBoundaryOverrides) == 0 {
			body = append(body, theme.StyleDim.Render("None"))
		}
		labels := []string{}
		for _, override := range state.DayBoundaryOverrides {
			labels = append(
				labels,
				fmt.Sprintf("%-22s %s", dayBoundaryDaysLabel(override.Days), override.Time),
			)
		}
		if state.DayBoundaryOverrideCursor > 0 {
			d.focus = len(body) + state.DayBoundaryOverrideCursor - 1
		}
		body = append(body, selectList(theme, labels, state.DayBoundaryOverrideCursor-1)...)
		add := "a add override"
		if allOverridden {
			add = "a add override (unavailable)"
		}
		d.primary, d.hints = "ctrl+s save", []string{"space toggle", add, "enter edit", "D remove"}
	}
	d.body = body
	return renderListDialog(theme, state, d)
}

func renderTelemetrySettingsDialog(theme Theme, state controllerpkg.State) string {
	steps := []string{"Usage", "Diagnostics", "Review"}
	active := min(max(0, state.TelemetryStep), len(steps)-1)
	body := []string{stepStrip(theme, steps, active), ""}
	d := listDialog{
		wide:    true,
		title:   "Privacy & Diagnostics",
		context: stepContext(steps, active),
		focus:   -1,
	}
	switch active {
	case 0:
		body = append(
			body,
			theme.StyleDim.Render(
				"Share anonymous usage signals so Crona can spot problems early and understand which parts of the app people use.",
			),
			theme.StyleDim.Render(
				"This does not include your work content, paths, notes, or anything you type into Crona.",
			),
			"",
		)
		body = append(
			body,
			selectList(
				theme,
				[]string{checkbox(state.TelemetryUsage, "Share usage signals")},
				0,
			)...)
		d.primary, d.hints = "space toggle", []string{"tab next"}
	case 1:
		body = append(
			body,
			theme.StyleDim.Render(
				"Share sanitized diagnostic reports when Crona runs into an error it can explain or recover from.",
			),
			theme.StyleDim.Render(
				"The report is stripped down to help us debug the problem without collecting your work content.",
			),
			"",
		)
		body = append(
			body,
			selectList(theme, []string{checkbox(state.TelemetryErrors, "Share diagnostics")}, 0)...)
		d.primary, d.hints = "space toggle", []string{"tab next", "shift+tab back"}
	default:
		column := len("Usage signals")
		body = append(body,
			Field(theme, "Usage signals", column, telemetryStateLabel(state.TelemetryUsage), false),
			Field(theme, "Diagnostics", column, telemetryStateLabel(state.TelemetryErrors), false),
			"",
			theme.StyleError.Render("Changes take effect after restart."),
			theme.StyleDim.Render("Save now, or restart immediately to apply them right away."),
			"",
		)
		d.focus = len(body) + state.ChoiceCursor
		body = append(
			body,
			selectList(
				theme,
				[]string{"Save", "Save and Restart Now", "Cancel"},
				state.ChoiceCursor,
			)...)
		d.primary, d.hints = "enter choose", []string{"↑/↓ move", "shift+tab back"}
	}
	d.body = body
	return renderListDialog(theme, state, d)
}
