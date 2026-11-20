package protocol

// Canonical machine-readable error codes returned in Response.Error.Code.
// Clients must fall back to the message for codes they do not know.
const (
	// CodeFocusStatusInvalid reports that an issue's lifecycle status does not
	// allow a focus session. Error data carries the current status and the
	// statuses that can start focus.
	CodeFocusStatusInvalid = "focus_status_invalid"
)
