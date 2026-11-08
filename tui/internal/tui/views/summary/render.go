// Package summary renders the Day summary: a day's focus stats, plan accountability,
// check-in and streaks, shown in Daily's detail pane or its Day dialog.
package summary

import (
	uistate "crona.work/tui/internal/tui/state"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"sort"
	"strings"
	"time"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	viewhelpers "crona.work/tui/internal/tui/views/helpers"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// DayLines is the Day summary at width: a row of the day's headline gauges (score, worked,
// issues, habits) with the focus facts under it, the plan's detail, then wellbeing and
// streaks. It waits on the summary snapshot for the shown date.
func DayLines(theme types.Theme, state types.ContentState, width int) []string {
	snapshot := state.SummarySnapshot
	if snapshot == nil || snapshot.Date != state.SummaryDate {
		return []string{theme.StyleDim.Render("Loading summary…")}
	}
	lines := renderVitals(theme, snapshot, width)
	lines = append(lines, "")
	lines = append(lines, renderPlan(theme, state, snapshot.Plan, width)...)
	lines = append(lines, renderPlanIssues(theme, state, width)...)
	lines = append(lines, "")
	lines = append(lines, renderWork(theme, state, width)...)
	lines = append(lines, "")
	lines = append(lines, renderWellbeing(theme, snapshot.CheckIn, snapshot.Rollup, state.Settings, width)...)
	lines = append(lines, lastCheckInLine(theme, snapshot.LastCheckIn, state.Settings, time.Now(), width)...)
	lines = append(lines, "")
	return append(lines, renderMomentum(theme, state, snapshot.Streaks, width)...)
}

// labelWidth is the label column every summary row lines its value up after.
const labelWidth = 10

// Row is one stat: a dim label column, then its value cut to the room left.
func Row(theme types.Theme, label, value string, width int) string {
	return theme.StyleDim.Render(fmt.Sprintf("%-*s", labelWidth, label)) +
		viewhelpers.TruncateANSI(value, max(1, width-labelWidth))
}

// renderVitals is the day at a glance: four gauges in one row, each naming its state under
// it, then the focus facts the gauges can't show on one line.
func renderVitals(theme types.Theme, snapshot *api.SummarySnapshot, width int) []string {
	if snapshot == nil || snapshot.Focus == nil || snapshot.Metrics == nil {
		return []string{theme.StyleDim.Render("Focus stats unavailable.")}
	}
	focus, metrics := snapshot.Focus, snapshot.Metrics
	duration := helperpkg.FormatCompactDurationSeconds
	target := max(1, focus.TargetWorkedSeconds)
	workedFraction := float64(metrics.WorkedSeconds) / float64(target)
	workedColor := theme.ColorCyan
	if workedFraction >= 1 {
		workedColor = theme.ColorGreen
	}
	workedTarget := ""
	if focus.TargetWorkedSeconds > 0 {
		workedTarget = theme.StyleDim.Render("of " + duration(focus.TargetWorkedSeconds))
	}
	gauges := []gauge{
		{
			label: "Score", value: fmt.Sprintf("%d", focus.Score),
			trend:    lipgloss.NewStyle().Foreground(focusColor(theme, focus.Level)).Render(focusLevelWord(focus.Level)),
			fraction: float64(focus.Score) / 100, color: focusColor(theme, focus.Level),
		},
		{
			label: "Worked", value: duration(metrics.WorkedSeconds), trend: workedTarget,
			fraction: workedFraction, color: workedColor,
		},
		issuesGauge(theme, snapshot.Plan, metrics),
		habitsGauge(theme, metrics),
	}
	lines := renderGauges(theme, gauges, width)
	return append(lines, focusFacts(theme, focus, metrics, width)...)
}

// focusFacts is the line under the gauges: time to go, sessions, rest, and the advice behind
// the score when there is any. A narrow pane wraps it between facts, never inside one.
func focusFacts(theme types.Theme, focus *sharedtypes.FocusScoreSummary, metrics *api.DailyMetricsDay, width int) []string {
	duration := helperpkg.FormatCompactDurationSeconds
	sep := theme.StyleDim.Render(" · ")
	var parts []string
	if target := focus.TargetWorkedSeconds; target > 0 {
		switch over := metrics.WorkedSeconds - target; {
		case over > 0:
			parts = append(parts, lipgloss.NewStyle().Foreground(theme.ColorOrange).Render("+"+duration(over)+" over"))
		case over == 0:
			parts = append(parts, lipgloss.NewStyle().Foreground(theme.ColorGreen).Render("target met"))
		default:
			parts = append(parts, theme.StyleDim.Render("to go ")+theme.StyleNormal.Render(duration(-over)))
		}
	}
	sessions := theme.StyleNormal.Render(fmt.Sprint(metrics.SessionCount)) + theme.StyleDim.Render(plural(metrics.SessionCount, " session", " sessions"))
	if metrics.SessionCount > 0 {
		sessions += theme.StyleDim.Render(", avg ") + theme.StyleNormal.Render(duration(metrics.WorkedSeconds/metrics.SessionCount))
	}
	parts = append(parts, sessions)
	rest := theme.StyleDim.Render("rest ") + theme.StyleNormal.Render(duration(metrics.RestSeconds))
	if metrics.RestSeconds == 0 && metrics.WorkedSeconds >= 45*60 {
		rest = theme.StyleDim.Render("rest ") + lipgloss.NewStyle().Foreground(theme.ColorOrange).Render(duration(0))
	}
	parts = append(parts, rest)
	if hint := focusHint(focus.Reason); hint != "" {
		parts = append(parts, theme.StyleDim.Render(hint))
	} else if metrics.RestSeconds == 0 && metrics.WorkedSeconds >= 45*60 {
		parts = append(parts, theme.StyleDim.Render("aim ~1 break an hour"))
	}
	return joinWrapped(parts, sep, width)
}

// joinWrapped joins parts with sep, starting a new line before a part that would not fit.
func joinWrapped(parts []string, sep string, width int) []string {
	var lines []string
	line := ""
	for _, part := range parts {
		switch {
		case line == "":
			line = part
		case ansi.StringWidth(line+sep+part) <= width:
			line += sep + part
		default:
			lines = append(lines, line)
			line = part
		}
	}
	if line != "" {
		lines = append(lines, ansi.Truncate(line, width, "…"))
	}
	for i := range lines {
		lines[i] = ansi.Truncate(lines[i], width, "…")
	}
	return lines
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// issuesGauge counts the day's done issues, the committed plan's when there is one, and
// names the plan's state under it.
func issuesGauge(theme types.Theme, plan *sharedtypes.DailyPlan, metrics *api.DailyMetricsDay) gauge {
	done, total := metrics.CompletedIssues, metrics.TotalIssues
	status := ""
	if plan != nil && strings.TrimSpace(plan.Date) != "" {
		done, total = plan.Summary.CompletedCount, plan.Summary.PlannedCount
		status = planVerdict(theme, plan.Summary)
	}
	fraction := float64(done) / float64(max(1, total))
	return gauge{
		label: "Issues", value: fmt.Sprintf("%d/%d", done, total), trend: status,
		fraction: fraction, color: doneColor(theme, fraction),
	}
}

// habitsGauge counts the due habits done, naming failures first, then what is left.
func habitsGauge(theme types.Theme, metrics *api.DailyMetricsDay) gauge {
	done, failed, due := metrics.HabitCompletedCount, metrics.HabitFailedCount, metrics.HabitDueCount
	g := gauge{label: "Habits", value: fmt.Sprintf("%d/%d", done, due)}
	if due == 0 {
		g.value = "—"
		g.trend = theme.StyleDim.Render("none due")
		return g
	}
	g.fraction = float64(done) / float64(due)
	g.color = doneColor(theme, g.fraction)
	switch left := due - done - failed; {
	case failed > 0:
		g.trend = lipgloss.NewStyle().Foreground(theme.ColorOrange).Render(fmt.Sprintf("%d failed", failed))
	case left > 0:
		g.trend = theme.StyleDim.Render(fmt.Sprintf("%d left", left))
	default:
		g.trend = lipgloss.NewStyle().Foreground(theme.ColorGreen).Render("all done")
	}
	return g
}

// renderPlan is the committed plan's detail, the gauges above carrying its count and state:
// what is slipping, late and resolved, and the open issues by status.
func renderPlan(
	theme types.Theme,
	state types.ContentState,
	plan *sharedtypes.DailyPlan,
	width int,
) []string {
	statusCounts := plannedStatusCounts(state)
	title := theme.StylePaneTitle.Render("Plan")
	// The focused summary offers the plan's list for selection; at rest the pane carries no keys.
	if state.DailyDetailFocus && state.DailyDetailSummary && !state.DailyPlanSelect &&
		len(PlanItems(state.SummarySnapshot, state.AllIssues)) > 0 {
		title += "   " + theme.StyleDim.Render("[P] select")
	}
	lines := []string{title}
	if plan == nil || strings.TrimSpace(plan.Date) == "" {
		lines = append(lines, theme.StyleDim.Render("No committed plan."))
	} else {
		s := plan.Summary
		lines = append(lines,
			Row(theme, "Slipping", riskDots(theme, s), width),
			Row(theme, "Delayed", delayText(theme, s), width),
			Row(theme, "Resolved", resolvedChips(theme, s), width),
		)
	}
	open := map[string]int{}
	for status, n := range statusCounts {
		if status != "done" {
			open[status] = n
		}
	}
	if legend := IssueStatusLegend(theme, open); legend != "" {
		lines = append(lines, Row(theme, "Issues", legend, width))
	}
	return lines
}

// besideArcs puts fact lines to the right of arc gauges, or under them when the pane is narrow.
func besideArcs(arcs, facts []string, arcWidth, width int) []string {
	if width < arcWidth+gaugeGap+28 {
		return append(append([]string{}, arcs...), facts...)
	}
	lines := []string{}
	for i := range max(len(arcs), len(facts)) {
		left := strings.Repeat(" ", arcWidth)
		if i < len(arcs) {
			left = padTo(arcs[i], arcWidth)
		}
		right := ""
		if i < len(facts) {
			right = facts[i]
		}
		lines = append(lines, ansi.Truncate(left+strings.Repeat(" ", gaugeGap+1)+right, width, "…"))
	}
	return lines
}

// focusLevelWord is the score's level in a word, shown under the score arc.
func focusLevelWord(level sharedtypes.FocusScoreLevel) string {
	if level == "" {
		return ""
	}
	return strings.ReplaceAll(string(level), "_", " ")
}

// focusHint is the advice behind the score, a quiet line under the facts. Reasons the facts
// already show (time to go, a balanced day) add nothing, so they get none.
func focusHint(reason sharedtypes.FocusScoreReason) string {
	switch reason {
	case sharedtypes.FocusScoreReasonNoActivity:
		return "no focus logged yet"
	case sharedtypes.FocusScoreReasonNeedsBreaks:
		return "take a break"
	case sharedtypes.FocusScoreReasonOverextended:
		return "past target, ease off"
	}
	return ""
}

func focusColor(theme types.Theme, level sharedtypes.FocusScoreLevel) lipgloss.Color {
	switch level {
	case sharedtypes.FocusScoreLevelStrong:
		return theme.ColorGreen
	case sharedtypes.FocusScoreLevelSteady:
		return theme.ColorYellow
	case sharedtypes.FocusScoreLevelOverextended:
		return theme.ColorOrange
	}
	return theme.ColorRed
}

// planVerdict sums the plan up in a word or two, coloured: at risk once anything is high
// risk, slipping while work is delayed, otherwise on track (or all done once nothing is open).
func planVerdict(theme types.Theme, s sharedtypes.DailyPlanAccountabilitySummary) string {
	word, color := "on track", theme.ColorGreen
	switch {
	case s.HighRiskIssueCount > 0:
		word, color = "at risk", theme.ColorRed
	case s.DelayedIssueCount > 0:
		word, color = "slipping", theme.ColorYellow
	case s.PlannedCount > 0 && s.CompletedCount == s.PlannedCount:
		word = "all done"
	}
	return lipgloss.NewStyle().Foreground(color).Render(word)
}

// riskDots draws one dot per planned task, high-risk ones filled red first.
func riskDots(theme types.Theme, s sharedtypes.DailyPlanAccountabilitySummary) string {
	if s.HighRiskIssueCount == 0 {
		return theme.StyleDim.Render("none high risk")
	}
	total := max(s.PlannedCount, s.HighRiskIssueCount)
	shown := min(total, 12)
	hot := min(shown, s.HighRiskIssueCount)
	dots := lipgloss.NewStyle().Foreground(theme.ColorRed).Render(strings.Repeat("●", hot)) +
		theme.StyleDim.Render(strings.Repeat("○", shown-hot))
	return dots + "  " + fmt.Sprintf("%d high risk", s.HighRiskIssueCount)
}

func delayText(theme types.Theme, s sharedtypes.DailyPlanAccountabilitySummary) string {
	if s.DelayedIssueCount == 0 {
		return theme.StyleDim.Render("nothing late")
	}
	color := theme.ColorYellow
	if s.MaxDelayDays >= 3 {
		color = theme.ColorOrange
	}
	return lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%d", s.DelayedIssueCount)) +
		theme.StyleDim.Render(fmt.Sprintf(" · avg %.1fd · max %dd", s.AvgDelayDays, s.MaxDelayDays))
}

// resolvedChips lists only the outcomes that happened, or says there were no misses.
func resolvedChips(theme types.Theme, s sharedtypes.DailyPlanAccountabilitySummary) string {
	var chips []string
	for _, c := range []struct {
		glyph, label string
		count        int
		color        lipgloss.Color
	}{
		{"✗", "failed", s.FailedCount, theme.ColorRed},
		{"⊘", "abandoned", s.AbandonedCount, theme.ColorOrange},
		{"↷", "carried", s.CarriedCount, theme.ColorLavender},
		{"↺", "to roll back", s.PendingRollbackCount, theme.ColorYellow},
	} {
		if c.count > 0 {
			chips = append(chips, outcome(theme, c.glyph, c.count, c.label, c.color))
		}
	}
	if len(chips) == 0 {
		return theme.StyleDim.Render("no misses yet")
	}
	return strings.Join(chips, theme.StyleDim.Render("  "))
}

// doneColor is green once every planned task is done, yellow past half and lavender before.
func doneColor(theme types.Theme, fraction float64) lipgloss.Color {
	switch {
	case fraction >= 1:
		return theme.ColorGreen
	case fraction >= 0.5:
		return theme.ColorYellow
	}
	return theme.ColorLavender
}

func outcome(theme types.Theme, glyph string, count int, label string, color lipgloss.Color) string {
	text := fmt.Sprintf("%s %d %s", glyph, count, label)
	if count == 0 {
		return theme.StyleDim.Render(text)
	}
	return lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("%s %d", glyph, count)) +
		theme.StyleNormal.Render(" "+label)
}

// PlanItem is one open entry of the day's committed plan, with its issue's title.
type PlanItem struct {
	Entry sharedtypes.DailyPlanEntry
	Title string
}

// PlanItems are the issues adding to the summarised day's accountability penalty, from any
// day's plan, riskiest first: the ones Slipping and Delayed count. The Plan section lists them
// and its selection mode walks them, so both read this one order. A daemon that predates the
// contributing list leaves only the day's own penalised entries.
func PlanItems(snapshot *api.SummarySnapshot, all []api.IssueWithMeta) []PlanItem {
	if snapshot == nil || snapshot.Plan == nil {
		return nil
	}
	titles, closed := map[int64]string{}, map[int64]bool{}
	for _, issue := range all {
		titles[issue.ID] = issue.Title
		closed[issue.ID] = issue.Status == "done" || issue.Status == "abandoned"
	}
	if snapshot.Issues != nil {
		for _, issue := range snapshot.Issues.Issues {
			titles[issue.ID] = issue.Title
		}
	}
	entries := snapshot.Plan.Contributing
	if entries == nil {
		for _, entry := range snapshot.Plan.Entries {
			if p := entry.Penalty; p != nil && !p.Historical && entry.FailScore > 0 {
				entries = append(entries, entry)
			}
		}
	}
	items := []PlanItem{}
	seen := map[int64]bool{}
	for _, entry := range entries {
		if planEntryResolved(entry.Status) || closed[entry.IssueID] || seen[entry.IssueID] {
			continue
		}
		seen[entry.IssueID] = true
		title := titles[entry.IssueID]
		if title == "" {
			title = fmt.Sprintf("#%d", entry.IssueID)
		}
		items = append(items, PlanItem{Entry: entry, Title: title})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Entry.FailScore > items[j].Entry.FailScore })
	return items
}

