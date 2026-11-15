package commands

import (
	"cmp"
	"slices"

	shareddto "crona.work/shared/dto"
	"crona.work/tui/internal/api"
	tea "github.com/charmbracelet/bubbletea"
)

// CheckInHistoryLoadedMsg carries a complete history snapshot and its request generation.
type CheckInHistoryLoadedMsg struct {
	Records    []api.DailyCheckIn
	Generation uint64
	Err        error
}

// CheckInHistoryMutationMsg signals completion without changing Daily's selected date.
type CheckInHistoryMutationMsg struct{ Err error }

// LoadCheckInHistory loads every recorded day using the existing indexed range API.
func LoadCheckInHistory(c *api.Client, generation uint64) tea.Cmd {
	return func() tea.Msg {
		const firstDate, lastDate = "0001-01-01", "9999-12-31"
		records, err := c.ListDailyCheckIns(firstDate, lastDate)
		slices.SortFunc(records, func(a, b api.DailyCheckIn) int { return cmp.Compare(b.Date, a.Date) })
		return CheckInHistoryLoadedMsg{Records: records, Generation: generation, Err: err}
	}
}

// SaveHistoryCheckIn updates the explicitly selected historical record.
func SaveHistoryCheckIn(c *api.Client, input shareddto.DailyCheckInUpsertRequest) tea.Cmd {
	return func() tea.Msg {
		_, err := c.UpsertDailyCheckIn(input)
		return CheckInHistoryMutationMsg{Err: err}
	}
}

// DeleteHistoryCheckIn deletes the explicitly confirmed date.
func DeleteHistoryCheckIn(c *api.Client, date string) tea.Cmd {
	return func() tea.Msg { return CheckInHistoryMutationMsg{Err: c.DeleteDailyCheckIn(date)} }
}
