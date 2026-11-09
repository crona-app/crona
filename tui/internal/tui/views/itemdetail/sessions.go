package itemdetail

import (
	"fmt"
	"slices"
	"strings"
	"time"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/x/ansi"
)

// recentRows caps the sessions or habit days a detail lists.
const recentRows = 5

// barCells is the width of a session's or a habit day's bar.
const barCells = 8

// issueSessionLines list the issue's latest work sessions, newest first: day, start time, a
// bar against the longest listed, and the duration. They show once loaded for this issue.
func issueSessionLines(theme types.Theme, state types.ContentState, issueID int64, width int) []string {
	if state.DetailSessionsIssueID != issueID || issueID == 0 {
		return nil
	}
	sessions := slices.Clone(state.DetailSessions)
	if len(sessions) == 0 {
		return append(heading(theme, "Sessions"), "  "+theme.StyleDim.Render("no work logged yet"))
	}
	slices.SortFunc(sessions, func(a, b api.SessionHistoryEntry) int { return strings.Compare(b.StartTime, a.StartTime) })
	total := 0
	for _, session := range sessions {
		total += sessionSeconds(session)
	}
	lines := heading(theme, "Sessions")
	lines[1] += theme.StyleDim.Render(fmt.Sprintf("  %d · %s total", len(sessions), helperpkg.FormatCompactDurationSeconds(total)))
	shown := sessions[:min(len(sessions), recentRows)]
	longest := 1
	for _, session := range shown {
		longest = max(longest, sessionSeconds(session))
	}
	for _, session := range shown {
		seconds := sessionSeconds(session)
		when := session.StartTime
		if start, err := time.Parse(time.RFC3339, session.StartTime); err == nil {
			when = start.Local().Format("Mon 2 Jan  15:04")
		}
		duration := helperpkg.FormatCompactDurationSeconds(seconds)
		if session.EndTime == nil {
			duration += " · running"
		}
		row := "  " + label(theme, when, 17) + "  " + Meter(theme, theme.ColorCyan, seconds, longest, barCells) + "  " +
			theme.StyleNormal.Render(duration)
		if commit := sessionCommitLine(session); commit != "" {
			row += "  " + theme.StyleDim.Render(commit)
		}
		lines = append(lines, ansi.Truncate(row, width, "…"))
	}
	if more := len(sessions) - len(shown); more > 0 {
		lines = append(lines, "  "+theme.StyleDim.Render(fmt.Sprintf("+%d earlier", more)))
	}
	return lines
}

// sessionCommitLine is the first line of the session's commit message, if it has one.
func sessionCommitLine(session api.SessionHistoryEntry) string {
	message := strings.TrimSpace(session.ParsedNotes[sharedtypes.SessionNoteSectionCommit])
	first, _, _ := strings.Cut(message, "\n")
	return strings.TrimSpace(first)
}

func sessionSeconds(session api.SessionHistoryEntry) int {
	if session.DurationSeconds != nil {
		return *session.DurationSeconds
	}
	return 0
}