// planEntryResolved reports whether a plan entry has left the open plan.
func planEntryResolved(status sharedtypes.DailyPlanEntryStatus) bool {
	switch status {
	case sharedtypes.DailyPlanEntryStatusCompleted, sharedtypes.DailyPlanEntryStatusFailed,
		sharedtypes.DailyPlanEntryStatusAbandoned, sharedtypes.DailyPlanEntryStatusCanceled,
		sharedtypes.DailyPlanEntryStatusCarried:
		return true
	}
	return false
}

// renderPlanIssues lists the issues behind the penalty under the Plan section. At rest it
// shows the top three; with the list selected (P on the focused summary) it shows them all, the
// cursor's row followed by the keys it takes.
func renderPlanIssues(theme types.Theme, state types.ContentState, width int) []string {
	items := PlanItems(state.SummarySnapshot, state.AllIssues)
	selecting := state.DailyPlanSelect
	if !selecting {
		items = items[:min(3, len(items))]
	}
	if len(items) == 0 {
		return nil
	}
	cursor := min(max(0, state.DailyPlanCursor), len(items)-1)
	label := "Top"
	if selecting {
		label = "Open"
	}
	lines := []string{}
	for i, item := range items {
		row := planRow(theme, item)
		if selecting {
			pointer := "  "
			if i == cursor {
				pointer = theme.StyleCursor.Render("▶ ")
			}
			row = pointer + row
		}
		rowLabel := ""
		if i == 0 {
			rowLabel = label
		}
		lines = append(lines, Row(theme, rowLabel, row, width))
		if selecting && i == cursor {
			keys := theme.StyleDim.Render("[s] status  [d] due")
			lines = append(lines, Row(theme, "", "    "+keys, width))
		}
	}
	return lines
}

