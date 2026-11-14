package summary

import (
	"fmt"
	"strings"
	"time"

	sharedtypes "crona.work/shared/types"
	"crona.work/tui/internal/api"
	types "crona.work/tui/internal/tui/views/types"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func renderRecentUpdates(theme types.Theme, state types.ContentState, width int) []string {
	snapshot := state.SummarySnapshot
	lines := []string{theme.StylePaneTitle.Render("Recent updates")}
	if snapshot.RecentUpdatesUnavailable {
		return append(lines, theme.StyleDim.Render("Updates unavailable."))
	}
	if len(snapshot.RecentUpdates) == 0 {
		return append(lines, theme.StyleDim.Render("No issue updates this day."))
	}
	titles := map[string]string{}
	for id, title := range snapshot.UpdateTitles {
		titles[id] = title
	}
	if snapshot.Issues != nil {
		for _, issue := range snapshot.Issues.Issues {
			titles[fmt.Sprint(issue.ID)] = issue.Title
		}
	}
	for _, issue := range state.AllIssues {
		if titles[fmt.Sprint(issue.ID)] == "" {
			titles[fmt.Sprint(issue.ID)] = issue.Title
		}
	}
	for _, op := range snapshot.RecentUpdates[:min(7, len(snapshot.RecentUpdates))] {
		at, err := time.Parse(time.RFC3339Nano, op.Timestamp)
		clock := "--:--"
		if err == nil {
			clock = at.Format("15:04")
		}
		label := "#" + op.EntityID
		if title := titles[op.EntityID]; title != "" {
			label += " " + strings.Join(strings.Fields(title), " ")
		}
		row := theme.StyleDim.Render(clock+"  ") +
			issueUpdateStyle(theme, op).Render(issueUpdateAction(op)) +
			theme.StyleDim.Render(" · ") + theme.StyleNormal.Render(label)
		lines = append(lines, ansi.Truncate(row, max(1, width), "…"))
	}
	return lines
}

func issueUpdateAction(op api.Op) string {
	switch op.Action {
	case sharedtypes.OpActionCreate:
		return "Added"
	case sharedtypes.OpActionDelete:
		return "Deleted"
	case sharedtypes.OpActionRestore:
		return "Restored"
	}
	payload, _ := op.Payload.(map[string]any)
	if to, ok := payload["to"].(string); ok && to != "" {
		if from, ok := payload["from"].(string); ok && from != "" {
			return from + " → " + to
		}
		return "Status → " + to
	}
	if date, exists := payload["todoForDate"]; exists {
		if date == nil || date == "" {
			return "Unscheduled"
		}
		return fmt.Sprint("Scheduled ", date)
	}
	if status, ok := payload["status"].(string); ok && status != "" {
		return "Status → " + status
	}
	return "Updated"
}

// issueUpdateStyle uses action colours, with status changes matching the Plan legend.
func issueUpdateStyle(theme types.Theme, op api.Op) lipgloss.Style {
	color := theme.ColorYellow
	switch op.Action {
	case sharedtypes.OpActionCreate, sharedtypes.OpActionRestore:
		color = theme.ColorGreen
	case sharedtypes.OpActionDelete:
		color = theme.ColorRed
	default:
		payload, _ := op.Payload.(map[string]any)
		status, _ := payload["to"].(string)
		if status == "" {
			if _, scheduled := payload["todoForDate"]; scheduled {
				return lipgloss.NewStyle().Foreground(theme.ColorBlue)
			}
			status, _ = payload["status"].(string)
		}
		for _, item := range issueStatuses(theme) {
			if item.status == status {
				color = item.color
				break
			}
		}
	}
	return lipgloss.NewStyle().Foreground(color)
}
