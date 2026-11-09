package itemdetail

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	issuecore "crona.work/tui/internal/tui/views/issuecore"
	momentumhelpers "crona.work/tui/internal/tui/views/momentum"
	viewruntime "crona.work/tui/internal/tui/views/runtime"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	habitTrailDays = 14
	habitHitWindow = 30
)

// HabitLook is how a habit's status on a day shows: its colour, its row's checkbox and its
// mark in a history.
type HabitLook struct {
	Color    lipgloss.Color
	Checkbox string
	Mark     string
}

// LookOf is the look of a habit status: green done, red failed, yellow anything else.
func LookOf(theme types.Theme, status string) HabitLook {
	switch status {
	case "completed":
		return HabitLook{theme.ColorGreen, "[x]", "✓"}
	case "failed":
		return HabitLook{theme.ColorRed, "[!]", "✗"}
	default:
		return HabitLook{theme.ColorYellow, "[ ]", "·"}
	}
}

// habitStatus is the word for a due habit's status today.
func habitStatus(habit api.HabitDailyItem) string {
	switch habit.Status {
	case "completed", "failed":
		return string(habit.Status)
	default:
		return "due"
	}
}

// HabitLines are a due habit's details.
func HabitLines(theme types.Theme, state types.ContentState, habit api.HabitDailyItem, width int) []string {
	return habitLines(theme, state, habit, habitStatus(habit), width)
}

// DefinitionLines are a habit's details from its definition: as due today when it is, else
// "not due today", or "paused" while inactive.
func DefinitionLines(theme types.Theme, state types.ContentState, habit api.HabitWithMeta, width int) []string {
	for _, due := range state.DueHabits {
		if due.ID == habit.ID {
			return HabitLines(theme, state, due, width)
		}
	}
	status := "not due today"
	if !habit.Active {
		status = "paused"
	}
	return habitLines(theme, state, api.HabitDailyItem{HabitWithMeta: habit}, status, width)
}

// HabitMomentumLines are the habit's momentum section, ladders included.
func HabitMomentumLines(theme types.Theme, state types.ContentState, habitID int64, width int) []string {
	return momentumLines(theme, state, momentumhelpers.ForHabit(state.MomentumCards, habitID), width)
}

// habitLines lay a habit out like an issue: a Today arc of minutes logged against the target
// (or done or not), a hit-rate arc over the last 30 days, and the facts beside them; then a
// two-week trail, recent entries, notes and momentum.
func habitLines(theme types.Theme, state types.ContentState, habit api.HabitDailyItem, status string, width int) []string {
	days := habitDays(state, habit.ID)
	today := viewruntime.ShownDate(state)
	off := status == "not due today" || status == "paused"
	arcs := []viewsummary.Arc{habitTodayArc(theme, habit, status, off)}
	if hit, ok := habitHitArc(theme, days, today); ok {
		arcs = append(arcs, hit)
	}
	facts := []string{habitStatusLine(theme, habit, status, off)}
	if context := Context(habit.RepoName, habit.StreamName); context != "-" {
		facts = append(facts, theme.StyleNormal.Render(context))
	}
	facts = append(facts, idLine(theme, habit.ID, nil))
	if len(habit.Tags) > 0 {
		facts = append(facts, issuecore.TagChips(theme, habit.Tags, state.Tags, max(8, width-viewsummary.ArcsWidth(2)-3)))
	}
	lines := description(theme, habit.Description, width)
	lines = append(lines, besideArcs(theme, arcs, facts, width)...)
	lines = append(lines, habitTrailLines(theme, days, today, width)...)
	lines = append(lines, habitRecentLines(theme, days, habit, today, width)...)
	lines = append(lines, titledText(theme, "Notes", habit.Notes, width)...)
	return append(lines, HabitMomentumLines(theme, state, habit.ID, width)...)
}

type habitDay struct {
	status  string
	minutes int
}

// habitDays fold the habit's history into one entry per date; a completion wins the day.
func habitDays(state types.ContentState, habitID int64) map[string]habitDay {
	days := map[string]habitDay{}
	history := state.HabitHistory
	if state.DetailHistoryHabitID == habitID {
		history = state.DetailHabitHistory
	}
	for _, entry := range history {
		if entry.HabitID != habitID {
			continue
		}
		day := days[entry.Date]
		if entry.DurationMinutes != nil {
			day.minutes += *entry.DurationMinutes
		}
		if status := string(entry.Status); status != "" && day.status != "completed" {
			day.status = status
		}
		days[entry.Date] = day
	}
	return days
}

// habitTodayArc fills with minutes logged against the target, or full once done without one.
func habitTodayArc(theme types.Theme, habit api.HabitDailyItem, status string, off bool) viewsummary.Arc {
	arc := viewsummary.Arc{Label: "Today", Color: LookOf(theme, string(habit.Status)).Color}
	logged := 0
	if habit.DurationMinutes != nil {
		logged = *habit.DurationMinutes
	}
	switch habit.Status {
	case "completed":
		arc.Fraction, arc.Value = 1, "✓ done"
	case "failed":
		arc.Value = "✗"
	default:
		arc.Value = status
		if off {
			arc.Value = "off"
		}
	}
	if habit.TargetMinutes != nil && *habit.TargetMinutes > 0 {
		arc.Label += "/" + helperpkg.FormatCompactDurationMinutes(*habit.TargetMinutes)
		arc.Fraction = float64(logged) / float64(*habit.TargetMinutes)
		if logged > 0 || habit.Status == "" {
			arc.Value = helperpkg.FormatCompactDurationMinutes(logged)
		}
		if habit.Status == "failed" {
			arc.Value = "✗ " + helperpkg.FormatCompactDurationMinutes(logged)
		}
	}
	return arc
}