// planRow is one plan issue: a red mark when high risk, its title, and why it matters (how
// late, its share of the accountability penalty).
func planRow(theme types.Theme, item PlanItem) string {
	entry := item.Entry
	mark, color := "  ", theme.ColorYellow
	if entry.Penalty != nil && entry.Penalty.HighRisk {
		mark = lipgloss.NewStyle().Foreground(theme.ColorRed).Render("▲ ")
		color = theme.ColorOrange
	}
	why := "due today"
	if entry.CurrentDelayedDays > 0 {
		why = fmt.Sprintf("late %dd", entry.CurrentDelayedDays)
	}
	if entry.Penalty != nil && entry.Penalty.Contribution != nil {
		why += fmt.Sprintf(" · %.0f%%", entry.Penalty.Contribution.AccountabilityPct*100)
	}
	return mark + item.Title + theme.StyleDim.Render(" · ") + lipgloss.NewStyle().Foreground(color).Render(why)
}

func renderWellbeing(
	theme types.Theme,
	checkIn *sharedtypes.DailyCheckIn,
	rollup *sharedtypes.MetricsRollup,
	settings *sharedtypes.CoreSettings,
	width int,
) []string {
	lines := []string{theme.StylePaneTitle.Render("Wellbeing")}
	if checkIn == nil || strings.TrimSpace(checkIn.Date) == "" {
		lines = append(lines, theme.StyleDim.Render("No check-in recorded."))
		if rollup != nil {
			if rollup.AverageMood != nil {
				lines = append(
					lines,
					Row(theme, "7d mood", fmt.Sprintf("%.1f/5", *rollup.AverageMood), width),
				)
			}
			if rollup.AverageEnergy != nil {
				lines = append(
					lines,
					Row(theme, "7d energy", fmt.Sprintf("%.1f/5", *rollup.AverageEnergy), width),
				)
			}
			if rollup.AverageSleepHours != nil {
				lines = append(
					lines,
					Row(
						theme,
						"7d sleep",
						helperpkg.FormatCompactDurationHours(*rollup.AverageSleepHours),
						width,
					),
				)
			}
		}
		return lines
	}
	var avgMood, avgEnergy, avgSleep, avgScreen *float64
	if rollup != nil {
		avgMood, avgEnergy = rollup.AverageMood, rollup.AverageEnergy
		avgSleep, avgScreen = rollup.AverageSleepHours, rollup.AverageScreenTimeMins
	}
	oneDecimal := func(v float64) string { return fmt.Sprintf("%.1f", v) }
	gauges := []gauge{
		{
			label: "Mood", value: fmt.Sprintf("%d/5", checkIn.Mood), fraction: float64(checkIn.Mood) / 5,
			color: scaleColor(theme, checkIn.Mood),
			trend: trend(theme, float64(checkIn.Mood), avgMood, 0.3, true, oneDecimal),
		},
		{
			label: "Energy", value: fmt.Sprintf("%d/5", checkIn.Energy), fraction: float64(checkIn.Energy) / 5,
			color: scaleColor(theme, checkIn.Energy),
			trend: trend(theme, float64(checkIn.Energy), avgEnergy, 0.3, true, oneDecimal),
		},
	}
	sleepTarget, screenLimit := wellbeingTargets(settings)
	if checkIn.SleepHours != nil {
		minutes := *checkIn.SleepHours * 60
		gauges = append(gauges, gauge{
			label: "Sleep", value: helperpkg.FormatCompactDurationHours(*checkIn.SleepHours),
			fraction: minutes / float64(sleepTarget),
			color:    targetColor(theme, minutes/float64(sleepTarget)),
			trend:    trend(theme, *checkIn.SleepHours, avgSleep, 0.25, true, helperpkg.FormatCompactDurationHours),
		})
	}
	if checkIn.ScreenTimeMinutes != nil {
		// The screen gauge is the limit left, so like the others a fuller arc is better.
		left := screenLimit - *checkIn.ScreenTimeMinutes
		value := helperpkg.FormatCompactDurationMinutes(max(0, left))
		if left < 0 {
			value = "+" + helperpkg.FormatCompactDurationMinutes(-left)
		}
		var avgLeft *float64
		if avgScreen != nil {
			v := float64(screenLimit) - *avgScreen
			avgLeft = &v
		}
		gauges = append(gauges, gauge{
			label: "Screen left", value: value,
			fraction: float64(left) / float64(screenLimit),
			color:    limitColor(theme, left, screenLimit),
			trend: trend(theme, float64(left), avgLeft, 15, true, func(v float64) string {
				if v < 0 {
					return "+" + helperpkg.FormatCompactDurationMinutes(int(-v))
				}
				return helperpkg.FormatCompactDurationMinutes(int(v))
			}),
		})
	}
	lines = append(lines, renderGauges(theme, gauges, width)...)
	if checkIn.Notes != nil && strings.TrimSpace(*checkIn.Notes) != "" {
		lines = append(lines, "")
		label, quote := "Notes", "❝ "
		for _, line := range strings.Split(ansi.Wordwrap(strings.TrimSpace(*checkIn.Notes), max(1, width-labelWidth-2), ""), "\n") {
			lines = append(lines, Row(theme, label, theme.StyleDim.Render(quote)+theme.StyleDim.Italic(true).Render(line), width))
			label, quote = "", "  "
		}
	}
	return lines
}

