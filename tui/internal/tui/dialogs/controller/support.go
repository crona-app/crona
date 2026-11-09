package controller

import (
	"fmt"
	"strings"

	sharedtypes "crona.work/shared/types"
	tea "github.com/charmbracelet/bubbletea"

	uistate "crona.work/tui/internal/tui/state"
)

type menuChoice struct {
	Key    string
	Label  string
	Value  string
	Detail string
}

// viewJumpChoices are each view's jump key and one-line description.
var viewJumpChoices = map[uistate.View]menuChoice{
	uistate.ViewDaily:          {Key: "d", Detail: "planned issues and habits"},
	uistate.ViewSessionActive:  {Key: "n", Detail: "the running session"},
	uistate.ViewAway:           {Key: "a", Detail: "away and rest mode"},
	uistate.ViewDefault:        {Key: "i", Detail: "every issue in scope"},
	uistate.ViewHabitList:      {Key: "m", Detail: "habit definitions"},
	uistate.ViewTags:           {Key: "+", Detail: "rename, merge and color tags"},
	uistate.ViewReview:         {Key: "r", Detail: "work and wellbeing over a window"},
	uistate.ViewMomentum:       {Key: "u", Detail: "momentum cards and progress"},
	uistate.ViewSessionHistory: {Key: "y", Detail: "past focus sessions"},
	uistate.ViewHabitHistory:   {Key: "b", Detail: "past habit check-offs"},
	uistate.ViewOps:            {Key: "o", Detail: "recent operations"},
	uistate.ViewSettings:       {Key: "s", Detail: "core settings"},
	uistate.ViewAlerts:         {Key: "l", Detail: "notifications and sounds"},
	uistate.ViewReports:        {Key: "p", Detail: "generated reports"},
	uistate.ViewConfig:         {Key: "c", Detail: "export assets and templates"},
	uistate.ViewUpdates:        {Key: "t", Detail: "releases and install status"},
	uistate.ViewSupport:        {Key: "h", Detail: "bug reports and links"},
}

// OpenViewJump opens the jump dialog over availableViews.
func OpenViewJump(state State, availableViews []uistate.View) State {
	return OpenViewJumpFrom(state, availableViews, "")
}

// OpenViewJumpFrom lists availableViews grouped by tab, in the tab bar's order and names, with
// the cursor on current.
func OpenViewJumpFrom(state State, availableViews []uistate.View, current uistate.View) State {
	state = Close(state)
	state.Kind = "view_jump"
	state.ViewTitle = "Jump to"
	state.ViewName = ""
	state.ChoiceCurrent = string(current)
	allowed := make(map[uistate.View]bool, len(availableViews))
	for _, view := range availableViews {
		allowed[view] = true
	}
	var choices []menuChoice
	var groups []string
	for i, tab := range uistate.TabOrder() {
		group := fmt.Sprintf("%s\t%d", uistate.TabLabel(tab), i+1)
		for _, view := range uistate.TabViews(tab) {
			choice, ok := viewJumpChoices[view]
			if !allowed[view] || !ok {
				continue
			}
			choice.Label, choice.Value = uistate.ViewLabel(view), string(view)
			choices = append(choices, choice)
			groups = append(groups, group)
			group = ""
		}
	}
	state.ChoiceItems, state.ChoiceValues, state.ChoiceDetails = menuChoiceLists(choices)
	state.ChoiceGroups = groups
	for i, value := range state.ChoiceValues {
		if value == state.ChoiceCurrent {
			state.ChoiceCursor = i
		}
	}
	return state
}

func OpenBetaSupport(state State) State {
	state = Close(state)
	state.Kind = "beta_support"
	state.ViewTitle = "Beta Support"
	state.ViewName = "Quick reporting and diagnostics for beta testers"
	choices := []menuChoice{
		{
			Key:    "o",
			Label:  "Feedback & Discussions",
			Value:  "open_support_feedback_roadmap",
			Detail: "Share feedback, report bugs and ask questions in GitHub Discussions.",
		},
		{
			Key:    "r",
			Label:  "Releases",
			Value:  "open_support_releases",
			Detail: "Open the GitHub releases feed.",
		},
		{
			Key:    "g",
			Label:  "Documentation",
			Value:  "open_support_documentation",
			Detail: "Open Crona's guides and reference documentation.",
		},
		{
			Key:    "c",
			Label:  "Copy Diagnostics",
			Value:  "copy_support_diagnostics",
			Detail: "Copy the lightweight diagnostics summary.",
		},
		{
			Key:    "b",
			Label:  "Generate Bundle",
			Value:  "generate_support_bundle",
			Detail: "Create a full redacted support bundle.",
		},
	}
	state.ChoiceItems, state.ChoiceValues, state.ChoiceDetails = menuChoiceLists(choices)
	return state
}

func OpenViewEntity(state State, title string, name string, meta string, body string) State {
	return OpenViewEntityWithPath(state, title, name, meta, body, "")
}

func OpenViewEntityWithPath(
	state State,
	title string,
	name string,
	meta string,
	body string,
	path string,
) State {
	state = Close(state)
	state.Kind = "view_entity"
	state.ViewTitle = title
	state.ViewName = name
	state.ViewMeta = meta
	state.ViewBody = body
	state.ViewPath = strings.TrimSpace(path)
	state.ViewScroll = 0
	return state
}

func OpenSupportBundleResult(state State, name, meta, body, path string) State {
	state = Close(state)
	state.Kind = "support_bundle_result"
	state.ViewTitle = "Support Bundle Ready"
	state.ViewName = name
	state.ViewMeta = meta
	state.ViewBody = body
	state.SupportBundlePath = path
	return state
}

