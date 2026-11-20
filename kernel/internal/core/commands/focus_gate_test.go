package commands

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"crona.work/shared/protocol"
	sharedtypes "crona.work/shared/types"
)

func TestFocusStatusErrorReportsProtocolDetails(t *testing.T) {
	gate := &FocusStatusError{IssueID: 22, Status: string(sharedtypes.IssueStatusBacklog)}

	message := gate.Error()
	for _, want := range []string{
		"issue #22",
		"backlog",
		"focus sessions cannot be started",
		"planned, ready, or in progress",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("message %q is missing %q", message, want)
		}
	}

	if got := gate.ProtocolErrorCode(); got != protocol.CodeFocusStatusInvalid {
		t.Fatalf("code = %q, want %q", got, protocol.CodeFocusStatusInvalid)
	}

	body, err := json.Marshal(gate.ProtocolErrorData())
	if err != nil {
		t.Fatalf("marshal data: %v", err)
	}
	var data struct {
		Status    string   `json:"status"`
		Focusable []string `json:"focusable"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if data.Status != "backlog" {
		t.Fatalf("status = %q, want backlog", data.Status)
	}
	wantFocusable := []string{"planned", "ready", "in_progress"}
	if !slices.Equal(data.Focusable, wantFocusable) {
		t.Fatalf("focusable = %v, want %v", data.Focusable, wantFocusable)
	}
}

func TestTimerStartRejectsBacklogIssueWithTypedError(t *testing.T) {
	ctx := context.Background()
	_, service, issue := newTimerTestContext(t, func() string { return "2026-05-24T10:00:00Z" })

	_, err := service.Start(ctx, nil, new(issue.StreamID), new(issue.ID), nil)

	var gate *FocusStatusError
	if !errors.As(err, &gate) {
		t.Fatalf("expected *FocusStatusError, got %T: %v", err, err)
	}
	if gate.IssueID != issue.ID {
		t.Fatalf("issue id = %d, want %d", gate.IssueID, issue.ID)
	}
	if gate.Status != string(sharedtypes.IssueStatusBacklog) {
		t.Fatalf("status = %q, want backlog", gate.Status)
	}
}

func TestManualSessionRejectsBacklogIssueWithTypedError(t *testing.T) {
	ctx := context.Background()
	coreCtx, _, issue := newTimerTestContext(t, func() string { return "2026-05-24T10:00:00Z" })

	_, err := LogManualSession(ctx, coreCtx, ManualSessionInput{
		IssueID:             issue.ID,
		Date:                "2026-05-24",
		WorkDurationSeconds: 600,
	})

	var gate *FocusStatusError
	if !errors.As(err, &gate) {
		t.Fatalf("expected *FocusStatusError, got %T: %v", err, err)
	}
	if gate.IssueID != issue.ID {
		t.Fatalf("issue id = %d, want %d", gate.IssueID, issue.ID)
	}
}
