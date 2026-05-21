package mcp

import "errors"

// errValidation is returned when a required field is missing or invalid.
// Wrapping it via fmt.Errorf("...: %w", errValidation(...)) lets the caller
// (mcp tool handler) surface a clean message back to the client; the SDK
// turns the returned error into a tool error response.
type validationError struct{ msg string }

func (e validationError) Error() string { return e.msg }

func errValidation(msg string) error {
	return validationError{msg: msg}
}

// IsValidation reports whether err originates from errValidation. Reserved
// for tests; the runtime treats every tool error the same way.
func IsValidation(err error) bool {
	var v validationError
	return errors.As(err, &v)
}