// DayLine is the day summary in one line for narrow screens, where the detail pane that
// holds it is hidden: focus, worked against target, rest, sessions and accountability. Parts
// drop from the right to fit; it is empty until the summary has loaded.
func DayLine(theme types.Theme, state types.ContentState, width int) string {
	snapshot := state.SummarySnapshot
	if snapshot == nil || snapshot.Date != state.SummaryDate || snapshot.Focus == nil || snapshot.Metrics == nil {
		return ""
	}
	duration := helperpkg.FormatCompactDurationSeconds
	focus, metrics := snapshot.Focus, snapshot.Metrics
	part := func(label, value string) string {
		return theme.StyleDim.Render(label+" ") + value
	}
	level := strings.ReplaceAll(string(focus.Level), "_", " ")
	parts := []string{
		part("Focus", theme.StyleHeader.Render(fmt.Sprint(focus.Score))+" "+theme.StyleNormal.Render(level)),
		part("Worked", duration(metrics.WorkedSeconds)+theme.StyleDim.Render("/"+duration(focus.TargetWorkedSeconds))),
		part("Rest", duration(metrics.RestSeconds)),
		part("Sessions", fmt.Sprint(metrics.SessionCount)),
	}
	if plan := snapshot.Plan; plan != nil && strings.TrimSpace(plan.Date) != "" {
		parts = append(parts, part("Score", fmt.Sprintf("%.0f%%", plan.Summary.AccountabilityScore)))
	}
	sep := theme.StyleDim.Render("  ·  ")
	for len(parts) > 1 && ansi.StringWidth(" "+strings.Join(parts, sep)) > width {
		parts = parts[:len(parts)-1]
	}
	return viewhelpers.TruncateANSI(" "+strings.Join(parts, sep), width)
}

