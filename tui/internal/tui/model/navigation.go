package model

import (
	"crona.work/tui/internal/tui/scope"
	"fmt"
	"slices"
	"strings"
	"time"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	"crona.work/tui/internal/tui/commands"
	"crona.work/tui/internal/tui/dialogs"
	dialogstate "crona.work/tui/internal/tui/dialogs/controller"
	layoutpkg "crona.work/tui/internal/tui/layout"
	navigationutil "crona.work/tui/internal/tui/navigationutil"
	selectionpkg "crona.work/tui/internal/tui/selection"
	uistate "crona.work/tui/internal/tui/state"
	viewruntime "crona.work/tui/internal/tui/views/runtime"
	viewtypes "crona.work/tui/internal/tui/views/types"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) availableViews() []View {
	if protected, _, _ := viewruntime.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02")); protected {
		return []View{ViewAway, ViewReports, ViewSessionHistory, ViewHabitHistory, ViewSettings}
	}
	ordered := uistate.ViewOrder()
	views := make([]View, 0, len(ordered))
	views = append(views, ordered...)
	if len(views) == 0 {
		return []View{ViewDaily}
	}
	return views
}

func (m Model) jumpAvailableViews() []View {
	if protected, _, _ := viewruntime.ProtectedRestMode(m.settings, time.Now().Format("2006-01-02")); protected {
		return []View{ViewAway, ViewReports, ViewSessionHistory, ViewHabitHistory, ViewSettings}
	}
	if m.timer != nil && m.timer.State != "idle" {
		return []View{ViewSessionActive, ViewSessionHistory}
	}
	return m.availableViews()
}

func (m Model) canJumpToView(target View) bool {
	canJump := slices.Contains(m.jumpAvailableViews(), target)

	return canJump
}

func (m Model) visibleTabViews(tab uistate.Tab) []View {
	available := m.jumpAvailableViews()
	views := []View{}
	for _, view := range uistate.TabViews(tab) {
		if slices.Contains(available, view) {
			views = append(views, view)
		}
	}
	return views
}

func (m Model) visibleTabs() []uistate.Tab {
	tabs := []uistate.Tab{}
	for _, tab := range uistate.TabOrder() {
		if len(m.visibleTabViews(tab)) > 0 {
			tabs = append(tabs, tab)
		}
	}
	return tabs
}

func (m Model) tabTargetView(tab uistate.Tab) (View, bool) {
	views := m.visibleTabViews(tab)
	if len(views) == 0 {
		return "", false
	}
	if slices.Contains(views, m.view) {
		return m.view, true
	}
	return views[0], true
}

func (m Model) nextSubTabView(dir int) View {
	views := m.visibleTabViews(uistate.TabForView(m.view))
	if !slices.Contains(views, m.view) {
		if len(views) > 0 {
			return views[0]
		}
		return m.view
	}
	if len(views) < 2 {
		return m.view
	}
	return navigationutil.NextView(views, m.view, dir)
}

func nextPane(view View, current Pane, dir int) Pane {
	return uistate.NextPane(view, current, dir)
}

// enterIssuesTab starts the scope tree on the checked-out context.
func (m *Model) enterIssuesTab() {
	m.browseScope = scope.FromContext(m.context)
	if m.cursor == nil {
		m.cursor = map[Pane]int{}
	}
	m.cursor[PaneScope] = scope.IndexOf(m.scopeNodes(), m.browseScope)
}

func (m Model) scopeNodes() []scope.Node {
	return scope.Nodes(m.repos, m.allStreams)
}

// browseToScopeCursor scopes the lists to the tree's node under its cursor.
func (m *Model) browseToScopeCursor() {
	nodes := m.scopeNodes()
	if cur := m.cursor[PaneScope]; cur >= 0 && cur < len(nodes) {
		m.browseScope = nodes[cur].Scope
	}
	m.clampIssuesTabLists()
}

// clampIssuesTabLists derives each list from its own view, independent of focus.
func (m *Model) clampIssuesTabLists() {
	snapshot := m.selectionSnapshot()
	snapshot.View = ViewDefault
	m.clamp(PaneIssues, len(selectionpkg.FilteredIndices(snapshot, PaneIssues)))
	snapshot.View = ViewHabitList
	m.clamp(PaneHabits, len(selectionpkg.FilteredIndices(snapshot, PaneHabits)))
	m.cursor[PaneIssues] = max(0, m.cursor[PaneIssues])
	m.cursor[PaneHabits] = max(0, m.cursor[PaneHabits])
}

