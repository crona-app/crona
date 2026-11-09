package itemdetail

import (
	"fmt"
	"strings"

	"crona.work/tui/internal/api"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	viewsummary "crona.work/tui/internal/tui/views/summary"
	types "crona.work/tui/internal/tui/views/types"

	sharedtypes "crona.work/shared/types"

	"github.com/charmbracelet/lipgloss"
)

// The kernel does all the scoring; these only lay out what it says about an issue's delay.

func score(value float64) string { return fmt.Sprintf("%.2f", value) }

// latePenalty is the entry's penalty when the issue is late (delayed, or pushed more than
// once); onSchedule is set for a live penalty on an issue that is not.
func latePenalty(entry *api.DailyPlanEntry) (penalty *sharedtypes.DailyPlanEntryPenalty, onSchedule bool) {
	if entry == nil || entry.Penalty == nil {
		return nil, false
	}
	if entry.CurrentDelayedDays > 0 || entry.PostponeCount > 1 {
		return entry.Penalty, false
	}
	return nil, !entry.Penalty.Historical
}

// delayParts say how late: the days, the pushes and the away days skipped.
func delayParts(entry *api.DailyPlanEntry, days string) string {
	parts := []string{days}
	if entry.PostponeCount > 1 {
		parts = append(parts, fmt.Sprintf("pushed %d×", entry.PostponeCount))
	}
	if skipped := entry.Penalty.ProtectedDaysSkipped; skipped > 0 {
		parts = append(parts, plural(skipped, "away day")+" skipped")
	}
	return strings.Join(parts, " · ")
}

// clearsText is what finishing today takes off, or "" when it takes nothing.
func clearsText(penalty *sharedtypes.DailyPlanEntryPenalty) string {
	if done := penalty.IfDoneToday; done != nil && done.Accountability < 0 {
		return "clears " + score(-done.Accountability)
	}
	return ""
}

func pushedText(push *sharedtypes.DailyPlanPenaltyProjection) string {
	return fmt.Sprintf("rises to %s (%+.2f)", score(push.FailScore), push.AccountabilityDelta)
}

func label(theme types.Theme, text string, width int) string {
	return theme.StyleDim.Render(viewhelpers.PadCells(text, width))
}

// penaltyLines explain the issue's delay penalty in rows: how late, a bar of the penalty
// against the high-risk limit, its share of today's penalty, and what finishing it today or
// pushing it again would do. A settled penalty keeps only how late and the bar.
func penaltyLines(theme types.Theme, state types.ContentState, issueID int64, width int) []string {
	entry := shownEntry(state, issueID)
	penalty, onSchedule := latePenalty(entry)
	if penalty == nil {
		if onSchedule {
			return []string{"", theme.StyleDim.Render("On schedule · no penalty")}
		}
		return nil
	}
	live := !penalty.Historical
	warn := lipgloss.NewStyle().Foreground(theme.ColorOrange)
	title, body := "Delay penalty", theme.StyleNormal
	if !live {
		title, body = "Delay penalty (final)", theme.StyleDim
	}
	lines := heading(theme, title)
	if penalty.HighRisk && live {
		lines[1] += "  " + warn.Render("▲ high risk")
	}
	lines = append(lines,
		label(theme, "Late", 14)+body.Render(delayParts(entry, plural(entry.CurrentDelayedDays, "day"))),
		label(theme, "Penalty", 14)+penaltyBar(theme, entry.FailScore, penalty.HighRiskThreshold, penalty.HighRisk && live)+
			" "+body.Render(score(entry.FailScore))+theme.StyleDim.Render("  limit "+score(penalty.HighRiskThreshold)),
	)
	if !live {
		return truncateLines(lines, width)
	}
	if share := penalty.Contribution; share != nil {
		lines = append(lines, label(theme, "Share", 14)+body.Render(
			fmt.Sprintf("%.0f%% of today's penalty", share.AccountabilityPct*100)))
	}
	if clears := clearsText(penalty); clears != "" {
		lines = append(lines, label(theme, "Finish today", 14)+lipgloss.NewStyle().Foreground(theme.ColorGreen).Render(clears))
	}
	if push := penalty.IfPostponedOneDay; push != nil {
		text := lipgloss.NewStyle().Foreground(theme.ColorRed).Render(pushedText(push))
		if push.BecomesHighRisk {
			text += "  " + warn.Render("▲ high risk")
		}
		lines = append(lines, label(theme, "Push a day", 14)+text)
	}
	return truncateLines(lines, width)
}

