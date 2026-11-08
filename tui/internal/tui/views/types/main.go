// Package viewtypes defines themes, content state, and sizing types shared by renderers.
package viewtypes

import (
	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	"crona.work/tui/internal/tui/scope"
	uistate "crona.work/tui/internal/tui/state"

	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	ColorBlue      lipgloss.Color
	ColorCyan      lipgloss.Color
	ColorGreen     lipgloss.Color
	ColorDullGreen lipgloss.Color
	ColorMagenta   lipgloss.Color
	ColorSubtle    lipgloss.Color
	ColorYellow    lipgloss.Color
	ColorRed       lipgloss.Color
	ColorDullRed   lipgloss.Color
	ColorOrange    lipgloss.Color
	ColorLavender  lipgloss.Color
	ColorDim       lipgloss.Color
	ColorWhite     lipgloss.Color

	StyleActive          lipgloss.Style
	StyleInactive        lipgloss.Style
	StylePaneTitle       lipgloss.Style
	StyleDim             lipgloss.Style
	StyleCursor          lipgloss.Style
	StyleHeader          lipgloss.Style
	StyleError           lipgloss.Style
	StyleSelected        lipgloss.Style
	StyleSelectedInverse lipgloss.Style
	StyleNormal          lipgloss.Style
}

type ContentState struct {
	View               string
	Pane               string
	SelectedIssueAdhoc bool
	// BrowseScope is the Issues tab's scope tree selection; AllStreams and AllHabits feed it.
	BrowseScope         scope.Scope
	AllStreams          []api.Stream
	AllHabits           []api.HabitWithMeta
	ScopedHabits        []api.HabitWithMeta
	Width               int
	Height              int
	Cursors             map[string]int
	Filters             map[string]string
	Elapsed             int
	DashboardDate       string
	SummaryDate         string
	RollupStartDate     string
	RollupEndDate       string
	MomentumDate        string
	MomentumWindowDays  int
	WellbeingDate       string
	WellbeingWindowDays int
	// ReviewWindowDays days end on ReviewWindowEnd.
	ReviewWindowDays    int
	ReviewWindowEnd     string
	WeekStart           sharedtypes.WeekStart
	DefaultIssueSection string
	DailyTaskSection    string
	Tags                []api.Tag
	SessionHistoryTitle string
	SessionHistoryMeta  string
	HabitHistoryTitle   string
	HabitHistoryMeta    string
	RestModeActive      bool
	AwayModeActive      bool
	RestModeMessage     string
	RestModeDetail      string
	HistoricalAwayDate  string

	Repos       []api.Repo
	Streams     []api.Stream
	Issues      []api.Issue
	DailyIssues []api.Issue
	// DailyPlannedIssues is the Planned section whatever tab is shown; the strip counts it.
	DailyPlannedIssues []api.Issue
	// DailyDetailFocus focuses Daily's detail pane on the selection (enter); otherwise it
	// shows the Day summary. DailyDetailScroll is how far the focused pane is scrolled.
	DailyDetailFocus bool
	// DailyDetailSummary keeps the focused detail pane on the Day summary (tab into it).
	DailyDetailSummary bool
	DailyDetailScroll  int
	// DailyPlanSelect makes the Day summary's plan list selectable (P on the focused summary);
	// DailyPlanCursor is the selected row in summary.PlanItems' order.
	DailyPlanSelect bool
	DailyPlanCursor int
	// WorkBreakdown is the shown day's worked time by DailyWorkGroup (or the drill's level),
	// nil while loading. DailyWorkSelect makes its rows selectable at DailyWorkCursor.
	WorkBreakdown *api.TimeDistributionSummary
	// WorkColumns is WorkBreakdown per Review column (a day, or a week of a long window).
	WorkColumns     []*api.TimeDistributionSummary
	DailyWorkSelect bool
	DailyWorkCursor int
	DailyWorkGroup  string
	// MomentumBucket is the selected momentum's chosen bucket, counted back from its latest;
	// MomentumBucketDetail is that bucket's contributors, nil while loading.
	MomentumBucket       int
	MomentumBucketDetail *api.MomentumDetail
	DailyWorkPath        []uistate.WorkCrumb
	// DailyIssueIndices and DailyHabitIndices are the model's text-filtered raw indices into
	// DailyIssues and DueHabits; the pane cursors are positions in these lists.
	DailyIssueIndices     []int
	DailyHabitIndices     []int
	Habits                []api.Habit
	AllIssues             []api.IssueWithMeta
	DefaultIssues         []api.IssueWithMeta
	DueHabits             []api.HabitDailyItem
	MomentumCards         []api.MomentumCard
	DailySummary          *api.DailyIssueSummary
	SummarySnapshot       *api.SummarySnapshot
	DailyPlan             *api.DailyPlan
	DailyCheckIn          *api.DailyCheckIn
	MetricsRange          []api.DailyMetricsDay
	MetricsRollup         *api.MetricsRollup
	RollupMetricsRange    []api.DailyMetricsDay
	RollupMetricsRollup   *api.MetricsRollup
	MomentumMetricsRange  []api.DailyMetricsDay
	MomentumMetricsRollup *api.MetricsRollup
	Streaks               *api.StreakSummary
	DailyStreaks          *api.StreakSummary
	DashboardWindow       *api.DashboardWindowSummary
	DailyFocusScore       *api.FocusScoreSummary
	WeeklyFocusScore      *api.FocusScoreSummary
	// ReviewScores are the window's focus scores by day; ReviewPrev is the window of the
	// same length just before it, for Review's comparison column.
	ReviewScores  []api.FocusScoreRangeDay
	ReviewPrev    *ReviewPrev
	GoalProgress  *api.GoalProgressSummary
	ExportAssets  *api.ExportAssetStatus
	ExportReports []api.ExportReportFile
	IssueSessions []api.Session
	// DetailSessions are the sessions of DetailSessionsIssueID, the issue Daily's detail pane
	// shows.
	DetailSessions        []api.SessionHistoryEntry
	DetailSessionsIssueID int64
	// DetailHabitHistory is the history of DetailHistoryHabitID, the habit the pane shows.
	DetailHabitHistory   []api.HabitCompletion
	DetailHistoryHabitID int64
	SessionHistory       []api.SessionHistoryEntry
	// SessionPreview is the selected session's loaded detail; ignore it when its ID differs.
	SessionPreview            *api.SessionDetail
	SessionPreviewFailed      bool
	HabitHistory              []api.HabitCompletion
	Ops                       []api.Op
	Context                   *api.ActiveContext
	Timer                     *api.TimerState
	Health                    *api.Health
	AlertStatus               *api.AlertStatus
	AlertReminders            []api.AlertReminder
	UpdateStatus              *api.UpdateStatus
	UpdateChecking            bool
	UpdateInstallAvailable    bool
	UpdateDiagnosticsExpanded bool
	SupportCursor             int
	UpdateCommand             string
	UpdateManualReason        string
	TUIExecutablePath         string
	KernelExecutablePath      string
	KernelInfo                *api.KernelInfo
	Settings                  *api.CoreSettings
}

func (s ContentState) AwayDates() []string {
	if s.Settings == nil {
		return nil
	}
	return s.Settings.AwayDates
}

// ReviewPrev is the window before Review's: its totals, plan and focus score.
type ReviewPrev struct {
	Rollup *api.MetricsRollup
	Window *api.DashboardWindowSummary
	Focus  *api.FocusScoreSummary
}
