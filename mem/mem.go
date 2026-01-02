// Package mem provides an in-memory command.Machine for tests and examples.
//
// The machine provides real implementations of common commands (echo, cat,
// tee, tr) operating on an in-memory filesystem.
//
// # Guarantees
//
// mem.Machine() makes the following guarantees for consistent testing:
//
//   - Filesystem starts empty (no files or directories)
//   - Default environment: HOME=/
//   - Platform-independent behavior on all hosts
//
// These guarantees mean tests using mem.Machine() work identically on
// Windows, macOS, and Linux without conditional logic.
package mem

import (
	"context"
	"fmt"

	"lesiw.io/command"
	"lesiw.io/fs"
	"lesiw.io/fs/memfs"
)

// Machine returns a new in-memory command machine.
func Machine() command.Machine { return &machine{memfs.New()} }

type fsys = fs.FS
type machine struct{ fsys }

func (m *machine) FS() fs.FS                   { return m.fsys }
func (m *machine) OS(context.Context) string   { return "linux" }
func (m *machine) Arch(context.Context) string { return "amd64" }

func (m *machine) Command(ctx context.Context, arg ...string) command.Buffer {
	if len(arg) == 0 {
		return command.Fail(&command.Error{
			Err: fmt.Errorf("bad command: no command given"),
		})
	}
	switch arg[0] {
	case "cat":
		return catCommand(ctx, m, arg...)
	case "echo":
		return echoCommand(ctx, arg...)
	case "tee":
		return teeCommand(ctx, m, arg...)
	case "tr":
		return trCommand(ctx, arg...)
	default:
		return command.Fail(&command.Error{
			Err: fmt.Errorf("command not found: %s", arg[0]),
		})
	}
}