func (m *Model) initializeIssuesTransition(previous View) {
	if uistate.TabForView(m.view) != uistate.TabIssues {
		return
	}
	entering := uistate.TabForView(previous) != uistate.TabIssues
	if entering {
		m.enterIssuesTab()
	}
	m.clampIssuesTabLists()
}

func (m Model) issuesTransitionRefresh(previous View) tea.Cmd {
	entering := uistate.TabForView(m.view) == uistate.TabIssues &&
		uistate.TabForView(previous) != uistate.TabIssues
	if entering || m.view == ViewHabitList && previous != ViewHabitList {
		return commands.LoadTodayHabits(m.client, m.currentLogicalDate())
	}
	return nil
}

// reconcileBrowseScope keeps the tree's cursor on the browsed node once all streams are
// known, falling back to All when that repo or stream is gone.
func (m *Model) reconcileBrowseScope() {
	nodes := m.scopeNodes()
	idx := scope.IndexOf(nodes, m.browseScope)
	if nodes[idx].Scope != m.browseScope {
		m.browseScope, idx = scope.Scope{}, 0
	}
	if m.cursor == nil {
		m.cursor = map[Pane]int{}
	}
	m.cursor[PaneScope] = idx
}

// selectedScopeNode is the tree's node under its cursor while the tree has focus.
func (m Model) selectedScopeNode() (scope.Node, bool) {
	return selectionpkg.SelectedScopeNode(m.selectionSnapshot())
}

func (m Model) scopeStream(id int64) (api.Stream, bool) {
	for _, stream := range m.allStreams {
		if stream.ID == id {
			return stream, true
		}
	}
	return api.Stream{}, false
}

func (m *Model) setDefaultIssueSection(section DefaultIssueSection) {
	if section != DefaultIssueSectionOpen && section != DefaultIssueSectionCompleted {
		section = DefaultIssueSectionOpen
	}
	m.defaultIssueSection = section
	if m.view != ViewDefault || m.pane != PaneIssues {
		return
	}
	m.cursor[PaneIssues] = 0
	snapshot := m.selectionSnapshot()
	indices := selectionpkg.FilteredIndices(snapshot, PaneIssues)
	m.clamp(PaneIssues, len(indices))
}

func (m *Model) setDailyTaskSection(section DailyTaskSection) {
	if section != DailyTaskSectionPlanned && section != DailyTaskSectionOverdue &&
		section != DailyTaskSectionUpcoming {
		section = DailyTaskSectionPlanned
	}
	m.dailyTaskSection = section
	if m.view != ViewDaily || m.pane != PaneIssues {
		return
	}
	snapshot := m.selectionSnapshot()
	indices := selectionpkg.FilteredIndices(snapshot, PaneIssues)
	m.clamp(PaneIssues, len(indices))
}

func (m Model) selectedIssueRecord() (*api.Issue, bool) {
	snapshot := m.selectionSnapshot()
	if m.timer != nil && m.timer.State != "idle" &&
		m.view == ViewSessionActive {
		if issue := selectionpkg.ActiveIssue(snapshot); issue != nil {
			copy := issue.Issue
			return &copy, true
		}
	}
	if m.pane != PaneIssues {
		return nil, false
	}
	return selectionpkg.SelectedIssue(snapshot)
}

func (m Model) selectedHabitRecord() (*api.Habit, bool) {
	return selectionpkg.SelectedHabit(m.selectionSnapshot())
}

func (m Model) selectedDailyHabitRecord() (*api.HabitDailyItem, bool) {
	return selectionpkg.SelectedDailyHabit(m.selectionSnapshot())
}

func (m Model) selectedHabitForAction() (*api.Habit, bool) {
	if m.view == ViewDaily {
		if habit, ok := m.selectedDailyHabitRecord(); ok {
			copy := habit.Habit
			return &copy, true
		}
		return nil, false
	}
	return m.selectedHabitRecord()
}

func (m Model) selectedSessionHistoryEntry() (*api.SessionHistoryEntry, bool) {
	return selectionpkg.SelectedSessionHistoryEntry(m.selectionSnapshot())
}

func (m Model) selectedHabitHistoryEntry() (*api.HabitCompletion, bool) {
	return selectionpkg.SelectedHabitHistoryEntry(m.selectionSnapshot())
}

func (m Model) selectedMomentumCard() (*api.MomentumCard, bool) {
	return selectionpkg.SelectedMomentumCard(m.selectionSnapshot())
}