// habitHitArc is the share of the last 30 days with an outcome that the habit was done.
func habitHitArc(theme types.Theme, days map[string]habitDay, today string) (viewsummary.Arc, bool) {
	end, err := time.Parse(time.DateOnly, today)
	if err != nil {
		return viewsummary.Arc{}, false
	}
	done, decided := 0, 0
	for i := range habitHitWindow {
		day := days[end.AddDate(0, 0, -i).Format(time.DateOnly)]
		if day.status == "" {
			continue
		}
		decided++
		if day.status == "completed" {
			done++
		}
	}
	if decided == 0 {
		return viewsummary.Arc{}, false
	}
	fraction := float64(done) / float64(decided)
	color := theme.ColorRed
	switch {
	case fraction >= 0.8:
		color = theme.ColorGreen
	case fraction >= 0.5:
		color = theme.ColorYellow
	}
	return viewsummary.Arc{
		Label: fmt.Sprintf("Hit/%dd", habitHitWindow), Value: fmt.Sprintf("%.0f%%", fraction*100),
		Fraction: fraction, Color: color,
	}, true
}

// habitStatusLine is the status as a coloured dot, then the schedule; off days are subtle.
func habitStatusLine(theme types.Theme, habit api.HabitDailyItem, status string, off bool) string {
	color := LookOf(theme, string(habit.Status)).Color
	if off {
		color = theme.ColorSubtle
	}
	line := lipgloss.NewStyle().Foreground(color).Render("● " + status)
	if schedule := Schedule(habit.Habit); schedule != "" {
		line += theme.StyleDim.Render(" · " + schedule)
	}
	return line
}

// Schedule reads the schedule in words: daily, weekdays, or the weekly days.
func Schedule(habit api.Habit) string {
	if habit.ScheduleType != "weekly" || len(habit.Weekdays) == 0 {
		return strings.ReplaceAll(string(habit.ScheduleType), "_", " ")
	}
	names := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	days := []string{}
	for _, day := range habit.Weekdays {
		if day >= 0 && day < len(names) {
			days = append(days, names[day])
		}
	}
	return strings.Join(days, " ")
}

// habitTrailLines are the last two weeks as marks, oldest first.
func habitTrailLines(theme types.Theme, days map[string]habitDay, today string, width int) []string {
	end, err := time.Parse(time.DateOnly, today)
	if err != nil {
		return nil
	}
	marks := make([]string, 0, habitTrailDays)
	for i := habitTrailDays - 1; i >= 0; i-- {
		status := days[end.AddDate(0, 0, -i).Format(time.DateOnly)].status
		switch status {
		case "completed":
			marks = append(marks, lipgloss.NewStyle().Foreground(theme.ColorGreen).Render("●"))
		case "failed":
			marks = append(marks, lipgloss.NewStyle().Foreground(theme.ColorRed).Render("✗"))
		default:
			marks = append(marks, theme.StyleDim.Render("·"))
		}
	}
	row := label(theme, fmt.Sprintf("Last %dd", habitTrailDays), 11) + strings.Join(marks, " ") + theme.StyleDim.Render("  today")
	return []string{"", ansi.Truncate(row, width, "…")}
}

// habitRecentLines list the latest days with an outcome or time, newest first: date, a bar of
// minutes against the target (or the longest listed), the minutes and the status.
func habitRecentLines(theme types.Theme, days map[string]habitDay, habit api.HabitDailyItem, today string, width int) []string {
	dates := []string{}
	for date, day := range days {
		if date <= today && (day.status != "" || day.minutes > 0) {
			dates = append(dates, date)
		}
	}
	if len(dates) == 0 {
		return append(heading(theme, "Recent"), "  "+theme.StyleDim.Render("nothing logged yet"))
	}
	slices.Sort(dates)
	slices.Reverse(dates)
	shown := dates[:min(len(dates), recentRows)]
	scale := 1
	if habit.TargetMinutes != nil && *habit.TargetMinutes > 0 {
		scale = *habit.TargetMinutes
	}
	for _, date := range shown {
		scale = max(scale, days[date].minutes)
	}
	lines := heading(theme, "Recent")
	for _, date := range shown {
		day := days[date]
		when := date
		if parsed, err := time.Parse(time.DateOnly, date); err == nil {
			when = parsed.Format("Mon 2 Jan")
		}
		look := LookOf(theme, day.status)
		minutes := "-"
		if day.minutes > 0 {
			minutes = helperpkg.FormatCompactDurationMinutes(day.minutes)
		}
		row := "  " + label(theme, when, 11) + "  " + Meter(theme, look.Color, day.minutes, scale, barCells) + "  " +
			theme.StyleNormal.Render(viewhelpers.PadCells(minutes, 6)) + " " + lipgloss.NewStyle().Foreground(look.Color).Render(look.Mark)
		lines = append(lines, ansi.Truncate(row, width, "…"))
	}
	if more := len(dates) - len(shown); more > 0 {
		lines = append(lines, "  "+theme.StyleDim.Render(fmt.Sprintf("+%d earlier", more)))
	}
	return lines
}
