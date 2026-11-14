package daily

import (
	"fmt"
	"strings"

	helperpkg "crona.work/tui/internal/tui/helpers"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	issuecore "crona.work/tui/internal/tui/views/issuecore"
	"crona.work/tui/internal/tui/views/itemdetail"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// previewBody uses list data only; selecting an item never requests its full history.
func previewBody(theme types.Theme, state types.ContentState, width, height int) (string, []string, bool) {
	if state.Pane == "habits" {
		habit, ok := selectedDailyHabit(state)
		if !ok {
			return "Selected habit", []string{theme.StyleDim.Render("No habit selected")}, false
		}
		status := string(habit.Status)
		if status == "" {
			status = "due"
		}
		look := itemdetail.LookOf(theme, status)
		facts := lipgloss.NewStyle().Foreground(look.Color).Render(status)
		if habit.DurationMinutes != nil {
			facts += " · logged " + helperpkg.FormatCompactDurationMinutes(*habit.DurationMinutes)
		}
		if habit.TargetMinutes != nil {
			facts += " · target " + helperpkg.FormatCompactDurationMinutes(*habit.TargetMinutes)
		}
		lines := []string{facts, theme.StyleDim.Render(itemdetail.Context(habit.RepoName, habit.StreamName))}
		text := habit.Notes
		if text == nil || strings.TrimSpace(*text) == "" {
			text = habit.Description
		}
		arcWidth := viewsummary.ArcsWidth(1)
		if height >= 6 && width >= arcWidth+30 {
			arc := itemdetail.HabitTodayArc(theme, habit, status, false)
			lines[0] = lipgloss.NewStyle().Foreground(look.Color).Render(status)
			lines = append(lines, previewExcerpt(theme, text, width-arcWidth-3)...)
			return habit.Name, viewsummary.BesideArcs(viewsummary.Arcs(theme, []viewsummary.Arc{arc}, arcWidth), lines, arcWidth, width), true
		}
		return habit.Name, append(lines, previewExcerpt(theme, text, width)...), true
	}
	issue, ok := selectedDailyIssue(state)
	if !ok {
		return "Selected task", []string{theme.StyleDim.Render("No task selected")}, false
	}
	repo, stream, _, worked := itemdetail.IssueWhere(state, issue)
	status := itemdetail.Status(state, issue.ID, issue.Status)
	style := theme.StyleNormal
	if statusStyle := issuecore.IssueStatusStyle(theme, status); statusStyle != nil {
		style = *statusStyle
	}
	arcWidth := viewsummary.ArcsWidth(1)
	if height >= 6 && width >= arcWidth+30 {
		estimate := 0
		if issue.EstimateMinutes != nil {
			estimate = *issue.EstimateMinutes
		}
		arc := itemdetail.EffortArc(theme, worked, estimate)
		facts := []string{style.Render(issuecore.PlainIssueStatus(status)), theme.StyleDim.Render(itemdetail.Context(repo, stream))}
		facts = append(facts, previewExcerpt(theme, issue.Description, width-arcWidth-3)...)
		return issue.Title, viewsummary.BesideArcs(viewsummary.Arcs(theme, []viewsummary.Arc{arc}, arcWidth), facts, arcWidth, width), true
	}
	facts := style.Render(issuecore.PlainIssueStatus(status)) + " · tracked " + helperpkg.FormatCompactDurationSeconds(worked)
	if issue.EstimateMinutes != nil {
		facts += " / " + helperpkg.FormatCompactDurationMinutes(*issue.EstimateMinutes) + " estimated"
	}
	lines := []string{facts, theme.StyleDim.Render(itemdetail.Context(repo, stream))}
	return issue.Title, append(lines, previewExcerpt(theme, issue.Description, width)...), true
}

func previewExcerpt(theme types.Theme, text *string, width int) []string {
	if text == nil || strings.TrimSpace(*text) == "" {
		return nil
	}
	plain := strings.Join(strings.Fields(*text), " ")
	lines := strings.Split(ansi.Wrap(plain, max(1, width), ""), "\n")
	if len(lines) > 2 {
		lines = lines[:2]
		lines[1] = ansi.Truncate(lines[1], max(1, width-1), "") + "…"
	}
	for i, line := range lines {
		lines[i] = theme.StyleNormal.Render(ansi.Truncate(line, max(1, width), "…"))
	}
	return lines
}

func renderPreview(theme types.Theme, state types.ContentState, width, height int) string {
	if height <= 0 {
		return ""
	}
	title, lines, selected := previewBody(theme, state, max(1, width-2-2*panePad), height)
	meta := ""
	if selected && !state.DailyDetailSummary {
		meta = viewchrome.KeyHint(theme, "enter", "expand")
	}
	title = ansi.Truncate(title, max(1, width-5-ansi.StringWidth(meta)), "…")
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: title, Meta: meta, Width: width, Height: height,
		Lines: padded(lines),
	})
}

// compactSummary keeps the day visible while the selected item's details are expanded.
func compactSummary(theme types.Theme, state types.ContentState, width, height int) string {
	if height <= 0 {
		return ""
	}
	lines := []string{theme.StyleDim.Render("Loading summary…")}
	snapshot := state.SummarySnapshot
	if snapshot != nil && snapshot.Date == state.SummaryDate {
		lines = nil
		if snapshot.Focus != nil {
			lines = append(lines, fmt.Sprintf("Score   %d", snapshot.Focus.Score))
		}
		if m := snapshot.Metrics; m != nil {
			lines = append(lines, "Worked  "+helperpkg.FormatCompactDurationSeconds(m.WorkedSeconds))
			done, total := m.CompletedIssues, m.TotalIssues
			if snapshot.Plan != nil && snapshot.Plan.Date != "" {
				done, total = snapshot.Plan.Summary.CompletedCount, snapshot.Plan.Summary.PlannedCount
			}
			lines = append(lines, fmt.Sprintf("Issues  %d/%d done", done, total),
				fmt.Sprintf("Habits  %d/%d done", m.HabitCompletedCount, m.HabitDueCount))
		}
	}
	return viewchrome.RenderSection(theme, viewchrome.Section{
		Title: dayTitle(state), Meta: viewchrome.KeyHint(theme, "S", "expand"),
		Width: width, Height: height, Lines: padded(lines),
	})
}
