package command

import (
	"context"
	"fmt"
	"sync"

	"lesiw.io/fs"
)

// FS returns a filesystem (lesiw.io/fs.FS) that executes commands on m to
// perform filesystem operations.
//
// If m implements FSMachine, FS() calls m.FS(ctx) and uses the returned
// filesystem. If m.FS(ctx) returns nil, FS falls back to creating a
// command-based filesystem.
func FS(m Machine) fs.FS {
	if fsm, ok := m.(FSMachine); ok {
		if fsys := fsm.FS(); fsys != nil {
			return fsys
		}
	}
	return &cmdFS{Machine: m}
}

type cfsKind int

const (
	kindUnknown cfsKind = iota
	kindGNU
	kindBSD
	kindWindows
	kindDOS
)

type cmdFS struct {
	Machine // After init(), the filesystem-capable machine.

	once    sync.Once
	initErr error

	// Capabilities discovered after init().
	kind   cfsKind
	hasTar bool
}

func (cfs *cmdFS) init(ctx context.Context) error {
	cfs.once.Do(func() { cfs.initErr = cfs.doInit(ctx) })
	return cfs.initErr
}

func (cfs *cmdFS) doInit(ctx context.Context) error {
	switch OS(ctx, cfs.Machine) {
	case "linux":
		cfs.kind = kindGNU
	case "darwin", "freebsd", "openbsd", "netbsd", "dragonfly":
		cfs.kind = kindBSD
	case "windows":
		if err := psDo(ctx, cfs.Machine, "exit 0"); err == nil {
			cfs.kind = kindWindows
		} else {
			cfs.kind = kindDOS
		}
	default:
		return fmt.Errorf("failed to detect OS type")
	}

	// Use a simple command to probe for a filesystem-capable Machine.
	var args []string
	switch cfs.kind {
	case kindGNU, kindBSD:
		args = []string{"cat", "/dev/null"}
	case kindWindows, kindDOS:
		args = []string{"cmd", "/c", "type", "NUL"}
	default:
		panic("unreachable")
	}

	m := cfs.Machine
	for {
		if err := Do(ctx, m, args...); err == nil || !NotFound(err) {
			break // Found a machine that can execute filesystem commands.
		}
		if u, ok := m.(Unsheller); ok {
			if unshelled := u.Unshell(); unshelled != nil {
				m = unshelled
				continue
			}
		}
		return fmt.Errorf("no filesystem-capable machine found")
	}
	cfs.Machine = m

	cfs.hasTar = !NotFound(Do(ctx, cfs, "tar"))
	return nil
}

var errUnsupportedOS = fmt.Errorf("unknown OS")
