// Package sub implements a command.Machine that prefixes all commands
// with a fixed set of arguments.
package sub

import (
	"context"

	"lesiw.io/command"
)

// Machine returns a command.Machine that prefixes all commands with the given
// prefix arguments, using the provided machine for execution.
func Machine(m command.Machine, prefix ...string) command.Machine {
	return &machine{m: m, prefix: prefix}
}

type machine struct {
	m      command.Machine
	prefix []string
}

func (m *machine) Command(ctx context.Context, arg ...string) command.Buffer {
	return m.m.Command(ctx, append(m.prefix, arg...)...)
}
