package types

import "testing"

func TestFocusRecoveryStatus(t *testing.T) {
	tests := []struct {
		name       string
		status     IssueStatus
		wantStatus IssueStatus
		wantOK     bool
	}{
		{"backlog needs ready", IssueStatusBacklog, IssueStatusReady, true},
		{"blocked needs ready", IssueStatusBlocked, IssueStatusReady, true},
		{"in review needs in progress", IssueStatusInReview, IssueStatusInProgress, true},
		{"done needs planned", IssueStatusDone, IssueStatusPlanned, true},
		{"abandoned needs planned", IssueStatusAbandoned, IssueStatusPlanned, true},
		{"todo alias needs ready", "todo", IssueStatusReady, true},
		{"planned is focusable", IssueStatusPlanned, IssueStatusPlanned, false},
		{"ready is focusable", IssueStatusReady, IssueStatusReady, false},
		{"in progress is focusable", IssueStatusInProgress, IssueStatusInProgress, false},
		{"active alias is focusable", "active", IssueStatusInProgress, false},
		{"unknown status has no recovery", IssueStatus("unknown"), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotOK := FocusRecoveryStatus(tt.status)
			if gotStatus != tt.wantStatus || gotOK != tt.wantOK {
				t.Fatalf(
					"FocusRecoveryStatus(%q) = (%q, %t), want (%q, %t)",
					tt.status,
					gotStatus,
					gotOK,
					tt.wantStatus,
					tt.wantOK,
				)
			}
		})
	}
}