func (m Model) openSelectedEditDialog() (Model, bool) {
	switch m.pane {
	case PaneScope:
		node, ok := m.selectedScopeNode()
		if !ok || node.Scope.IsAll() {
			return m, false
		}
		if node.Scope.StreamID == 0 {
			return m.openEditRepoDialog(node.Scope.RepoID, node.Name), true
		}
		return m.openEditStreamDialog(
			node.Scope.StreamID,
			node.Scope.RepoID,
			node.Name,
			m.repoNameByID(node.Scope.RepoID),
		), true
	case PaneIssues:
		if issue, ok := m.selectedIssueRecord(); ok {
			if issue.Adhoc {
				if meta, found := m.issueMetaByID(issue.ID); found {
					return m.withDialogState(m.dialogSnapshot().OpenPromoteAdhocIssue(*meta)), true
				}
			}
			if stream, found := m.scopeStream(issue.StreamID); found && stream.System {
				if meta, found := m.issueMetaByID(issue.ID); found {
					return m.withDialogState(m.dialogSnapshot().OpenMoveInboxIssue(*meta)), true
				}
			}
			return m.openEditIssueDialog(
				issue.ID,
				issue.StreamID,
				issue.Title,
				issue.Description,
				issue.EstimateMinutes,
				issue.TodoForDate,
			), true
		}
	case PaneHabits:
		if m.view == ViewDaily {
			if habit, ok := m.selectedDailyHabitRecord(); ok {
				return m.openEditHabitDialog(
					habit.ID,
					habit.StreamID,
					habit.Name,
					habit.Description,
					string(habit.ScheduleType),
					habit.Weekdays,
					habit.TargetMinutes,
					habit.Active,
				), true
			}
		}
		if habit, ok := m.selectedHabitRecord(); ok {
			return m.openEditHabitDialog(
				habit.ID,
				habit.StreamID,
				habit.Name,
				habit.Description,
				string(habit.ScheduleType),
				habit.Weekdays,
				habit.TargetMinutes,
				habit.Active,
			), true
		}
	}
	return m, false
}

// viewEntityBodyWidth matches the modal content width, including narrow-screen clamping.
func viewEntityBodyWidth(width int) int {
	return dialogs.FrameInner(width, []string{strings.Repeat(" ", 76)})
}

func (m Model) openSelectedViewDialog() (Model, bool) {
	switch m.pane {
	case PaneScope:
		node, ok := m.selectedScopeNode()
		if !ok || node.Scope.IsAll() {
			return m, false
		}
		if node.Scope.StreamID == 0 {
			body := strings.Join(
				[]string{"Description", optionalText(m.repoDescriptionByID(node.Scope.RepoID))},
				"\n",
			)
			return m.openViewEntityDialog(
				"Repo",
				node.Name,
				fmt.Sprintf("ID %d", node.Scope.RepoID),
				body,
			), true
		}
		stream, _ := m.scopeStream(node.Scope.StreamID)
		meta := strings.Join([]string{
			fmt.Sprintf("Repo %s", m.repoNameByID(node.Scope.RepoID)),
			fmt.Sprintf("Visibility %s", stream.Visibility),
			fmt.Sprintf("ID %d", node.Scope.StreamID),
		}, "   ")
		body := strings.Join([]string{"Description", optionalText(stream.Description)}, "\n")
		return m.openViewEntityDialog("Stream", node.Name, meta, body), true
	case PaneIssues, PaneHabits:
		return m.openItemDetailDialog()
	}

	return m, false
}

func optionalText(text *string) string {
	if text == nil || strings.TrimSpace(*text) == "" {
		return "-"
	}
	return strings.TrimSpace(*text)
}

func (m Model) openSelectedDeleteDialog() (Model, bool) {
	if m.timer != nil && m.timer.State != "idle" {
		return m.withStatus("Stop the active session before deleting work items", true), true
	}
	switch m.pane {
	case PaneScope:
		node, ok := m.selectedScopeNode()
		if !ok || node.Scope.IsAll() {
			return m, false
		}
		if node.System {
			return m.withStatus(node.Name+" is made by Crona and cannot be deleted", true), true
		}
		if node.Scope.StreamID == 0 {
			return m.openConfirmDeleteEntity(
				"repo",
				fmt.Sprintf("%d", node.Scope.RepoID),
				node.Name,
			), true
		}
		next := m.openConfirmDeleteEntity(
			"stream",
			fmt.Sprintf("%d", node.Scope.StreamID),
			node.Name,
		)
		next.dialogRepoID = node.Scope.RepoID
		return next, true
	case PaneIssues:
		if issue, ok := m.selectedIssueRecord(); ok {
			next := m.openConfirmDeleteEntity("issue", fmt.Sprintf("%d", issue.ID), issue.Title)
			next.dialogStreamID = issue.StreamID
			return next, true
		}
	case PaneHabits:
		if m.view == ViewDaily {
			if habit, ok := m.selectedDailyHabitRecord(); ok {
				next := m.openConfirmDeleteEntity("habit", fmt.Sprintf("%d", habit.ID), habit.Name)
				next.dialogStreamID = habit.StreamID
				return next, true
			}
		}
		if habit, ok := m.selectedHabitRecord(); ok {
			next := m.openConfirmDeleteEntity("habit", fmt.Sprintf("%d", habit.ID), habit.Name)
			next.dialogStreamID = habit.StreamID
			return next, true
		}
	}
	return m, false
}

