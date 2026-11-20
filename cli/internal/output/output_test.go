package output

import (
	"errors"
	"testing"

	"crona.work/shared/protocol"
)

func TestErrorMessageOmitsRPCCode(t *testing.T) {
	err := &protocol.RPCError{
		Code:    protocol.CodeFocusStatusInvalid,
		Message: "issue #22 is backlog; focus sessions cannot be started until it is planned, ready, or in progress",
	}

	got := ErrorMessage(err)
	if got != err.Message {
		t.Fatalf("ErrorMessage() = %q, want %q", got, err.Message)
	}
}

func TestErrorMessagePassesThroughOthers(t *testing.T) {
	err := errors.New("local validation failed")

	if got := ErrorMessage(err); got != err.Error() {
		t.Fatalf("ErrorMessage() = %q, want %q", got, err.Error())
	}
}
