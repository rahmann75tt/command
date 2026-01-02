package command

import (
	"context"
	"errors"

	"lesiw.io/fs"
)

var _ fs.RemoveAllFS = (*cmdFS)(nil)

func (cfs *cmdFS) RemoveAll(ctx context.Context, name string) (err error) {
	if err = cfs.init(ctx); err != nil {
		return
	}

	switch cfs.kind {
	case kindGNU, kindBSD:
		err = Do(ctx, cfs, "rm", "-rf", name)
	case kindWindows:
		err = psDo(ctx, cfs,
			"Remove-Item -Path '%s' -Recurse -Force "+
				"-ErrorAction SilentlyContinue",
			name,
		)
		if e := new(Error); errors.As(err, &e) && e.Code == 1 {
			err = nil // Ignore "path not found" errors.
		}
	case kindDOS:
		err = Do(ctx, cfs, "cmd", "/c", "rmdir", "/S", "/Q", name)
		if e := new(Error); errors.As(err, &e) && e.Code == 2 {
			err = nil // Ignore "path not found" errors.
		}
	default:
		err = errUnsupportedOS
	}

	return
}