// plannedStatusCounts counts the shown day's planned tasks by status, for the status bar.
// plannedStatusCounts counts the summarised day's issues by status: every issue in its
// committed plan (carried-in ones included, as Slipping and Delayed count them), or without a
// plan the issues due that day.
func plannedStatusCounts(state types.ContentState) map[string]int {
	counts := map[string]int{}
	if snapshot := state.SummarySnapshot; snapshot != nil && snapshot.Plan != nil && len(snapshot.Plan.Entries) > 0 {
		statuses := make(map[int64]string, len(state.AllIssues))
		for _, issue := range state.AllIssues {
			statuses[issue.ID] = string(issue.Status)
		}
		seen := map[int64]bool{}
		for _, entry := range snapshot.Plan.Entries {
			status, ok := statuses[entry.IssueID]
			if !ok || seen[entry.IssueID] {
				continue
			}
			seen[entry.IssueID] = true
			counts[status]++
		}
		return counts
	}
	for _, issue := range state.DailyPlannedIssues {
		counts[string(issue.Status)]++
	}
	return counts
}

// wellbeingTargets are the sleep target and screen-time limit from settings, in minutes,
// falling back to the defaults.
func wellbeingTargets(settings *sharedtypes.CoreSettings) (sleep, screen int) {
	sleep, screen = sharedtypes.DefaultSleepTargetMinutes, sharedtypes.DefaultScreenTimeLimitMinutes
	if settings != nil && settings.SleepTargetMinutes > 0 {
		sleep = settings.SleepTargetMinutes
	}
	if settings != nil && settings.ScreenTimeLimitMins > 0 {
		screen = settings.ScreenTimeLimitMins
	}
	return sleep, screen
}

