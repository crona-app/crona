package model

import (
	"crona.work/tui/internal/api"
	"crona.work/tui/internal/tui/commands"
	inputpkg "crona.work/tui/internal/tui/input"
	uistate "crona.work/tui/internal/tui/state"

	tea "github.com/charmbracelet/bubbletea"
)

// workRows are the loaded breakdown rows, when they answer the current query.
func (m Model) workRows() []uistate.WorkRow {
	if m.workBreakdown == nil || m.workBreakdownKey != m.workKey() {
		return nil
	}
	rows := make([]uistate.WorkRow, 0, len(m.workBreakdown.Rows))
	for _, row := range m.workBreakdown.Rows {
		if row.WorkedSeconds <= 0 {
			continue
		}
		rows = append(rows, uistate.WorkRow{Key: row.Key, Label: row.Label})
	}
	return rows
}

func (m Model) workKey() string {
	return uistate.WorkKey(m.currentDate(), m.workGroup, m.workPath)
}

// applyWorkSelect takes the breakdown's grouping, drill and selection from input. The
// selection lasts only while the Day summary is focused; leaving it drops the drill.
func (m Model) applyWorkSelect(state inputpkg.State) Model {
	m.workGroup = state.DailyWorkGroup
	m.workSelect = state.DailyWorkSelect && m.dailyDetailFocus && m.dailyDetailSummary
	if !m.workSelect {
		m.workCursor, m.workPath = 0, nil
		return m
	}
	m.workPath = append([]uistate.WorkCrumb(nil), state.DailyWorkPath...)
	m.workCursor = max(0, state.DailyWorkCursor)
	return m
}

// ensureWorkBreakdown loads the breakdown for the shown day, grouping and drill when it is
// not what was last asked for, or always when force is set.
func (m Model) ensureWorkBreakdown(force bool) (Model, tea.Cmd) {
	if m.client == nil {
		return m, nil
	}
	key := m.workKey()
	if !force && key == m.workWantKey {
		return m, nil
	}
	m.workWantKey = key
	repoID, streamID, tag := uistate.WorkFilters(m.workPath)
	group := uistate.WorkLevelGroup(m.workGroup, m.workPath)
	return m, commands.LoadWorkBreakdown(m.client, key, m.currentDate(), group, repoID, streamID, tag)
}

// workBreakdownShown is the breakdown when it answers the current query, else nil (loading).
func (m Model) workBreakdownShown() *api.TimeDistributionSummary {
	if m.workBreakdownKey != m.workKey() {
		return nil
	}
	return m.workBreakdown
}
