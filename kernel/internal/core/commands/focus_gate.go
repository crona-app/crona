package commands

import (
	"fmt"

	"crona.work/shared/protocol"
	sharedtypes "crona.work/shared/types"
)

// FocusStatusError reports that an issue's lifecycle status does not allow a
// focus session. The protocol code and data let clients offer a recovery path
// instead of only printing the message.
type FocusStatusError struct {
	IssueID int64
	Status  string
}

func (e *FocusStatusError) Error() string {
	return fmt.Sprintf(
		"issue #%d is %s; focus sessions cannot be started until it is planned, ready, or in progress",
		e.IssueID,
		e.Status,
	)
}

// ProtocolErrorCode implements the kernel's coded-error interface.
func (e *FocusStatusError) ProtocolErrorCode() string {
	return protocol.CodeFocusStatusInvalid
}

// ProtocolErrorData implements the kernel's structured-error interface.
func (e *FocusStatusError) ProtocolErrorData() any {
	return focusStatusErrorData{
		Status:    e.Status,
		Focusable: focusableStatuses(),
	}
}

type focusStatusErrorData struct {
	Status    string   `json:"status"`
	Focusable []string `json:"focusable"`
}

func focusableStatuses() []string {
	return []string{
		string(sharedtypes.IssueStatusPlanned),
		string(sharedtypes.IssueStatusReady),
		string(sharedtypes.IssueStatusInProgress),
	}
}
