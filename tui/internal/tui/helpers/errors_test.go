package helpers

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"

	"crona.work/shared/protocol"
)

func TestErrorMessageUsesRPCMessage(t *testing.T) {
	err := fmt.Errorf("saving issue: %w", &protocol.RPCError{Code: "request_failed", Message: "Issue title is required"})
	if got := ErrorMessage(err); got != "Issue title is required" {
		t.Fatalf("ErrorMessage()=%q, want human-readable RPC message", got)
	}
	if !strings.Contains(err.Error(), "request_failed") {
		t.Fatal("original diagnostic error lost code")
	}
}

func TestConnectionErrorExplainsRecovery(t *testing.T) {
	err := &net.OpError{Op: "dial", Net: "unix", Err: errors.New("no such file")}
	got := ErrorMessage(err)
	if !strings.Contains(got, "daemon") || !strings.Contains(got, "running") {
		t.Fatalf("no connection recovery guidance: %q", got)
	}
}

func TestLostReplyWarnsBeforeRetrying(t *testing.T) {
	got := ErrorMessage(fmt.Errorf("read response: %w", io.EOF))
	if !strings.Contains(got, "may have been saved") {
		t.Fatalf("lost reply incorrectly implies no change occurred: %q", got)
	}
}
