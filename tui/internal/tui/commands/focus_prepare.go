package commands

import (
	"crona.work/tui/internal/api"
	tea "github.com/charmbracelet/bubbletea"
)

// PrepareIssueForFocus changes lifecycle status before a focus dialog is opened.
func PrepareIssueForFocus(c *api.Client, issueID int64, status string) tea.Cmd {
	return func() tea.Msg {
		if err := c.ChangeIssueStatus(issueID, status, nil); err != nil {
			return ErrMsg{Err: err, Operation: "prepare issue for focus"}
		}
		return IssueReadyForFocusMsg{IssueID: issueID}
	}
}
