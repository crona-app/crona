package controller

import (
	"encoding/json"
	"slices"
	"strings"

	sharedtypes "crona.work/shared/types"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// The momentum wizard has two steps: 1 "What counts" (kind, targets, match) and 2 "Goal"
// (name, description, period, required), which saves.

func updateHabitStreaks(state State, msg tea.KeyMsg) (State, *Action, string) {
	state = syncMomentumStepFocus(state)
	if state.Momentum.ConfirmDiscard {
		switch msg.String() {
		case "y", "enter":
			return Close(state), nil, ""
		case "n", "esc":
			state.Momentum.ConfirmDiscard = false
		}
		return state, nil, ""
	}
	if state.HabitStreakStep == 1 && state.Momentum.FilterFocused {
		return updateMomentumFilter(state, msg)
	}
	if msg.String() == "esc" {
		if state.HabitStreakStep == 1 && !momentumContextKind(state) &&
			state.Momentum.Filter.Value() != "" {
			state.Momentum.Filter.SetValue("")
			state.HabitStreakCursor = 0
			return state, nil, ""
		}
		if MomentumDirty(state) {
			state.Momentum.ConfirmDiscard = true
			return state, nil, ""
		}
		return Close(state), nil, ""
	}
	if state.HabitStreakStep == 2 {
		return updateHabitStreakDetails(state, msg)
	}
	if state.Momentum.KindFocused {
		return updateMomentumKindRow(state, msg)
	}
	if momentumContextKind(state) {
		return updateMomentumContextTargets(state, msg)
	}
	return updateMomentumHabitTargets(state, msg)
}

func openHabitStreakEditor(state State, idx int, def sharedtypes.HabitStreakDefinition) State {
	def = sharedtypes.NormalizeHabitStreakDefinition(def)
	name := textinput.New()
	name.Placeholder = "Health streak"
	name.SetValue(def.Name)
	name.CharLimit = 80
	name.Width = 36
	name.Focus()
	count := textinput.New()
	count.CharLimit = 24
	count.Width = 24
	state.Inputs = []textinput.Model{name, count}
	state = habitStreakSetDetailFocus(state, habitStreakDetailRowName)
	state.HabitStreakEditIdx = idx
	state.HabitStreakDraft = def
	state = syncHabitStreakThresholdInput(state)
	state.HabitStreakStep = 1
	state.HabitStreakCursor = 0
	state.FocusIdx = 0
	state.ErrorMessage = ""
	filter := textinput.New()
	filter.Placeholder = "name, repo, stream or tag"
	filter.CharLimit = 80
	filter.Width = 36
	state.Momentum = MomentumWizard{Filter: filter}
	return state
}

// MomentumDirty reports whether the wizard holds changes since it opened.
func MomentumDirty(state State) bool {
	return momentumFingerprint(state) != state.Momentum.Initial
}

func momentumFingerprint(state State) string {
	draft, _ := json.Marshal(state.HabitStreakDraft)
	parts := []string{string(draft), state.Description.Value()}
	for _, input := range state.Inputs {
		parts = append(parts, input.Value())
	}
	return strings.Join(parts, "\x00")
}

func momentumContextKind(state State) bool {
	return sharedtypes.NormalizeMomentumTargetKind(state.HabitStreakDraft.TargetKind) ==
		sharedtypes.MomentumTargetKindContext
}

func focusMomentumKindRow(state State) State {
	if state.Kind != "create_momentum" {
		return state
	}
	state.Momentum.KindFocused = true
	state.Momentum.FilterFocused = false
	return syncMomentumStepFocus(state)
}

func leaveMomentumKindRow(state State) State {
	state.Momentum.KindFocused = false
	state.FocusIdx = 0
	state.HabitStreakCursor = 0
	return syncMomentumStepFocus(state)
}

// updateMomentumKindRow drives the "Counts ‹ Habits › Contexts" row. Switching keeps each
// kind's picks; only the required amount resets to the kind's default.
func updateMomentumKindRow(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch dialogActionForKey(state, msg.String()) {
	case dialogActionMoveLeft, dialogActionMoveRight, dialogActionToggle, dialogActionActivate:
		next := sharedtypes.MomentumTargetKindContext
		if momentumContextKind(state) {
			next = sharedtypes.MomentumTargetKindHabit
		}
		state.HabitStreakDraft.TargetKind = next
		state.HabitStreakDraft.RequiredCount = habitStreakDefaultRequiredCount(next)
		state = syncHabitStreakThresholdInput(state)
	case dialogActionMoveDown, dialogActionFocusNext:
		return leaveMomentumKindRow(clearDialogError(state)), nil, ""
	case dialogActionPrimary:
		return momentumGoToGoal(state), nil, ""
	}
	return clearDialogError(state), nil, ""
}

func updateMomentumFilter(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		if state.Momentum.Filter.Value() != "" {
			state.Momentum.Filter.SetValue("")
			state.HabitStreakCursor = 0
			return state, nil, ""
		}
		state.Momentum.FilterFocused = false
		return syncMomentumStepFocus(state), nil, ""
	case "enter", "down", "up", "tab", "shift+tab":
		state.Momentum.FilterFocused = false
		state.FocusIdx = int(habitTargetFocusList)
		state.HabitStreakCursor = 0
		return syncMomentumStepFocus(state), nil, ""
	case "ctrl+s":
		state.Momentum.FilterFocused = false
		return momentumGoToGoal(state), nil, ""
	}
	before := state.Momentum.Filter.Value()
	state.Momentum.Filter, _ = state.Momentum.Filter.Update(msg)
	if state.Momentum.Filter.Value() != before {
		state.HabitStreakCursor = 0
	}
	return clearDialogError(state), nil, ""
}

