package dialogs

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	sharedtypes "crona.work/shared/types"
	controllerpkg "crona.work/tui/internal/tui/dialogs/controller"
	viewchrome "crona.work/tui/internal/tui/views/chrome"

	"github.com/charmbracelet/lipgloss"
)

// renderHabitStreakDialog draws the momentum wizard: 1 What counts, 2 Goal. ctrl+s on the
// Goal saves; esc leaves from either step, asking first when something changed.
func renderHabitStreakDialog(theme Theme, state controllerpkg.State) string {
	steps := habitStreakWizardSteps()
	active := min(max(0, state.HabitStreakStep-1), len(steps)-1)
	title := "New Momentum"
	submit := "ctrl+s create"
	if state.Kind == "edit_momentum" {
		title, submit = "Edit Momentum", "ctrl+s save"
	}
	body := []string{stepStrip(theme, steps, active), ""}
	d := listDialog{wide: true, title: title, context: stepContext(steps, active), focus: -1}
	if active == 0 {
		d.body, d.focus = momentumTargetsStep(theme, state, body)
		d.primary = "ctrl+s next"
		switch {
		case state.Momentum.KindFocused:
			d.hints = []string{"←/→ switch", "↓ list"}
		case state.Momentum.FilterFocused:
			d.hints = []string{"type to search", "enter list", "esc clear"}
		case momentumContextMode(state):
			d.hints = []string{"type filter", "←/→ cycle", "enter add/remove", "tab next"}
		default:
			d.hints = []string{"space pick", "/ search", "a all shown", "c none shown", "tab match"}
		}
	} else {
		d.body, d.focus = momentumDetailsStep(theme, state, body, true)
		d.primary, d.hints = submit, []string{"tab next", "shift+tab back"}
	}
	if state.Momentum.ConfirmDiscard {
		d.body = append(d.body, "", theme.StyleError.Render("Discard changes?"))
		d.primary, d.hints, d.esc = "y discard", nil, "n keep editing"
	}
	return renderListDialog(theme, state, d)
}

func momentumContextMode(state controllerpkg.State) bool {
	return sharedtypes.NormalizeMomentumTargetKind(state.HabitStreakDraft.TargetKind) ==
		sharedtypes.MomentumTargetKindContext
}

// segmented draws options side by side, the current one highlighted.
func segmented(theme Theme, options []string, current int) string {
	cells := make([]string, 0, len(options))
	for i, option := range options {
		style := theme.StyleDim
		if i == current {
			style = theme.StyleSelected
		}
		cells = append(cells, style.Render(" "+option+" "))
	}
	return strings.Join(cells, " ")
}

func momentumTargetsStep(theme Theme, state controllerpkg.State, body []string) ([]string, int) {
	column := len("Based on")
	full := formInner(state.Width) - column - 3
	focus := -1
	if state.Kind == "create_momentum" {
		if state.Momentum.KindFocused {
			focus = len(body)
		}
		body = append(body,
			Field(theme, "Based on", column, momentumKindSwitch(theme, state), state.Momentum.KindFocused),
			"",
		)
	}
	listFocused := !state.Momentum.KindFocused && !state.Momentum.FilterFocused
	matchFocus := 1
	if momentumContextMode(state) {
		matchFocus = 3
		repo := state.MomentumRepoInput
		stream := state.MomentumStreamInput
		body = append(body,
			Picker(theme, PickerSpec{Label: "Repo", Column: column, Focused: listFocused && state.FocusIdx == 0,
				Input: bareInput(repo, full, "type to filter"), Placeholder: state.RepoSelectorLabel})[0],
			Picker(theme, PickerSpec{Label: "Stream", Column: column, Focused: listFocused && state.FocusIdx == 1,
				Input: bareInput(stream, full, "any stream"), Placeholder: state.StreamSelectorLabel})[0],
			"", theme.StyleDim.Render("Picked contexts"),
		)
		if listFocused && state.FocusIdx < 2 {
			focus = len(body) - 4 + state.FocusIdx
		}
		if len(state.HabitStreakDraft.Contexts) == 0 {
			body = append(body, theme.StyleDim.Render("  None yet: pick a repo, then enter"))
		}
		labels := []string{}
		for _, item := range state.HabitStreakDraft.Contexts {
			labels = append(labels, "[x] "+momentumContextLabel(
				item, state.MomentumRepos, state.MomentumAllIssues, state.MomentumStreams,
			))
		}
		cursor := -1
		if listFocused && state.FocusIdx == 2 {
			cursor = state.HabitStreakCursor
			focus = len(body) + cursor
		}
		body = append(body, selectList(theme, labels, cursor)...)
		for _, warning := range momentumContextRedundancyWarnings(state.HabitStreakDraft.Contexts, state.MomentumRepos, state.MomentumAllIssues, state.MomentumStreams) {
			body = append(body, lipgloss.NewStyle().Foreground(theme.ColorOrange).Render(warning))
		}
	} else {
		body, focus = momentumHabitList(theme, state, body, column, full, focus)
	}
	matchFocused := listFocused && state.FocusIdx == matchFocus
	if matchFocused {
		focus = len(body) + 1
	}
	any := sharedtypes.NormalizeMomentumMatchMode(state.HabitStreakDraft.MatchMode) !=
		sharedtypes.MomentumMatchModeAll
	mode := 0
	if !any {
		mode = 1
	}
	return append(body,
		"",
		Field(theme, "Match", column, segmented(theme, []string{"Any", "All"}, mode)+
			theme.StyleDim.Render("  "+momentumMatchHelp(state.HabitStreakDraft)), matchFocused),
	), focus
}

