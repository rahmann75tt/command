package command

import (
	"context"

	"lesiw.io/fs"
)

var _ fs.MkdirFS = (*cmdFS)(nil)

func (cfs *cmdFS) Mkdir(ctx context.Context, name string) (err error) {
	if err = cfs.init(ctx); err != nil {
		return
	}

	switch cfs.kind {
	case kindGNU, kindBSD:
		err = Do(ctx, cfs, "mkdir", name)
	case kindWindows, kindDOS:
		err = Do(ctx, cfs, "cmd", "/c", "mkdir", name)
	default:
		err = errUnsupportedOS
	}

	return
}