func updateMomentumHabitTargets(state State, msg tea.KeyMsg) (State, *Action, string) {
	if state.FocusIdx == int(habitTargetFocusMatchMode) {
		switch dialogActionForKey(state, msg.String()) {
		case dialogActionMoveLeft, dialogActionMoveUp:
			if msg.String() == "up" {
				state.FocusIdx = int(habitTargetFocusList)
				return clearDialogError(state), nil, ""
			}
			state.HabitStreakDraft.MatchMode = nextMomentumMatchMode(state.HabitStreakDraft.MatchMode, -1)
		case dialogActionMoveRight, dialogActionToggle, dialogActionActivate:
			state.HabitStreakDraft.MatchMode = nextMomentumMatchMode(state.HabitStreakDraft.MatchMode, 1)
		case dialogActionFocusPrev:
			state.FocusIdx = int(habitTargetFocusList)
		case dialogActionFocusNext, dialogActionPrimary:
			return momentumGoToGoal(state), nil, ""
		}
		return clearDialogError(state), nil, ""
	}
	rows := MomentumHabitRows(state)
	switch dialogActionForKey(state, msg.String()) {
	case dialogActionMoveDown:
		state.HabitStreakCursor = ShiftSelection(state.HabitStreakCursor, len(rows), 1)
	case dialogActionMoveUp:
		if state.HabitStreakCursor <= 0 {
			return focusMomentumKindRow(clearDialogError(state)), nil, ""
		}
		state.HabitStreakCursor = ShiftSelection(state.HabitStreakCursor, len(rows), -1)
	case dialogActionToggle, dialogActionActivate:
		state = toggleMomentumHabitRow(state, rows)
	case dialogActionFocusNext:
		state.FocusIdx = int(habitTargetFocusMatchMode)
	case dialogActionFocusPrev:
		return focusMomentumKindRow(clearDialogError(state)), nil, ""
	case dialogActionPrimary:
		return momentumGoToGoal(state), nil, ""
	default:
		switch msg.String() {
		case "/":
			state.Momentum.FilterFocused = true
			return syncMomentumStepFocus(clearDialogError(state)), nil, ""
		case "a":
			for _, row := range rows {
				if !row.Heading && !slices.Contains(state.HabitStreakDraft.HabitIDs, row.HabitID) {
					state.HabitStreakDraft.HabitIDs = append(state.HabitStreakDraft.HabitIDs, row.HabitID)
				}
			}
		case "c":
			visible := momentumVisibleHabitIDs(rows)
			state.HabitStreakDraft.HabitIDs = slices.DeleteFunc(
				slices.Clone(state.HabitStreakDraft.HabitIDs),
				func(id int64) bool { return slices.Contains(visible, id) },
			)
		}
	}
	return clearDialogError(state), nil, ""
}

