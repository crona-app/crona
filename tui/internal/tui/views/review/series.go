package review

import (
	"cmp"
	"fmt"
	"math"
	"strconv"
	"time"

	"crona.work/tui/internal/api"
	helperpkg "crona.work/tui/internal/tui/helpers"
	uistate "crona.work/tui/internal/tui/state"
	types "crona.work/tui/internal/tui/views/types"

	sharedtypes "crona.work/shared/types"

	"github.com/charmbracelet/lipgloss"
)

// daySeries is the window's values by day, oldest first; NaN where a day has none.
type daySeries struct {
	days                               []time.Time
	mood, energy, sleep, worked, focus []float64
	done, planned                      []int
	selected                           int
	long                               bool
}

func collectSeries(state types.ContentState) daySeries {
	s := daySeries{selected: -1}
	end := cmp.Or(state.ReviewWindowEnd, state.DashboardDate)
	start, end := uistate.ReviewRange(end, end, state.ReviewWindowDays)
	first, err := time.Parse(time.DateOnly, start)
	if err != nil {
		return s
	}
	last, err := time.Parse(time.DateOnly, end)
	if err != nil {
		return s
	}
	metrics := map[string]api.DailyMetricsDay{}
	for _, day := range state.MetricsRange {
		metrics[day.Date] = day
	}
	plans := map[string]sharedtypes.DashboardWindowDay{}
	if state.DashboardWindow != nil {
		for _, day := range state.DashboardWindow.Days {
			plans[day.Date] = day
		}
	}
	scores := map[string]api.FocusScoreRangeDay{}
	for _, day := range state.ReviewScores {
		scores[day.Date] = day
	}
	nan := math.NaN()
	for t := first; !t.After(last); t = t.AddDate(0, 0, 1) {
		date := t.Format(time.DateOnly)
		if date == state.DashboardDate {
			s.selected = len(s.days)
		}
		day := metrics[date]
		mood, energy, sleep := nan, nan, nan
		if c := day.CheckIn; c != nil {
			mood, energy = float64(c.Mood), float64(c.Energy)
			if c.SleepHours != nil {
				sleep = *c.SleepHours
			}
		}
		focus := nan
		if score := scores[date]; score.HasData {
			focus = float64(score.Score)
		}
		s.days = append(s.days, t)
		s.mood, s.energy, s.sleep = append(s.mood, mood), append(s.energy, energy), append(s.sleep, sleep)
		s.focus = append(s.focus, focus)
		s.worked = append(s.worked, hours(day.WorkedSeconds))
		s.done = append(s.done, plans[date].CompletedCount)
		s.planned = append(s.planned, plans[date].PlannedCount)
	}
	s.long = uistate.LongReviewWindow(len(s.days))
	return s
}

// metric is how one series looks wherever it shows: tile, chart, legend and readout.
type metric struct {
	name   string
	color  lipgloss.Color
	format func(float64) string
}

func (m metric) title() string { return string(m.name[0]-'a'+'A') + m.name[1:] }

type metricSet struct{ mood, energy, sleep, worked, focus metric }

func metricsFor(theme types.Theme) metricSet {
	return metricSet{
		mood:   metric{"mood", theme.ColorMagenta, rating},
		energy: metric{"energy", theme.ColorYellow, rating},
		sleep:  metric{"sleep", theme.ColorBlue, helperpkg.FormatCompactDurationHours},
		worked: metric{"worked", theme.ColorCyan, helperpkg.FormatCompactDurationHours},
		focus:  metric{"focus", theme.ColorLavender, whole},
	}
}

// rating shows a 1–5 value to one decimal, whole when it is: a day's 4, a window's 3.8.
func rating(v float64) string { return strconv.FormatFloat(math.Round(v*10)/10, 'f', -1, 64) }

func whole(v float64) string { return fmt.Sprintf("%.0f", v) }

func hours(seconds int) float64 { return float64(seconds) / 3600 }

func planCount(done, planned int) string { return fmt.Sprintf("%d/%d", done, planned) }

// mean is the average of the values that are set, NaN when none are.
func mean(values []float64) float64 {
	sum, n := 0.0, 0
	for _, v := range values {
		if !math.IsNaN(v) {
			sum, n = sum+v, n+1
		}
	}
	if n == 0 {
		return math.NaN()
	}
	return sum / float64(n)
}

// rolling averages each day with the radius days either side of it, gaps left out.
func rolling(values []float64, radius int) []float64 {
	out := make([]float64, len(values))
	for i := range values {
		out[i] = mean(values[max(0, i-radius):min(len(values), i+radius+1)])
	}
	return out
}