// targetColor colours progress toward a target: green from the target, yellow within an
// eighth below it, red further off.
func targetColor(theme types.Theme, fraction float64) lipgloss.Color {
	switch {
	case fraction >= 1:
		return theme.ColorGreen
	case fraction >= 0.875:
		return theme.ColorYellow
	}
	return theme.ColorRed
}

// scaleColor colours a 1-5 check-in value: red low, yellow middling, green high.
func scaleColor(theme types.Theme, value int) lipgloss.Color {
	switch {
	case value <= 2:
		return theme.ColorRed
	case value == 3:
		return theme.ColorYellow
	}
	return theme.ColorGreen
}

// trend compares today with the 7-day average: an arrow for the direction, green when it
// moved the good way, red when the bad way, then the average, dim. Within tolerance it shows
// only the average.
func trend(theme types.Theme, today float64, average *float64, tolerance float64, higherIsBetter bool, format func(float64) string) string {
	if average == nil {
		return ""
	}
	avg := theme.StyleDim.Render("7d " + format(*average))
	diff := today - *average
	if diff > -tolerance && diff < tolerance {
		return theme.StyleDim.Render("=") + " " + avg
	}
	arrow, good := "▲", diff > 0
	if diff < 0 {
		arrow = "▼"
	}
	if !higherIsBetter {
		good = !good
	}
	color := theme.ColorRed
	if good {
		color = theme.ColorGreen
	}
	return lipgloss.NewStyle().Foreground(color).Render(arrow) + " " + avg
}