// toggleMomentumHabitRow flips a habit, or a whole group from its heading: a group with
// every visible habit picked clears, any other group fills.
func toggleMomentumHabitRow(state State, rows []MomentumHabitRow) State {
	if state.HabitStreakCursor < 0 || state.HabitStreakCursor >= len(rows) {
		return state
	}
	row := rows[state.HabitStreakCursor]
	if !row.Heading {
		state.HabitStreakDraft.HabitIDs = toggleHabitMembership(
			slices.Clone(state.HabitStreakDraft.HabitIDs), row.HabitID,
		)
		return state
	}
	group := []int64{}
	for _, candidate := range rows {
		if !candidate.Heading && candidate.Group == row.Group {
			group = append(group, candidate.HabitID)
		}
	}
	ids := slices.Clone(state.HabitStreakDraft.HabitIDs)
	if row.Selected == row.Total {
		ids = slices.DeleteFunc(ids, func(id int64) bool { return slices.Contains(group, id) })
	} else {
		for _, id := range group {
			if !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
	}
	state.HabitStreakDraft.HabitIDs = ids
	return state
}

func momentumGoToGoal(state State) State {
	if err := momentumTargetsError(state.HabitStreakDraft); err != "" {
		state.ErrorMessage = err
		return state
	}
	state.Momentum.KindFocused = false
	state.Momentum.FilterFocused = false
	state.HabitStreakStep = 2
	state.FocusIdx = 0
	return syncMomentumStepFocus(clearDialogError(state))
}

func momentumTargetsError(def sharedtypes.HabitStreakDefinition) string {
	if sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) == sharedtypes.MomentumTargetKindContext {
		if len(def.Contexts) == 0 {
			return "Pick at least one context"
		}
		return ""
	}
	if len(def.HabitIDs) == 0 {
		return "Pick at least one habit"
	}
	return ""
}

// momentumDraftForSave drops the picks of the kind not chosen; the wizard keeps both while
// open so switching kinds loses nothing.
func momentumDraftForSave(def sharedtypes.HabitStreakDefinition) sharedtypes.HabitStreakDefinition {
	if sharedtypes.NormalizeMomentumTargetKind(def.TargetKind) == sharedtypes.MomentumTargetKindContext {
		def.HabitIDs = nil
	} else {
		def.Contexts = nil
	}
	return def
}

// MomentumHabitRow is one line of the grouped habit list: a context heading or a habit.
type MomentumHabitRow struct {
	Heading  bool
	Group    string
	Label    string
	HabitID  int64
	Selected int
	Total    int
}

func momentumHabitGroup(habit sharedtypes.HabitWithMeta) string {
	repo, stream := strings.TrimSpace(habit.RepoName), strings.TrimSpace(habit.StreamName)
	switch {
	case repo != "" && stream != "":
		return repo + " › " + stream
	case repo != "":
		return repo
	case stream != "":
		return stream
	default:
		return "No context"
	}
}

func momentumHabitMatches(habit sharedtypes.HabitWithMeta, query string) bool {
	hay := strings.ToLower(strings.Join(
		append([]string{habit.Name, habit.RepoName, habit.StreamName}, habit.Tags...), " ",
	))
	for _, term := range strings.Fields(strings.ToLower(query)) {
		if !strings.Contains(hay, term) {
			return false
		}
	}
	return true
}

// MomentumHabitRows lists the habits the search matches, grouped under their context in the
// order contexts first appear. Headings count the picked habits among the visible ones.
func MomentumHabitRows(state State) []MomentumHabitRow {
	query := state.Momentum.Filter.Value()
	order := []string{}
	groups := map[string][]sharedtypes.HabitWithMeta{}
	for _, habit := range state.HabitItems {
		if !momentumHabitMatches(habit, query) {
			continue
		}
		group := momentumHabitGroup(habit)
		if _, ok := groups[group]; !ok {
			order = append(order, group)
		}
		groups[group] = append(groups[group], habit)
	}
	rows := []MomentumHabitRow{}
	for _, group := range order {
		habits := groups[group]
		heading := MomentumHabitRow{Heading: true, Group: group, Label: group, Total: len(habits)}
		for _, habit := range habits {
			if slices.Contains(state.HabitStreakDraft.HabitIDs, habit.ID) {
				heading.Selected++
			}
		}
		rows = append(rows, heading)
		for _, habit := range habits {
			rows = append(rows, MomentumHabitRow{Group: group, Label: habit.Name, HabitID: habit.ID})
		}
	}
	return rows
}

func momentumVisibleHabitIDs(rows []MomentumHabitRow) []int64 {
	ids := []int64{}
	for _, row := range rows {
		if !row.Heading {
			ids = append(ids, row.HabitID)
		}
	}
	return ids
}

// MomentumHabitCounts returns how many habits the search shows, how many exist, and how
// many picked habits the search hides.
func MomentumHabitCounts(state State) (visible, total, hiddenSelected int) {
	ids := momentumVisibleHabitIDs(MomentumHabitRows(state))
	for _, id := range state.HabitStreakDraft.HabitIDs {
		if !slices.Contains(ids, id) {
			hiddenSelected++
		}
	}
	return len(ids), len(state.HabitItems), hiddenSelected
}
