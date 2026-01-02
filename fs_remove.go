package command

import (
	"context"

	"lesiw.io/fs"
)

var _ fs.RemoveFS = (*cmdFS)(nil)

func (cfs *cmdFS) Remove(ctx context.Context, name string) (err error) {
	if err = cfs.init(ctx); err != nil {
		return
	}

	switch cfs.kind {
	case kindGNU, kindBSD:
		if err = Do(ctx, cfs, "rm", name); err != nil {
			err = Do(ctx, cfs, "rmdir", name)
		}
	case kindWindows:
		err = psDo(ctx, cfs, "Remove-Item -Path '%s'", name)
	case kindDOS:
		if err = Do(ctx, cfs, "cmd", "/c", "del", "/Q", name); err != nil {
			err = Do(ctx, cfs, "cmd", "/c", "rmdir", name)
		}
	default:
		return errUnsupportedOS
	}

	return
}