func updateViewEntity(state State, msg tea.KeyMsg) (State, *Action, string) {
	// The renderer clamps the far end; see dialogs.ClampViewScroll.
	page := max(1, state.Height-12)
	switch msg.String() {
	case "up", "k":
		state.ViewScroll = max(0, state.ViewScroll-1)
		return state, nil, ""
	case "down", "j":
		state.ViewScroll++
		return state, nil, ""
	case "pgup", "ctrl+u":
		state.ViewScroll = max(0, state.ViewScroll-page)
		return state, nil, ""
	case "pgdown", "ctrl+d", " ":
		state.ViewScroll += page
		return state, nil, ""
	case "home", "g":
		state.ViewScroll = 0
		return state, nil, ""
	case "end", "G":
		state.ViewScroll = 1 << 20
		return state, nil, ""
	case "esc", "enter", "?":
		return Close(state), nil, ""
	case "e":
		if state.ViewEntityAdhoc && state.IssueID > 0 {
			return Close(state), &Action{Kind: "open_adhoc_conversion", IssueID: state.IssueID}, ""
		}
		if strings.TrimSpace(state.ViewPath) != "" {
			return Close(state), &Action{Kind: "open_view_entity_editor", Path: state.ViewPath}, ""
		}
		if state.HabitStreakDraft.ID != "" {
			return Close(state), &Action{
				Kind:            "open_momentum_editor",
				HabitStreakDefs: []sharedtypes.HabitStreakDefinition{state.HabitStreakDraft},
			}, ""
		}
	case "c":
		switch state.ViewName {
		case "Reports directory":
			return Close(state), &Action{Kind: "open_export_reports_dir_dialog"}, ""
		case "ICS export directory":
			return Close(state), &Action{Kind: "open_export_ics_dir_dialog"}, ""
		}
	case "r":
		switch state.ViewName {
		case "Reports directory":
			return Close(state), &Action{Kind: "reset_export_reports_dir"}, ""
		case "ICS export directory":
			return Close(state), &Action{Kind: "reset_export_ics_dir"}, ""
		}
	default:
		return clearDialogError(state), nil, ""
	}
	return clearDialogError(state), nil, ""
}

func updateSupportBundleResult(state State, msg tea.KeyMsg) (State, *Action, string) {
	switch msg.String() {
	case "esc", "enter":
		return Close(state), nil, ""
	case "o":
		if strings.TrimSpace(state.SupportBundlePath) == "" {
			return state, nil, "Bundle path is unavailable"
		}
		return clearDialogError(
				state,
			), &Action{
				Kind: "open_support_bundle_folder",
				Path: state.SupportBundlePath,
			}, ""
	case "c":
		if strings.TrimSpace(state.SupportBundlePath) == "" {
			return state, nil, "Bundle path is unavailable"
		}
		return clearDialogError(
				state,
			), &Action{
				Kind: "copy_support_bundle_path",
				Path: state.SupportBundlePath,
			}, ""
	case "g":
		return clearDialogError(state), &Action{Kind: "open_support_feedback_roadmap"}, ""
	default:
		return clearDialogError(state), nil, ""
	}
}

func updateViewJump(state State, msg tea.KeyMsg) (State, *Action, string) {
	// No view's key is j or k, so they move like the arrows.
	switch msg.String() {
	case "j":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "k":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	}
	return updateChoiceMenu(state, msg, true)
}

func updateBetaSupport(state State, msg tea.KeyMsg) (State, *Action, string) {
	return updateChoiceMenu(state, msg, false)
}

func updateChoiceMenu(state State, msg tea.KeyMsg, jumpMenu bool) (State, *Action, string) {
	switch msg.String() {
	case "esc":
		return Close(state), nil, ""
	case "down":
		if state.ChoiceCursor < len(state.ChoiceItems)-1 {
			state.ChoiceCursor++
		}
		return clearDialogError(state), nil, ""
	case "up":
		if state.ChoiceCursor > 0 {
			state.ChoiceCursor--
		}
		return clearDialogError(state), nil, ""
	case "enter":
		return resolveChoiceSelection(state, jumpMenu)
	default:
		for idx, item := range state.ChoiceItems {
			if msg.String() == menuChoiceKey(item) {
				state.ChoiceCursor = idx
				return resolveChoiceSelection(state, jumpMenu)
			}
		}
		return clearDialogError(state), nil, ""
	}
}

func resolveChoiceSelection(state State, jumpMenu bool) (State, *Action, string) {
	if state.ChoiceCursor < 0 || state.ChoiceCursor >= len(state.ChoiceValues) {
		return clearDialogError(state), nil, ""
	}
	value := strings.TrimSpace(state.ChoiceValues[state.ChoiceCursor])
	if value == "" {
		return clearDialogError(state), nil, ""
	}
	if jumpMenu {
		return Close(state), &Action{Kind: "jump_view", TargetView: value}, ""
	}
	return Close(state), &Action{Kind: value}, ""
}

func menuChoiceLists(choices []menuChoice) ([]string, []string, []string) {
	items := make([]string, 0, len(choices))
	values := make([]string, 0, len(choices))
	details := make([]string, 0, len(choices))
	for _, choice := range choices {
		items = append(items, "["+choice.Key+"] "+choice.Label)
		values = append(values, choice.Value)
		details = append(details, choice.Detail)
	}
	return items, values, details
}

func menuChoiceKey(label string) string {
	label = strings.TrimSpace(label)
	if !strings.HasPrefix(label, "[") {
		return ""
	}
	end := strings.Index(label, "]")
	if end <= 1 {
		return ""
	}
	return strings.TrimSpace(label[1:end])
}
