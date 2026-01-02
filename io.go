package command

import (
	"errors"
	"io"
	"os"

	"lesiw.io/prefix"
)

var (
	Trace   = io.Discard
	ShTrace = prefix.NewWriter("+ ", stderr)

	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

// ErrClosed is returned when attempting to read from or write to a closed
// reader or writer.
var ErrClosed = errors.New("command: write to closed buffer")

// ErrReadOnly is returned when attempting to write to a read-only command.
var ErrReadOnly = errors.New("command: write to read-only buffer")