func (m Model) repoNameByID(repoID int64) string {
	for _, repo := range m.repos {
		if repo.ID == repoID {
			return repo.Name
		}
	}
	return ""
}

func (m Model) repoDescriptionByID(repoID int64) *string {
	for _, repo := range m.repos {
		if repo.ID == repoID {
			return repo.Description
		}
	}
	return nil
}

func (m Model) streamDescriptionByID(streamID int64) *string {
	for _, stream := range m.streams {
		if stream.ID == streamID {
			return stream.Description
		}
	}
	return nil
}

// currentDate is the shared date every dated view shows; unset, it follows the logical today.
func (m Model) currentDate() string {
	if m.date != "" {
		return m.date
	}
	return m.currentLogicalDate()
}

func (m Model) currentLogicalDate() string {
	if m.health != nil && strings.TrimSpace(m.health.CurrentDate) != "" {
		return strings.TrimSpace(m.health.CurrentDate)
	}
	return time.Now().Format("2006-01-02")
}

// reviewRange is Review's window: reviewWindowDays ending on its anchor, carried along
// when the shared date leaves it.
func (m Model) reviewRange() (string, string) {
	return uistate.ReviewRange(m.currentDate(), m.reviewWindowEnd, m.reviewWindowDays)
}

// reloadReview loads Review's whole window.
func (m Model) reloadReview() tea.Cmd {
	start, end := m.reviewRange()
	return commands.LoadReviewWindow(m.client, start, end)
}

// anchorReviewWindow pins Review's window where the shared date now leaves it.
func (m Model) anchorReviewWindow() Model {
	_, m.reviewWindowEnd = m.reviewRange()
	return m
}

// reloadDateSurfaces reloads every surface bound to the shared date.
func (m Model) reloadDateSurfaces() tea.Cmd {
	date := m.currentDate()
	return tea.Batch(
		commands.LoadDailySummary(m.client, date),
		commands.LoadDailyPlan(m.client, date),
		commands.LoadDueHabits(m.client, date),
		commands.LoadDailyStreaks(m.client, date),
		commands.LoadSummarySnapshot(m.client, date),
		commands.LoadMomentumRange(m.client, date, m.currentMomentumWindowDays()),
		commands.LoadWellbeingDay(m.client, date),
		m.reloadReview(),
	)
}

func (m Model) currentMomentumWindowDays() int {
	if m.momentumWindowDays < 1 {
		return 30
	}
	return m.momentumWindowDays
}

func (m Model) momentumCursorForView() int {
	return m.cursor[PaneMomentumCards]
}

func (m Model) currentWeekStart() sharedtypes.WeekStart {
	if m.settings == nil {
		return sharedtypes.WeekStartMonday
	}
	return sharedtypes.NormalizeWeekStart(m.settings.WeekStart)
}

// detailContentState is the content state the detail dialogs read penalties and momentums from.
func (m Model) detailContentState(snapshot selectionpkg.Snapshot) viewtypes.ContentState {
	return m.viewContentState(m.mainContentWidth(), m.height, snapshot, nil)
}

// openDetailDialog shows what the view's details pane shows for its selection, at the
// dialog's width; adjust sets anything the dialog's keys need.
func (m Model) openDetailDialog(
	kind string,
	detail func(viewtypes.Theme, viewtypes.ContentState, int) (string, []string, bool),
	snapshot selectionpkg.Snapshot,
	adjust func(*dialogstate.State),
) (Model, bool) {
	title, lines, ok := detail(layoutpkg.ViewTheme(), m.detailContentState(snapshot), viewEntityBodyWidth(m.width))
	if !ok {
		return m, false
	}
	dialog := m.dialogSnapshot().OpenViewEntity(kind, title, "", strings.Join(lines, "\n"))
	if adjust != nil {
		adjust(&dialog)
	}
	return m.withDialogState(dialog), true
}
