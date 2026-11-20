//go:build e2e

package e2e

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	shareddto "crona.work/shared/dto"
	"crona.work/shared/protocol"
	sharedtypes "crona.work/shared/types"
)

func TestFocusStartRequiresEligibleIssueStatusAndSessionEndDoesNotCompleteIssue(t *testing.T) {
	kernel := startTestKernel(t)
	defer kernel.close(t)

	repo := createRepo(t, kernel, "Repo")
	stream := createStream(t, kernel, repo.ID, "main")
	issue := createIssue(t, kernel, stream.ID, "Lifecycle issue", nil)

	var ctx sharedtypes.ActiveContext
	kernel.call(t, protocol.MethodContextSet, map[string]any{
		"repoId":   repo.ID,
		"streamId": stream.ID,
		"issueId":  issue.ID,
	}, &ctx)

	focusDenied := kernel.callErrorResponse(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{})
	if focusDenied.Code != protocol.CodeFocusStatusInvalid {
		t.Fatalf("expected code %q, got %q", protocol.CodeFocusStatusInvalid, focusDenied.Code)
	}
	if !strings.Contains(focusDenied.Message, "focus sessions cannot be started") {
		t.Fatalf("expected focus eligibility error, got %q", focusDenied.Message)
	}
	var focusDeniedData struct {
		Status    string   `json:"status"`
		Focusable []string `json:"focusable"`
	}
	if err := json.Unmarshal(focusDenied.Data, &focusDeniedData); err != nil {
		t.Fatalf("decode focus error data: %v", err)
	}
	if focusDeniedData.Status != string(sharedtypes.IssueStatusBacklog) {
		t.Fatalf("expected backlog status, got %q", focusDeniedData.Status)
	}
	if !slices.Contains(focusDeniedData.Focusable, string(sharedtypes.IssueStatusPlanned)) {
		t.Fatalf("expected planned in focusable statuses, got %v", focusDeniedData.Focusable)
	}

	changeIssueStatus(t, kernel, issue.ID, sharedtypes.IssueStatusPlanned)

	var timer sharedtypes.TimerState
	kernel.call(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{}, &timer)
	if timer.State != "running" {
		t.Fatalf("expected running timer, got %+v", timer)
	}

	errMessage := kernel.callError(t, protocol.MethodIssueChangeStatus, shareddto.ChangeIssueStatusRequest{
		ID:     issue.ID,
		Status: sharedtypes.IssueStatusDone,
	})
	if !strings.Contains(errMessage, "cannot change issue status while a focus session is active") {
		t.Fatalf("expected active-session status change denial, got %q", errMessage)
	}

	issues := listIssues(t, kernel, stream.ID)
	if len(issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(issues))
	}
	if issues[0].Status != sharedtypes.IssueStatusInProgress {
		t.Fatalf("expected issue to auto-move to in_progress, got %s", issues[0].Status)
	}

	commitMessage := new(string)
	*commitMessage = "session ended"
	outcome := new(string)
	*outcome = "Kernel lifecycle migrated"
	nextStep := new(string)
	*nextStep = "Polish footer and CRUD"
	kernel.call(t, protocol.MethodTimerEnd, shareddto.EndSessionRequest{
		CommitMessage: commitMessage,
		Outcome:       outcome,
		NextStep:      nextStep,
	}, &timer)
	if timer.State != "idle" {
		t.Fatalf("expected idle timer after end, got %+v", timer)
	}

	issues = listIssues(t, kernel, stream.ID)
	if issues[0].Status != sharedtypes.IssueStatusInProgress {
		t.Fatalf("expected issue to remain in_progress after session end, got %s", issues[0].Status)
	}

	var history []sharedtypes.SessionHistoryEntry
	kernel.call(t, protocol.MethodSessionHistory, shareddto.SessionHistoryQuery{}, &history)
	if len(history) == 0 {
		t.Fatalf("expected session history entry after timer end")
	}
	notes := history[0].ParsedNotes[sharedtypes.SessionNoteSectionNotes]
	if !strings.Contains(notes, "Outcome: Kernel lifecycle migrated") {
		t.Fatalf("expected structured outcome note, got %q", notes)
	}
	if !strings.Contains(notes, "Next step: Polish footer and CRUD") {
		t.Fatalf("expected structured next-step note, got %q", notes)
	}

	changeIssueStatus(t, kernel, issue.ID, sharedtypes.IssueStatusDone)
	doneDenied := kernel.callErrorResponse(t, protocol.MethodTimerStart, shareddto.TimerStartRequest{})
	if doneDenied.Code != protocol.CodeFocusStatusInvalid {
		t.Fatalf("expected code %q for a done issue, got %q", protocol.CodeFocusStatusInvalid, doneDenied.Code)
	}
	var doneDeniedData struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(doneDenied.Data, &doneDeniedData); err != nil {
		t.Fatalf("decode done focus error data: %v", err)
	}
	if doneDeniedData.Status != string(sharedtypes.IssueStatusDone) {
		t.Fatalf("expected done status, got %q", doneDeniedData.Status)
	}

	issue2 := createIssue(t, kernel, stream.ID, "Blocked lifecycle issue", nil)
	changeIssueStatus(t, kernel, issue2.ID, sharedtypes.IssueStatusPlanned)
	errMessage = kernel.callError(t, protocol.MethodIssueChangeStatus, shareddto.ChangeIssueStatusRequest{
		ID:     issue2.ID,
		Status: sharedtypes.IssueStatusAbandoned,
	})
	if !strings.Contains(errMessage, "requires a reason") {
		t.Fatalf("expected abandoned reason validation, got %q", errMessage)
	}
}
