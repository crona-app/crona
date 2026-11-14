package model

import (
	"fmt"
	"maps"
	"strings"

	"crona.work/tui/internal/tui/dialogs"
	dialogstate "crona.work/tui/internal/tui/dialogs/controller"
	inputpkg "crona.work/tui/internal/tui/input"
	layoutpkg "crona.work/tui/internal/tui/layout"
	selectionpkg "crona.work/tui/internal/tui/selection"
	viewchrome "crona.work/tui/internal/tui/views/chrome"
	"crona.work/tui/internal/tui/views/itemdetail"

	tea "github.com/charmbracelet/bubbletea"
)

type itemDialogContent struct {
	kind, title string
	body        []string
	footer      [][2]string
	adhoc       bool
	issueID     int64
}

func (m Model) itemDialogContent() (itemDialogContent, bool) {
	snapshot := m.selectionSnapshot()
	state := m.detailContentState(snapshot)
	state.ItemDetailActions = true
	theme, width := layoutpkg.ViewTheme(), viewEntityBodyWidth(m.width)
	if m.pane == PaneIssues {
		issue, ok := selectionpkg.SelectedIssue(snapshot)
		if !ok {
			return itemDialogContent{}, false
		}
		return itemDialogContent{
			kind: "Issue", title: issue.Title, issueID: issue.ID, adhoc: issue.Adhoc,
			body:   itemdetail.IssueLines(theme, state, *issue, width),
			footer: itemdetail.IssueFooterKeys(state, *issue),
		}, true
	}
	if m.pane != PaneHabits {
		return itemDialogContent{}, false
	}
	if habit, ok := selectionpkg.SelectedDailyHabit(snapshot); ok {
		return itemDialogContent{
			kind: "Habit", title: habit.Name,
			body:   itemdetail.HabitLines(theme, state, *habit, width),
			footer: itemdetail.HabitFooterKeys(state, habit.Habit),
		}, true
	}
	if m.view == ViewHabitList {
		idx := selectionpkg.FilteredIndexAtCursor(snapshot, PaneHabits)
		habits := selectionpkg.ScopedHabits(snapshot)
		if idx >= 0 && idx < len(habits) {
			habit := habits[idx]
			return itemDialogContent{
				kind: "Habit", title: habit.Name,
				body:   itemdetail.DefinitionLines(theme, state, habit, width),
				footer: itemdetail.HabitFooterKeys(state, habit.Habit),
			}, true
		}
	}
	return itemDialogContent{}, false
}

func (m Model) openItemDetailDialog() (Model, bool) {
	content, ok := m.itemDialogContent()
	if !ok {
		return m, false
	}
	state := m.dialogSnapshot().OpenViewEntity(content.kind, content.title, "", strings.Join(content.body, "\n"))
	state.ViewItemDetail = true
	state.ViewEntityAdhoc, state.IssueID = content.adhoc, content.issueID
	m = m.withDialogState(state)
	return m.refreshItemDetailDialog(), true
}

func (m Model) isItemDetailDialog() bool {
	return m.dialog == "view_entity" && m.dialogViewItemDetail
}

func (m Model) refreshItemDetailDialog() Model {
	if !m.isItemDetailDialog() {
		return m
	}
	content, ok := m.itemDialogContent()
	if !ok {
		return m.withDialogState(dialogstate.Close(m.dialogState()))
	}
	state := m.dialogState()
	state.ViewTitle, state.ViewName, state.ViewBody = content.kind, content.title, strings.Join(content.body, "\n")
	state.ViewEntityAdhoc, state.IssueID = content.adhoc, content.issueID
	state.ViewDetailHints = nil
	for _, key := range content.footer {
		state.ViewDetailHints = append(state.ViewDetailHints, fmt.Sprintf("%s %s", key[0], key[1]))
	}
	return m.withDialogState(dialogs.ClampViewScroll(layoutpkg.DialogTheme(), state))
}

// updateItemDetailDialog routes advertised actions to the same handlers as the source list.
func (m Model) updateItemDetailDialog(key tea.KeyMsg) (Model, tea.Cmd, bool) {
	if !m.isItemDetailDialog() {
		return m, nil, false
	}
	switch key.String() {
	case "left", "right":
		delta := 1
		if key.String() == "left" {
			delta = -1
		}
		m.cursor = maps.Clone(m.cursor)
		if m.cursor == nil {
			m.cursor = map[Pane]int{}
		}
		m.cursor[m.pane] = max(0, min(m.listLen(m.pane)-1, m.cursor[m.pane]+delta))
		m.dialogViewScroll = 0
		return m.refreshItemDetailDialog(), nil, true
	case "s", "d", "P", "+", "r", "f", "e", "D", "m", "x", "F":
		content, ok := m.itemDialogContent()
		if !ok {
			return m, nil, true
		}
		allowed := false
		for _, action := range content.footer {
			allowed = allowed || action[0] == key.String()
		}
		if content.kind == "Issue" && !content.adhoc {
			allowed = allowed || key.String() == "s" || key.String() == "d" || key.String() == "P" || key.String() == "+" || key.String() == "m"
		}
		if content.kind == "Habit" {
			allowed = allowed || key.String() == "+" || key.String() == "m" || (m.view == ViewDaily && (key.String() == "x" || key.String() == "F"))
		}
		if !allowed {
			return m, nil, true
		}
		previous := m.dialogState()
		next := m.withDialogState(dialogstate.Close(previous))
		state, cmd := inputpkg.Handle(next.inputState(), key, next.inputDeps())
		next = next.applyInputState(state)
		if next.dialog == "" && next.view == m.view && key.String() != "f" {
			next = next.withDialogState(previous)
		}
		return next.refreshItemDetailDialog(), cmd, true
	}
	return m, nil, false
}

// workDetailPaneShown mirrors the Work layouts: a side pane from 160 columns, or a
// stacked pane when the body (minus the narrow scope row) is tall enough.
func (m Model) workDetailPaneShown() bool {
	if m.view != ViewDefault && m.view != ViewHabitList {
		return false
	}
	height := m.contentHeight()
	if m.width >= 160 {
		return height > 0
	}
	if m.width < 100 {
		height--
	}
	return height >= viewchrome.StackMinHeight
}