// penaltyView is the issue's delay penalty laid out beside its effort: an arc while it is
// live, a verdict for the fact column, and the rows under the arcs.
type penaltyView struct {
	arc     *viewsummary.Arc
	verdict string
	lines   []string
}

// issuePenaltyView splits a live penalty into an arc of the score against twice the
// high-risk limit, a verdict (risk and share of today), and the delay and what-if rows. A
// settled penalty keeps its plain rows.
func issuePenaltyView(theme types.Theme, state types.ContentState, issueID int64, width int) penaltyView {
	entry := shownEntry(state, issueID)
	penalty, onSchedule := latePenalty(entry)
	switch {
	case onSchedule:
		return penaltyView{verdict: theme.StyleDim.Render("on schedule · no penalty")}
	case penalty == nil:
		return penaltyView{}
	case penalty.Historical:
		return penaltyView{lines: penaltyLines(theme, state, issueID, width)}
	}
	warn := lipgloss.NewStyle().Foreground(theme.ColorOrange)
	limit := penalty.HighRiskThreshold
	color := theme.ColorYellow
	verdict := lipgloss.NewStyle().Foreground(theme.ColorYellow).Render("◆ late")
	if penalty.HighRisk {
		color = theme.ColorOrange
		verdict = warn.Render("▲ high risk")
	}
	if share := penalty.Contribution; share != nil {
		verdict += theme.StyleDim.Render(fmt.Sprintf(" · %.0f%% of today", share.AccountabilityPct*100))
	}
	arcLabel := "Penalty"
	if limit > 0 {
		arcLabel += "/" + strings.TrimRight(strings.TrimRight(score(limit), "0"), ".")
	}
	arc := &viewsummary.Arc{
		Label: arcLabel, Value: score(entry.FailScore),
		Fraction: entry.FailScore / max(limit*2, 0.01), Color: color,
	}
	lines := []string{"", label(theme, "Delay", 11) +
		theme.StyleNormal.Render(delayParts(entry, plural(entry.CurrentDelayedDays, "day")+" late"))}
	if clears := clearsText(penalty); clears != "" {
		lines = append(lines, label(theme, "If today", 11)+lipgloss.NewStyle().Foreground(theme.ColorGreen).Render("✓ "+clears))
	}
	if push := penalty.IfPostponedOneDay; push != nil {
		text := lipgloss.NewStyle().Foreground(theme.ColorRed).Render("▲ " + pushedText(push))
		if push.BecomesHighRisk {
			text += "  " + warn.Render("turns high risk")
		}
		lines = append(lines, label(theme, "If pushed", 11)+text)
	}
	return penaltyView{arc: arc, verdict: verdict, lines: truncateLines(lines, width)}
}

const penaltyBarWidth = 16

// penaltyBar fills to score on a scale of twice the high-risk limit (or the score, if
// higher), with a │ at the limit; it is orange when high risk, yellow otherwise.
func penaltyBar(theme types.Theme, score, limit float64, highRisk bool) string {
	scale := max(limit*2, score, 0.01)
	filled := min(penaltyBarWidth, int(score/scale*penaltyBarWidth+0.5))
	mark := -1
	if limit > 0 {
		mark = min(penaltyBarWidth-1, int(limit/scale*penaltyBarWidth+0.5))
	}
	fill := lipgloss.NewStyle().Foreground(theme.ColorYellow)
	if highRisk {
		fill = lipgloss.NewStyle().Foreground(theme.ColorOrange)
	}
	var out strings.Builder
	for i := range penaltyBarWidth {
		switch {
		case i == mark:
			out.WriteString(theme.StyleNormal.Render("│"))
		case i < filled:
			out.WriteString(fill.Render("█"))
		default:
			out.WriteString(theme.StyleDim.Render("░"))
		}
	}
	return out.String()
}

// PenaltyMarker is the compact high-risk marker for issue rows, unstyled so it takes the
// row's style, or "" when the issue's shown-day entry is not high risk.
func PenaltyMarker(state types.ContentState, issueID int64) string {
	entry := shownEntry(state, issueID)
	if entry == nil || entry.Penalty == nil || entry.Penalty.Historical || !entry.Penalty.HighRisk {
		return ""
	}
	return fmt.Sprintf("▲%.1f", entry.FailScore)
}
