package types

// FocusRecoveryStatus returns the status a client can set to make an issue
// focusable again, and whether a change is needed at all. Focusable statuses
// report themselves with false so callers can branch on a single call.
func FocusRecoveryStatus(status IssueStatus) (IssueStatus, bool) {
	if CanStartFocus(status) {
		return NormalizeIssueStatus(status), false
	}
	switch NormalizeIssueStatus(status) {
	case IssueStatusBacklog, IssueStatusBlocked:
		return IssueStatusReady, true
	case IssueStatusInReview:
		return IssueStatusInProgress, true
	case IssueStatusDone, IssueStatusAbandoned:
		return IssueStatusPlanned, true
	default:
		return "", false
	}
}
