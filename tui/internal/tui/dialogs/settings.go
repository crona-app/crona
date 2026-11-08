package dialogs

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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

// renderDayBoundaryDialog is one screen: Enabled, the default time and a row per weekday in
// week-start order. A day on the default shows that time dim; its own time reads "custom".
func renderDayBoundaryDialog(theme Theme, state controllerpkg.State) string {
	title, help := "Start of Day", "A new day begins here; work before it counts toward the previous day."
	if state.SettingKey == "endOfDay" {
		title, help = "End of Day", "The day winds down here; end-of-day prompts and summaries use it."
	}
	column := len("Default")
	focused := func(row int) bool { return state.FocusIdx == row }
	enabled := theme.StyleDim.Render("○ off")
	if state.DayBoundaryEnabled {
		enabled = theme.StyleSelected.Render("● on")
	}
	body := []string{
		Field(theme, "Enabled", column, enabled, focused(controllerpkg.DayBoundaryRowEnabled)),
		Field(theme, "Default", column, bareInput(state.Inputs[0], 5, "HH:mm"), focused(controllerpkg.DayBoundaryRowDefault)),
		"",
	}
	focus := -1
	if state.FocusIdx <= controllerpkg.DayBoundaryRowDefault {
		focus = state.FocusIdx
	}
	defaultTime := strings.TrimSpace(state.Inputs[0].Value())
	names := []string{"", "Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	for i, day := range controllerpkg.DayBoundaryWeekdays(state.DueDateWeekStart) {
		row := controllerpkg.DayBoundaryRowDefault + 1 + i
		own := strings.TrimSpace(state.Inputs[day].Value())
		value := ""
		switch {
		case focused(row):
			value = bareInput(state.Inputs[day], 5, defaultTime)
		case own == "" || own == defaultTime:
			value = theme.StyleDim.Render(fallback(defaultTime, "--:--"))
		default:
			value = own
		}
		if own != "" && own != defaultTime {
			value += theme.StyleDim.Render("   custom")
		}
		if focused(row) {
			focus = len(body)
		}
		line := Field(theme, names[day], column, value, focused(row))
		if !state.DayBoundaryEnabled {
			line = theme.StyleDim.Render(ansi.Strip(line))
		}
		body = append(body, line)
	}
	body = append(body, "", theme.StyleDim.Render(help))
	hints := []string{"↑/↓ row", "type a time", "r use default"}
	if focused(controllerpkg.DayBoundaryRowEnabled) {
		hints = []string{"↑/↓ row", "space toggle"}
	}
	d := listDialog{
		wide: true, title: title, context: fallbackTimezone(state.DayBoundaryTimezone),
		focus: focus, body: body, primary: "ctrl+s save", hints: hints,
	}
	if state.ConfirmDiscard {
		d.body = append(d.body, "", theme.StyleError.Render("Discard changes?"))
		d.primary, d.hints, d.esc = "y discard", nil, "n keep editing"
	}
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
