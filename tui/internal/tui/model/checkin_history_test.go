package model

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"crona.work/tui/internal/api"
	"crona.work/tui/internal/tui/commands"
	"crona.work/tui/internal/tui/dialogs"
	layoutpkg "crona.work/tui/internal/tui/layout"
	uistate "crona.work/tui/internal/tui/state"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func checkInModel() Model {
	return Model{
		view: ViewCheckInHistory, pane: PaneCheckIns, width: 80, height: 30, date: "2026-11-15",
		cursor: map[Pane]int{}, filters: map[Pane]string{}, checkInHistoryLoaded: true,
		dailyCheckIn: &api.DailyCheckIn{Date: "2026-11-15", Mood: 1, Energy: 1},
		checkInHistory: []api.DailyCheckIn{
			{Date: "2026-11-15", Mood: 5, Energy: 4, CreatedAt: "2026-11-15T09:15:00Z", Notes: new("Today")},
			{Date: "2026-11-14", Mood: 3, Energy: 2, CreatedAt: "2026-11-14T09:00:00Z", Notes: new("Yesterday")},
		},
	}
}

func TestCheckInHistoryEditAndDeleteTargetSelectedDate(t *testing.T) {
	m := previewKey(checkInModel(), tea.KeyDown)
	m = pressKey(m, 'e')
	if m.dialog != "edit_checkin" || m.dialogCheckInDate != "2026-11-14" || m.dialogInputs[0].Value() != "3" {
		t.Fatalf("edit targeted wrong record: %s %s %s", m.dialog, m.dialogCheckInDate, m.dialogInputs[0].Value())
	}
	if m.date != "2026-11-15" || m.dailyCheckIn.Mood != 1 {
		t.Fatal("historical edit changed Daily state")
	}
	m = previewKey(m, tea.KeyEsc)
	m = pressKey(m, 'D')
	state := m.dialogState()
	if state.Kind != "confirm_delete" || state.DeleteID != "2026-11-14" || !strings.Contains(state.DeleteLabel, "2026-11-14") {
		t.Fatalf("wrong delete confirmation: %+v", state)
	}
	m = previewKey(m, tea.KeyEsc)
	if len(m.checkInHistory) != 2 || m.date != "2026-11-15" {
		t.Fatal("cancelling delete changed history/date")
	}
}

func TestCheckInHistoryModalFiltersBrowsingAndResize(t *testing.T) {
	m := previewKey(checkInModel(), tea.KeyEnter)
	if m.dialog != "checkin_detail" {
		t.Fatalf("Enter opened %s", m.dialog)
	}
	m = previewKey(m, tea.KeyRight)
	if m.checkInDetailDate != "2026-11-14" {
		t.Fatal("modal did not browse to next record")
	}
	m = previewKey(m, tea.KeyRight)
	if m.cursor[PaneCheckIns] != 1 {
		t.Fatal("modal wrapped past list end")
	}
	next, _ := m.Update(tea.WindowSizeMsg{Width: 50, Height: 20})
	m = next.(Model)
	out := ansi.Strip(dialogs.Render(layoutpkg.DialogTheme(), m.dialogState()))
	t.Logf("Narrow detail modal:\n%s", out)
	if len(strings.Split(out, "\n")) > 20 || !strings.Contains(out, "[e] edit") {
		t.Fatalf("modal layout/actions missing:\n%s", out)
	}
	m = previewKey(m, tea.KeyEsc)
	m.filters[PaneCheckIns] = "Today"
	m.cursor[PaneCheckIns] = 0
	m = previewKey(m, tea.KeyEnter)
	m = previewKey(m, tea.KeyRight)
	if m.checkInDetailDate != "2026-11-15" {
		t.Fatal("modal browsing escaped filtered history")
	}
}

func TestCheckInHistoryGenerationsAndSelectionRefresh(t *testing.T) {
	m := checkInModel()
	m.cursor[PaneCheckIns] = 1
	m.checkInHistoryGeneration = 2
	old := commands.CheckInHistoryLoadedMsg{Generation: 1, Records: nil}
	next, _ := m.Update(old)
	m = next.(Model)
	if len(m.checkInHistory) != 2 {
		t.Fatal("stale response replaced history")
	}
	newer := append([]api.DailyCheckIn{{Date: "2026-11-16", Mood: 3, Energy: 3}}, m.checkInHistory...)
	next, _ = m.Update(commands.CheckInHistoryLoadedMsg{Generation: 2, Records: newer})
	m = next.(Model)
	entry, _ := m.selectedHistoryCheckIn()
	if entry.Date != "2026-11-14" {
		t.Fatal("refresh lost selected date")
	}
	next, _ = m.Update(commands.CheckInHistoryLoadedMsg{Generation: 2, Records: newer[:2]})
	m = next.(Model)
	entry, _ = m.selectedHistoryCheckIn()
	if entry.Date != "2026-11-15" {
		t.Fatal("deletion did not select nearest remaining date")
	}
	next, _ = m.Update(commands.CheckInHistoryLoadedMsg{Generation: 2, Err: errors.New("offline")})
	m = next.(Model)
	if m.checkInHistoryError == "" || len(m.checkInHistory) != 2 {
		t.Fatal("refresh error lost cached history")
	}
	m = pressKey(m, 'r')
	if !m.checkInHistoryLoading || m.checkInHistoryGeneration != 3 {
		t.Fatal("retry did not start a new request")
	}
}

func TestCheckInHistoryRemainsInLogDuringTimerAndRest(t *testing.T) {
	m := checkInModel()
	m.timer = &api.TimerState{State: "running"}
	if !slices.Contains(m.visibleTabViews(uistate.TabLog), ViewCheckInHistory) || !m.canJumpToView(ViewCheckInHistory) {
		t.Fatal("timer hid Check-ins history")
	}
	m.settings = &api.CoreSettings{RestSpecificDates: []string{time.Now().Format(time.DateOnly)}}
	if m.effectiveView() != ViewCheckInHistory {
		t.Fatal("rest hid check-in history")
	}
	m = pressKey(m, 'e')
	if m.dialog != "" {
		t.Fatal("protected rest allowed historical editing")
	}
}

func TestCheckInModalActionsRemainBoundToDisplayedDate(t *testing.T) {
	for _, key := range []rune{'e', 'D'} {
		m := checkInModel()
		m.cursor[PaneCheckIns] = 1
		m = previewKey(m, tea.KeyEnter)
		m.cursor[PaneCheckIns] = 0 // Simulate a source cursor adjustment during an async refresh.
		m = pressKey(m, key)
		state := m.dialogState()
		if key == 'e' && state.CheckInDate != "2026-11-14" {
			t.Fatal("modal edit followed cursor instead of displayed date")
		}
		if key == 'D' && state.DeleteID != "2026-11-14" {
			t.Fatal("modal delete followed cursor instead of displayed date")
		}
	}
}
