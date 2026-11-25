package model

import (
	"maps"
	"strings"

	"crona.work/tui/internal/api"
	"crona.work/tui/internal/tui/commands"
	"crona.work/tui/internal/tui/dialogs"
	dialogstate "crona.work/tui/internal/tui/dialogs/controller"
	helperpkg "crona.work/tui/internal/tui/helpers"
	layoutpkg "crona.work/tui/internal/tui/layout"
	"crona.work/tui/internal/tui/views/checkins"
	viewruntime "crona.work/tui/internal/tui/views/runtime"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) selectedHistoryCheckIn() (*api.DailyCheckIn, bool) {
	indices := checkins.FilteredIndices(m.checkInHistory, m.filters[PaneCheckIns], m.settings)
	cur := m.cursor[PaneCheckIns]
	if cur < 0 || cur >= len(indices) {
		return nil, false
	}
	entry := m.checkInHistory[indices[cur]]
	return &entry, true
}

func (m *Model) syncCheckInHistory(entering bool) tea.Cmd {
	if m.view != ViewCheckInHistory {
		return nil
	}
	if entering {
		m.checkInHistoryStale = true
	}
	if !m.checkInHistoryStale && (m.checkInHistoryLoaded || m.checkInHistoryLoading || m.checkInHistoryError != "") {
		return nil
	}
	m.checkInHistoryGeneration++
	m.checkInHistoryLoading, m.checkInHistoryStale = true, false
	m.checkInHistoryError = ""
	if m.client == nil {
		return nil
	}
	return commands.LoadCheckInHistory(m.client, m.checkInHistoryGeneration)
}

func (m Model) applyCheckInHistory(msg commands.CheckInHistoryLoadedMsg) Model {
	if msg.Generation != m.checkInHistoryGeneration {
		return m
	}
	m.checkInHistoryLoading = false
	if msg.Err != nil {
		m.checkInHistoryError = helperpkg.ErrorMessage(msg.Err)
		return m
	}
	selectedDate := ""
	if selected, ok := m.selectedHistoryCheckIn(); ok {
		selectedDate = selected.Date
	}
	m.cursor = maps.Clone(m.cursor)
	m.checkInHistory, m.checkInHistoryLoaded, m.checkInHistoryError = msg.Records, true, ""
	if m.cursor == nil {
		m.cursor = map[Pane]int{}
	}
	indices := checkins.FilteredIndices(m.checkInHistory, m.filters[PaneCheckIns], m.settings)
	m.cursor[PaneCheckIns] = max(0, min(m.cursor[PaneCheckIns], len(indices)-1))
	for pos, idx := range indices {
		if m.checkInHistory[idx].Date == selectedDate {
			m.cursor[PaneCheckIns] = pos
			break
		}
	}
	return m.refreshCheckInModal()
}

func (m Model) openCheckInHistoryDetail() Model {
	entry, ok := m.selectedHistoryCheckIn()
	if !ok {
		return m
	}
	state := m.dialogSnapshot().OpenViewEntity("Check-in", entry.Date, "", "")
	state.Kind = "checkin_detail"
	if !m.checkInHistoryReadOnly() {
		state.ViewDetailHints = []string{"e edit", "D delete"}
	}
	m.checkInDetailDate = entry.Date
	return m.withDialogState(state).refreshCheckInModal()
}

func (m Model) refreshCheckInModal() Model {
	if m.dialog != "checkin_detail" {
		return m
	}
	for _, entry := range m.checkInHistory {
		if entry.Date == m.checkInDetailDate {
			state := m.dialogState()
			state.ViewDetailHints = nil
			if !m.checkInHistoryReadOnly() {
				state.ViewDetailHints = []string{"e edit", "D delete"}
			}
			state.ViewName = entry.Date
			state.ViewBody = strings.Join(checkins.DetailLines(layoutpkg.ViewTheme(), m.layoutState().ContentState, &entry, viewEntityBodyWidth(m.width)), "\n")
			return m.withDialogState(dialogs.ClampViewScroll(layoutpkg.DialogTheme(), state))
		}
	}
	return m.withDialogState(dialogstate.Close(m.dialogState()))
}

// checkInActionTarget keeps modal actions bound to the displayed date even if a refresh
// changes the source list's cursor while the dialog is open.
func (m Model) checkInActionTarget() (*api.DailyCheckIn, bool) {
	if m.dialog == "checkin_detail" {
		for _, entry := range m.checkInHistory {
			if entry.Date == m.checkInDetailDate {
				return &entry, true
			}
		}
		return nil, false
	}
	return m.selectedHistoryCheckIn()
}

func (m Model) editHistoryCheckIn() Model {
	entry, ok := m.checkInActionTarget()
	if !ok {
		return m
	}
	return m.withDialogState(dialogstate.OpenEditCheckIn(m.dialogState(), entry, entry.Date))
}

func (m Model) handleCheckInHistoryKey(key tea.KeyMsg) (Model, tea.Cmd, bool) {
	if m.view != ViewCheckInHistory {
		return m, nil, false
	}
	modal := m.dialog == "checkin_detail"
	if m.dialog != "" && !modal {
		return m, nil, false
	}
	switch key.String() {
	case "enter":
		if modal {
			return m, nil, false
		}
		return m.openCheckInHistoryDetail(), nil, true
	case "e":
		if m.checkInHistoryReadOnly() {
			return m, nil, true
		}
		return m.editHistoryCheckIn(), nil, true
	case "D":
		if m.checkInHistoryReadOnly() {
			return m, nil, true
		}
		if entry, ok := m.checkInActionTarget(); ok {
			return m.openConfirmDeleteEntity("checkin", entry.Date, "check-in for "+entry.Date), nil, true
		}
		return m, nil, true
	case "r":
		if modal {
			return m, nil, false
		}
		m.checkInHistoryStale = true
		return m, nil, true
	case "left", "right":
		if !modal {
			return m, nil, true
		}
		delta := 1
		if key.String() == "left" {
			delta = -1
		}
		m.cursor = maps.Clone(m.cursor)
		if m.cursor == nil {
			m.cursor = map[Pane]int{}
		}
		m.cursor[PaneCheckIns] = max(0, min(m.listLen(PaneCheckIns)-1, m.cursor[PaneCheckIns]+delta))
		if entry, ok := m.selectedHistoryCheckIn(); ok {
			m.checkInDetailDate = entry.Date
		}
		m.dialogViewScroll = 0
		return m.refreshCheckInModal(), nil, true
	case "a", "A", "I", "f", "m", "s", "d", "P", "+", "x", "F", "y", "w", "W", "c", "E":
		if !modal {
			return m, nil, true
		}
	}
	return m, nil, false
}

func (m Model) checkInHistoryReadOnly() bool {
	protected, _, _ := viewruntime.ProtectedRestMode(m.settings, m.currentLogicalDate())
	return protected
}
