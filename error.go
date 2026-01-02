package command

import (
	"errors"
	"fmt"
	"strings"
)

// Error represents a command execution failure.
//
// Commands attached to their controlling terminal via Exec will have an
// empty Log, since stderr is attached directly to the terminal rather than
// being captured.
type Error struct {
	// Log contains the log output. This usually corresponds to stderr.
	Log []byte

	// Err is the underlying error.
	Err error

	// Code is the exit code. A value of 0 does not indicate success.
	Code int
}

func (e *Error) Error() string {
	var sb strings.Builder
	if e.Err != nil {
		sb.WriteString(e.Err.Error())
	} else {
		sb.WriteString(fmt.Sprintf("exit status %d", e.Code))
	}
	if len(e.Log) > 0 {
		sb.WriteString(
			"\n\t" +
				strings.TrimSuffix(
					strings.ReplaceAll(string(e.Log), "\n", "\n\t"),
					"\n\t",
				),
		)
	}
	return sb.String()
}

func (e *Error) Unwrap() error { return e.Err }

// NotFound returns true if err represents a command that failed to start,
// typically indicating the command was not found.
//
// A command.Error is considered "not found" when Err is non-nil and Code is 0.
// This combination means the command never ran (failed to start).
//
// NotFound uses errors.As to probe the error chain for a command.Error.
// If no command.Error exists in the chain, NotFound returns false.
//
// Example:
//
//	_, err := command.Read(ctx, sh, "nonexistent")
//	if command.NotFound(err) {
//	    // Command wasn't found or failed to start
//	}
func NotFound(err error) bool {
	var cmdErr *Error
	if !errors.As(err, &cmdErr) {
		return false
	}
	return cmdErr.Err != nil && cmdErr.Code == 0
}
