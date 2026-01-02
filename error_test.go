package command_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"lesiw.io/command"
)

func TestErrorStringFormat(t *testing.T) {
	tests := []struct {
		name string
		err  *command.Error
		want string
	}{{
		name: "code only",
		err:  &command.Error{Code: 1},
		want: "exit status 1",
	}, {
		name: "code with underlying error",
		err: &command.Error{
			Code: 127,
			Err:  fmt.Errorf("command not found"),
		},
		want: "command not found",
	}, {
		name: "code with log",
		err:  &command.Error{Code: 1, Log: []byte("permission denied\n")},
		want: strings.TrimSpace(`
exit status 1
	permission denied
`),
	}, {
		name: "error with multiline log",
		err: &command.Error{
			Err:  fmt.Errorf("failed"),
			Code: 2,
			Log:  []byte("line 1\nline 2\n"),
		},
		want: strings.TrimSpace(`
failed
	line 1
	line 2
`),
	}, {
		name: "empty log ignored",
		err:  &command.Error{Code: 1, Log: []byte{}},
		want: "exit status 1",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorUnwrap(t *testing.T) {
	cmdErr := &command.Error{
		Err:  fmt.Errorf("underlying error"),
		Code: 1,
	}
	if got, want := errors.Is(cmdErr, cmdErr.Err), true; got != want {
		t.Errorf("errors.Is(cmdErr, cmdErr.Err) = %v, want %v", got, want)
	}
	if got, want := errors.Unwrap(cmdErr), cmdErr.Err; got != want {
		t.Errorf("errors.Unwrap(cmdErr) = %v, want %v", got, want)
	}
}

func TestNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{{
		name: "nil error",
		err:  nil,
		want: false,
	}, {
		name: "non-command error",
		err:  fmt.Errorf("some error"),
		want: false,
	}, {
		name: "command error with code but no err",
		err:  &command.Error{Code: 127},
		want: false,
	}, {
		name: "command error with err but non-zero code",
		err:  &command.Error{Err: fmt.Errorf("failed"), Code: 1},
		want: false,
	}, {
		name: "command not found (error with exit code unset)",
		err:  &command.Error{Err: fmt.Errorf("command not found")},
		want: true,
	}, {
		name: "wrapped command not found",
		err: fmt.Errorf("wrapped: %w",
			&command.Error{Err: fmt.Errorf("not found"), Code: 0},
		),
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := command.NotFound(tt.err); got != tt.want {
				t.Errorf("NotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}
