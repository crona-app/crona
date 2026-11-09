package itemdetail

import (
	"strings"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	issuecore "crona.work/tui/internal/tui/views/issuecore"
	momentumhelpers "crona.work/tui/internal/tui/views/momentum"
	viewruntime "crona.work/tui/internal/tui/views/runtime"
	sessionmeta "crona.work/tui/internal/tui/views/sessionmeta"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
)

// IssueLines are an issue's details: its description, the effort and penalty arcs beside its
// facts (status, context, tags, the penalty's verdict), then the penalty, sessions, carry-over,
// notes and the momentums it feeds.
func IssueLines(theme types.Theme, state types.ContentState, issue api.Issue, width int) []string {
	repo, stream, repoID, worked := IssueWhere(state, issue)
	status := Status(state, issue.ID, issue.Status)
	statusStyle := theme.StyleNormal
	if style := issuecore.IssueStatusStyle(theme, status); style != nil {
		statusStyle = *style
	}
	estimate := 0
	if issue.EstimateMinutes != nil {
		estimate = *issue.EstimateMinutes
	}
	head := statusStyle.Render("● " + issuecore.PlainIssueStatus(status))
	if date := issuecore.IssueDateLabel(issue.Status, issue.TodoForDate, issue.CompletedAt, issue.AbandonedAt, state.Settings); date != "" {
		head += theme.StyleDim.Render(" · ") + theme.StyleNormal.Render(date)
	}
	if issue.PinnedDaily {
		head += theme.StyleDim.Render(" · pinned")
	}
	facts := []string{head, theme.StyleNormal.Render(Context(repo, stream))}
	if len(issue.Tags) > 0 {
		facts = append(facts, issuecore.TagChips(theme, issue.Tags, state.Tags, max(8, width-viewsummary.ArcsWidth(2)-3)))
	}
	penalty := issuePenaltyView(theme, state, issue.ID, width)
	if penalty.verdict != "" {
		facts = append(facts, penalty.verdict)
	}
	arcs := []viewsummary.Arc{effortArc(theme, worked, estimate)}
	if penalty.arc != nil {
		arcs = append(arcs, *penalty.arc)
	}
	lines := description(theme, issue.Description, width)
	lines = append(lines, besideArcs(theme, arcs, facts, width)...)
	lines = append(lines, penalty.lines...)
	lines = append(lines, issueSessionLines(theme, state, issue.ID, width)...)
	lines = append(lines, carryLines(theme, state, issue.ID, width)...)
	lines = append(lines, issuecore.NotesLines(theme, state.Settings, issue.Notes, width)...)
	return append(lines, momentumLines(theme, state, momentumhelpers.ForContext(state.MomentumCards, repoID, issue.StreamID), width)...)
}

// IssueWhere is the issue's repo and stream and its time worked, from session metadata once
// loaded (which also knows the repo's ID), else from the issue itself.
func IssueWhere(state types.ContentState, issue api.Issue) (repo, stream string, repoID int64, worked int) {
	repo, stream, worked = issue.RepoName, issue.StreamName, issue.WorkedSeconds
	if meta := sessionmeta.IssueMetaByID(state.AllIssues, issue.ID); meta != nil {
		repo, stream, repoID = meta.RepoName, meta.StreamName, meta.RepoID
		if worked <= 0 {
			worked = meta.WorkedSeconds
		}
	}
	return repo, stream, repoID, worked
}

// IssueInsightLines are the penalty, carry-over and momentum sections of an issue's details,
// for the issue dialog to show the same.
func IssueInsightLines(theme types.Theme, state types.ContentState, issueID, repoID, streamID int64, width int) []string {
	lines := penaltyLines(theme, state, issueID, width)
	lines = append(lines, carryLines(theme, state, issueID, width)...)
	return append(lines, momentumLines(theme, state, momentumhelpers.ForContext(state.MomentumCards, repoID, streamID), width)...)
}

// effortArc fills with worked time against the estimate: cyan on the way, green at the
// estimate and orange past it. Without an estimate it shows the time worked on a dim track.
func effortArc(theme types.Theme, workedSeconds, estimateMinutes int) viewsummary.Arc {
	arc := viewsummary.Arc{Label: "Effort", Value: helperpkg.FormatCompactDurationSeconds(workedSeconds), Color: theme.ColorCyan}
	if estimateMinutes <= 0 {
		return arc
	}
	arc.Label += "/" + helperpkg.FormatCompactDurationMinutes(estimateMinutes)
	arc.Fraction = float64(workedSeconds) / float64(estimateMinutes*60)
	switch {
	case arc.Fraction > 1.05:
		arc.Color = theme.ColorOrange
	case arc.Fraction >= 1:
		arc.Color = theme.ColorGreen
	}
	return arc
}

// Context is "repo › stream", whichever of them is set, or "-".
func Context(repo, stream string) string {
	repo, stream = strings.TrimSpace(repo), strings.TrimSpace(stream)
	switch {
	case repo != "" && stream != "":
		return repo + " › " + stream
	case repo != "":
		return repo
	case stream != "":
		return stream
	default:
		return "-"
	}
}

func besideArcs(theme types.Theme, arcs []viewsummary.Arc, facts []string, width int) []string {
	arcWidth := viewsummary.ArcsWidth(len(arcs))
	return viewsummary.BesideArcs(viewsummary.Arcs(theme, arcs, arcWidth), facts, arcWidth, width)
}

// description leads the details with the item's description and a blank row.
func description(theme types.Theme, text *string, width int) []string {
	if text == nil || strings.TrimSpace(*text) == "" {
		return []string{}
	}
	return append(wrap(theme.StyleNormal, strings.TrimSpace(*text), width), "")
}

func titledText(theme types.Theme, title string, text *string, width int) []string {
	if text == nil || strings.TrimSpace(*text) == "" {
		return nil
	}
	return append(heading(theme, title), wrap(theme.StyleNormal, strings.TrimSpace(*text), width)...)
}

func wrap(style lipgloss.Style, text string, width int) []string {
	lines := strings.Split(lipgloss.NewStyle().Width(max(8, width)).Render(text), "\n")
	for i, line := range lines {
		lines[i] = style.Render(strings.TrimRight(line, " "))
	}
	return lines
}

// heading opens a section of the details: a blank row, then its dim bold title.
func heading(theme types.Theme, title string) []string {
	return []string{"", theme.StyleDim.Bold(true).Render(title)}
}

// momentumLines list momentums one per line: cadence, the current period's progress (time for
// context momentums), the ladder and the streak.
func momentumLines(theme types.Theme, state types.ContentState, cards []sharedtypes.MomentumCard, width int) []string {
	if len(cards) == 0 {
		return nil
	}
	lines := heading(theme, "Momentum")
	for _, row := range viewsummary.MomentumCardRows(theme, cards, viewruntime.ShownDate(state), max(1, width-2)) {
		lines = append(lines, "  "+row)
	}
	return lines
}

// Meter is a bar of cells filled in color for value against scale; any value shows a cell.
func Meter(theme types.Theme, color lipgloss.Color, value, scale, cells int) string {
	filled := 0
	if value > 0 {
		filled = min(cells, max(1, value*cells/max(1, scale)))
	}
	return lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", filled)) +
		theme.StyleDim.Render(strings.Repeat("░", cells-filled))
}

func truncateLines(lines []string, width int) []string {
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lines[i] = viewhelpers.TruncateANSI(line, width)
		}
	}
	return lines
}