// momentumMatchHelp says in words what the match mode does with the picks.
func momentumMatchHelp(def sharedtypes.HabitStreakDefinition) string {
	noun := "habit"
	if sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) == sharedtypes.MomentumTargetKindContext {
		noun = "context"
	}
	if sharedtypes.NormalizeMomentumMatchMode(def.MatchMode) == sharedtypes.MomentumMatchModeAll {
		return "each picked " + noun + " must reach the goal"
	}
	return "every picked " + noun + " adds to one goal"
}

func momentumHabitList(theme Theme, state controllerpkg.State, body []string, column, full, focus int) ([]string, int) {
	visible, total, hidden := controllerpkg.MomentumHabitCounts(state)
	count := theme.StyleDim.Render(fmt.Sprintf("  %d of %d", visible, total))
	if state.Momentum.FilterFocused {
		focus = len(body)
	}
	search := bareInput(state.Momentum.Filter, min(full-12, max(17, len(state.Momentum.Filter.Value())+1)), "press / to search")
	body = append(body, Field(theme, "Search", column, search+count, state.Momentum.FilterFocused))
	if total == 0 {
		return append(body, theme.StyleDim.Render("  No habits yet")), focus
	}
	rows := controllerpkg.MomentumHabitRows(state)
	if len(rows) == 0 {
		body = append(body, theme.StyleDim.Render("  No habits match"))
	}
	listFocused := !state.Momentum.KindFocused && !state.Momentum.FilterFocused &&
		state.FocusIdx == 0
	for i, row := range rows {
		cursor := listFocused && i == state.HabitStreakCursor
		if cursor {
			focus = len(body)
		}
		line := ""
		if row.Heading {
			line = lipgloss.NewStyle().Bold(true).Render(row.Label) +
				theme.StyleDim.Render(fmt.Sprintf("  %d/%d", row.Selected, row.Total))
		} else {
			box := "[ ] "
			if slices.Contains(state.HabitStreakDraft.HabitIDs, row.HabitID) {
				box = "[x] "
			}
			line = "  " + box + row.Label
		}
		if cursor {
			body = append(body, theme.StyleCursor.Render(viewchrome.SelectionCursor+" ")+line)
		} else {
			body = append(body, "  "+line)
		}
	}
	if hidden > 0 {
		body = append(body, theme.StyleDim.Render(fmt.Sprintf(
			"  %d picked %s hidden by the search", hidden, plural(hidden, "habit", "habits"),
		)))
	}
	return body, focus
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func momentumDetailsStep(
	theme Theme,
	state controllerpkg.State,
	body []string,
	momentumMode bool,
) ([]string, int) {
	column := len("Description")
	inner := formInner(state.Width)
	full, _ := formColumns(inner, column)
	periodFocus, countFocus := 1, 2
	if momentumMode {
		periodFocus, countFocus = 2, 3
	}
	rows := []FormRow{
		fieldRow(
			0,
			Field(
				theme,
				"Name",
				column,
				bareInput(state.Inputs[0], full, "required"),
				state.FocusIdx == 0,
			),
		),
	}
	if momentumMode {
		rows = append(
			rows,
			textareaRow(
				1,
				Textarea(
					theme,
					"Description",
					column,
					bareTextarea(state),
					state.FocusIdx == 1,
					inner,
				),
			),
		)
	}
	current := sharedtypes.NormalizeHabitStreakPeriod(state.HabitStreakDraft.Period)
	periods := []string{}
	for _, option := range []sharedtypes.HabitStreakPeriod{sharedtypes.HabitStreakPeriodDay, sharedtypes.HabitStreakPeriodWeek, sharedtypes.HabitStreakPeriodMonth} {
		label := habitStreakPeriodLabel(option)
		if option == current {
			periods = append(periods, theme.StyleSelected.Render(" "+label+" "))
		} else {
			periods = append(periods, theme.StyleDim.Render(" "+label+" "))
		}
	}
	rows = append(
		rows,
		fieldRow(periodFocus, Field(theme, "Period", column, strings.Join(periods, " "),
			state.FocusIdx == periodFocus)),
		fieldRow(countFocus, Field(theme, "Required", column,
			bareInput(state.Inputs[1], max(1, len(state.Inputs[1].Value())), "1")+theme.StyleDim.Render("  "+momentumRequiredUnit(state)),
			state.FocusIdx == countFocus)),
	)
	if !momentumContextMode(state) {
		capacity := habitMomentumViewCapacity(state.HabitStreakDraft, state.HabitItems)
		if !capacity.Valid && strings.TrimSpace(capacity.Reason) != "" {
			rows = append(rows, staticRow(theme.StyleError.Render(capacity.Reason)))
		}
	}
	rows = append(rows, staticRow("", theme.StyleDim.Render(momentumGoalSentence(state))))
	lines, focus := formBody(rows, state.FocusIdx)
	if focus >= 0 {
		focus += len(body)
	}
	return append(body, lines...), focus
}

func momentumTargetsSummary(state controllerpkg.State) string {
	return habitStreakTargetSelectionSummary(
		state.HabitStreakDraft,
		state.MomentumRepos,
		state.MomentumAllIssues,
		state.MomentumStreams,
		state.HabitItems,
	)
}

// momentumRequiredUnit names what the required number counts, per period, with the cap
// the picked habits allow.
func momentumRequiredUnit(state controllerpkg.State) string {
	period := habitStreakPeriodUnitLabel(state.HabitStreakDraft.Period)
	if momentumContextMode(state) {
		return "of work a " + period
	}
	n, _ := strconv.Atoi(strings.TrimSpace(state.Inputs[1].Value()))
	unit := plural(n, "completion", "completions") + " a " + period
	if sharedtypes.NormalizeMomentumMatchMode(state.HabitStreakDraft.MatchMode) ==
		sharedtypes.MomentumMatchModeAll {
		unit += " from each"
	}
	capacity := habitMomentumViewCapacity(state.HabitStreakDraft, state.HabitItems)
	if capacity.Valid && capacity.MaxCount > 0 {
		unit += fmt.Sprintf(" (max %d)", capacity.MaxCount)
	}
	return unit
}

// momentumGoalSentence reads the whole momentum back as one line, e.g.
// "Deep work: 3 completions a week from any of Journal, Run".
func momentumGoalSentence(state controllerpkg.State) string {
	def := state.HabitStreakDraft
	name := strings.TrimSpace(state.Inputs[0].Value())
	if name == "" {
		name = "This momentum"
	}
	raw := strings.TrimSpace(state.Inputs[1].Value())
	amount := ""
	if momentumContextMode(state) {
		amount = fallback(raw, "?") + " of work"
	} else {
		n, err := strconv.Atoi(raw)
		if err != nil {
			amount = fallback(raw, "?") + " completions"
		} else {
			amount = fmt.Sprintf("%d %s", n, plural(n, "completion", "completions"))
		}
	}
	scope := "any of"
	if sharedtypes.NormalizeMomentumMatchMode(def.MatchMode) == sharedtypes.MomentumMatchModeAll {
		scope = "each of"
	}
	return fmt.Sprintf("%s: %s a %s from %s %s", name, amount,
		habitStreakPeriodUnitLabel(def.Period), scope, momentumTargetsSummary(state))
}

// momentumKindSwitch draws the kind as a pair of radio options, with the switch key shown
// so it reads as something to change.
func momentumKindSwitch(theme Theme, state controllerpkg.State) string {
	options := []string{"Habit completions", "Time in contexts"}
	current := 0
	if momentumContextMode(state) {
		current = 1
	}
	cells := make([]string, 0, len(options))
	for i, option := range options {
		if i == current {
			style := theme.StyleNormal
			if state.Momentum.KindFocused {
				style = theme.StyleSelected
			}
			cells = append(cells, style.Render("◉ "+option))
			continue
		}
		cells = append(cells, theme.StyleDim.Render("○ "+option))
	}
	hint := "  shift+tab to change"
	if state.Momentum.KindFocused {
		hint = "  ←/→ switch"
	}
	return strings.Join(cells, "   ") + theme.StyleDim.Render(hint)
}