// limitColor colours what is left of a limit: green with a quarter or more left, yellow
// below that, red once over.
func limitColor(theme types.Theme, left, limit int) lipgloss.Color {
	switch {
	case left < 0:
		return theme.ColorRed
	case left*4 < limit:
		return theme.ColorYellow
	}
	return theme.ColorGreen
}

// lastCheckInLine names the latest check-in and how long ago it was; empty when there is none.
func lastCheckInLine(theme types.Theme, date string, settings *sharedtypes.CoreSettings, now time.Time, width int) []string {
	day, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return nil
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	ago := "today"
	switch days := int(today.Sub(day).Hours() / 24); {
	case days == 1:
		ago = "yesterday"
	case days > 1:
		ago = fmt.Sprintf("%d days ago", days)
	case days < 0:
		ago = "ahead"
	}
	return []string{"", Row(theme, "Check-in", day.Format("Mon ")+helperpkg.FormatDisplayDate(date, settings)+theme.StyleDim.Render(" ("+ago+")"), width)}
}

var workGroupNames = map[string][2]string{
	"repo": {"Repo", "repos"}, "stream": {"Stream", "streams"}, "issue": {"Issue", "issues"},
	"tag": {"Tag", "tags"}, "segment_type": {"Segment", "segment types"},
}

