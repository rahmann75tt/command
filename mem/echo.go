package mem

import (
	"context"
	"strings"

	"lesiw.io/command"
	"lesiw.io/command/internal/sh"
)

func echoCommand(ctx context.Context, args ...string) command.Buffer {
	return struct {
		*strings.Reader
		sh.Stringer
	}{
		strings.NewReader(strings.Join(args[1:], " ") + "\n"),
		sh.String(command.Envs(ctx), args...),
	}
}
