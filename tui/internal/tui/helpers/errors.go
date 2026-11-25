package helpers

import (
	"errors"
	"io"
	"net"

	"crona.work/shared/protocol"
)

// ErrorMessage keeps protocol codes in diagnostics rather than user-facing copy.
func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	var rpcErr *protocol.RPCError
	if errors.As(err, &rpcErr) {
		return rpcErr.Message
	}
	var connectionErr *net.OpError
	if errors.As(err, &connectionErr) {
		if connectionErr.Op == "dial" {
			return "Cannot connect to the Crona daemon. Check that it is running, then try again."
		}
		return "Lost connection to the Crona daemon. Check its status before retrying; the last change may have been saved."
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return "Lost connection to the Crona daemon. Check its status before retrying; the last change may have been saved."
	}
	return err.Error()
}