// renderWork is where the day's worked time went, grouped by repo, stream, issue, tag or
// segment (b cycles). With its rows selected (B) enter opens a repo's streams or a stream's
// issues and esc steps back; at rest it shows the top five.
func renderWork(theme types.Theme, state types.ContentState, width int) []string {
	group := uistate.WorkLevelGroup(state.DailyWorkGroup, state.DailyWorkPath)
	selecting := state.DailyWorkSelect
	title := theme.StylePaneTitle.Render("Work")
	if state.DailyDetailFocus && state.DailyDetailSummary && !selecting && !state.DailyPlanSelect {
		keys := "[b] group"
		if uistate.WorkDrillable(group) {
			keys += "  [B] select"
		}
		title += "   " + theme.StyleDim.Render(keys)
	}
	lines := []string{title, workHeading(theme, state, group, width)}
	summary := state.WorkBreakdown
	if summary == nil {
		return append(lines, theme.StyleDim.Render("Loading"))
	}
	// Empty rows (a segment type never used) say nothing.
	rows := make([]sharedtypes.TimeDistributionRow, 0, len(summary.Rows))
	for _, row := range summary.Rows {
		if row.WorkedSeconds > 0 {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		return append(lines, theme.StyleDim.Render("No work logged."))
	}
	cursor := min(max(0, state.DailyWorkCursor), len(rows)-1)
	start, end := 0, min(5, len(rows))
	if selecting {
		start, end = viewchrome.ListWindow(cursor, len(rows), 8)
	}
	labelWidth := min(24, max(8, width/3))
	barWidth := max(4, min(14, width-labelWidth-20))
	for i := start; i < end; i++ {
		row := rows[i]
		filled := int(row.Percent/100*float64(barWidth) + 0.5)
		filled = min(barWidth, max(0, filled))
		if row.WorkedSeconds > 0 && filled == 0 {
			filled = 1
		}
		bar := lipgloss.NewStyle().Foreground(theme.ColorCyan).Render(strings.Repeat("█", filled)) +
			theme.StyleDim.Render(strings.Repeat("░", barWidth-filled))
		label := viewchrome.LabelCell(row.Label, labelWidth)
		pointer := "  "
		if selecting && i == cursor {
			pointer = theme.StyleCursor.Render("▶ ")
		}
		line := pointer + label + " " + bar + " " +
			fmt.Sprintf("%6s", helperpkg.FormatCompactDurationSeconds(row.WorkedSeconds)) +
			theme.StyleDim.Render(fmt.Sprintf("  %3.0f%%", row.Percent))
		lines = append(lines, viewhelpers.TruncateANSI(line, width))
		if selecting && i == cursor {
			keys := "[esc] back"
			if uistate.WorkDrillable(group) {
				keys = "[enter] open  " + keys
			}
			lines = append(lines, "    "+theme.StyleDim.Render(keys))
		}
	}
	names := workGroupNames[group]
	noun := names[1]
	if len(rows) == 1 {
		noun = strings.ToLower(names[0])
	}
	foot := fmt.Sprintf("%d %s · %s", len(rows), noun, helperpkg.FormatCompactDurationSeconds(summary.TotalSeconds))
	if !selecting && len(rows) > end {
		foot = fmt.Sprintf("+%d more · ", len(rows)-end) + foot
	}
	if group == "tag" {
		foot += " · tags overlap"
	}
	return append(lines, "  "+theme.StyleDim.Render(foot))
}

// workHeading names what the rows are: the groupings with the current one lit, or the drill
// path down to the level shown.
func workHeading(theme types.Theme, state types.ContentState, group string, width int) string {
	if len(state.DailyWorkPath) > 0 {
		parts := []string{theme.StyleDim.Render(workGroupNames[state.DailyWorkPath[0].Group][0])}
		for _, crumb := range state.DailyWorkPath {
			parts = append(parts, crumb.Label)
		}
		parts = append(parts, theme.StyleHeader.Render(cases.Title(language.Und).String(workGroupNames[group][1])))
		return viewhelpers.TruncateANSI("  "+strings.Join(parts, theme.StyleDim.Render(" › ")), width)
	}
	chips := make([]string, 0, len(uistate.WorkGroups))
	for _, candidate := range uistate.WorkGroups {
		name := workGroupNames[candidate][0]
		if candidate == group {
			chips = append(chips, theme.StyleHeader.Render(name))
		} else {
			chips = append(chips, theme.StyleDim.Render(name))
		}
	}
	return viewhelpers.TruncateANSI("  "+strings.Join(chips, theme.StyleDim.Render(" · ")), width)
}
