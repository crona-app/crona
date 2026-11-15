// Package checkins renders the global wellbeing check-in history.
package checkins

import (
	"fmt"
	"strings"
	"time"

	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	"crona.work/tui/internal/tui/views/itemdetail"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// Items provides the same searchable date, ratings and notes for every history surface.
func Items(history []api.DailyCheckIn, settings *api.CoreSettings) []string {
	items := make([]string, 0, len(history))
	for _, entry := range history {
		text := fmt.Sprintf("%s %s mood %d energy %d", entry.Date, helperpkg.FormatDisplayDate(entry.Date, settings), entry.Mood, entry.Energy)
		if entry.Notes != nil {
			text += " " + *entry.Notes
		}
		items = append(items, text)
	}
	return items
}

// FilteredIndices returns matches in the history's newest-first order.
func FilteredIndices(history []api.DailyCheckIn, filter string, settings *api.CoreSettings) []int {
	filter = strings.ToLower(strings.TrimSpace(filter))
	indices := []int{}
	for i, item := range Items(history, settings) {
		if filter == "" || strings.Contains(strings.ToLower(item), filter) {
			indices = append(indices, i)
		}
	}
	return indices
}

func clock(raw string) string {
	if at, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return at.Local().Format("15:04")
	}
	return ""
}

// DetailLines is shared by the history detail pane and its scrollable modal.
func DetailLines(theme types.Theme, state types.ContentState, entry *api.DailyCheckIn, width int) []string {
	if entry == nil {
		return []string{theme.StyleDim.Render("Nothing selected")}
	}
	lines := []string{viewsummary.Row(theme, "Date", helperpkg.FormatDisplayDate(entry.Date, state.Settings), width)}
	for _, stamp := range [][2]string{{"Created", entry.CreatedAt}, {"Updated", entry.UpdatedAt}} {
		if stamp[0] == "Updated" && (stamp[1] == "" || stamp[1] == entry.CreatedAt) {
			continue
		}
		if at, err := time.Parse(time.RFC3339Nano, stamp[1]); err == nil {
			value := helperpkg.FormatDisplayDate(at.Local().Format(time.DateOnly), state.Settings) + " " + at.Local().Format("15:04")
			lines = append(lines, viewsummary.Row(theme, stamp[0], value, width))
		}
	}
	return append(append(lines, ""), viewsummary.CheckInLines(theme, entry, state.Settings, width)...)
}

// Render uses the same day-grouped history list and responsive columns as other Log views.
func Render(theme types.Theme, state types.ContentState) string {
	if state.Width < 1 || state.Height < 1 {
		return ""
	}
	indices := FilteredIndices(state.CheckInHistory, state.Filters["checkins"], state.Settings)
	cursor := max(0, min(state.Cursors["checkins"], len(indices)-1))
	var selected *api.DailyCheckIn
	if len(indices) > 0 {
		selected = &state.CheckInHistory[indices[cursor]]
	}
	listW, detailW := viewchrome.LogPaneWidths(state.Width)
	actions := [][2]string{{"/", "filter"}, {"r", "refresh"}}
	if selected != nil {
		selectedActions := [][2]string{{"enter", "details"}}
		if !state.CheckInHistoryReadOnly {
			selectedActions = append(selectedActions, [2]string{"e", "edit"}, [2]string{"D", "delete"})
		}
		actions = append(selectedActions, actions...)
	}
	footerH := len(itemdetail.ActionLines(theme, max(1, listW-2), actions))
	rows := make([]viewchrome.DayRow, 0, len(indices))
	byDate := map[string]string{}
	for _, idx := range indices {
		entry := state.CheckInHistory[idx]
		line := fmt.Sprintf("Mood %d/5 · Energy %d/5", entry.Mood, entry.Energy)
		if entry.SleepHours != nil {
			line += " · Sleep " + helperpkg.FormatCompactDurationHours(*entry.SleepHours)
		}
		if entry.Notes != nil && strings.TrimSpace(*entry.Notes) != "" {
			line += " · " + strings.Join(strings.Fields(*entry.Notes), " ")
		}
		rows = append(rows, viewchrome.DayRow{Date: entry.Date, Line: ansi.Truncate(line, max(1, listW-4), "…")})
		byDate[entry.Date] = clock(entry.CreatedAt)
	}
	lines := viewchrome.DateColumnList(theme, rows, cursor, listW-2, max(1, state.Height-1-footerH), func(date string) string { return byDate[date] })
	if len(rows) == 0 {
		message := "No check-ins recorded"
		switch {
		case state.CheckInHistoryLoading:
			message = "Loading check-ins…"
		case state.CheckInHistoryError != "" && len(state.CheckInHistory) == 0:
			message = "Unable to load check-ins. Press r to retry."
		case strings.TrimSpace(state.Filters["checkins"]) != "":
			message = "No check-ins match this filter"
		}
		lines = []string{theme.StyleDim.Render(message)}
	} else if state.CheckInHistoryError != "" {
		lines = append([]string{theme.StyleDim.Render("Refresh failed · r to retry")}, lines...)
	}
	list := itemdetail.RenderPane(theme, viewchrome.Section{
		Title: "Check-ins", Meta: viewhelpers.FilterMeta(fmt.Sprint(len(indices)), state.Filters["checkins"]),
		Width: listW, Height: state.Height, Focused: true, Lines: lines,
	}, actions)
	if detailW == 0 {
		return list
	}
	detail := viewchrome.RenderSection(theme, viewchrome.Section{Title: "Detail", Width: detailW, Height: state.Height, Lines: DetailLines(theme, state, selected, detailW-2)})
	return viewchrome.JoinColumns(theme, state.Height, list, detail)
}
