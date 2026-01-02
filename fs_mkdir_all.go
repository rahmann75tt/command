package command

import (
	"context"
	"errors"

	"lesiw.io/fs"
)

var _ fs.MkdirAllFS = (*cmdFS)(nil)

func (cfs *cmdFS) MkdirAll(ctx context.Context, name string) (err error) {
	if err = cfs.init(ctx); err != nil {
		return
	}

	switch cfs.kind {
	case kindGNU, kindBSD:
		err = Do(ctx, cfs, "mkdir", "-p", name)
	case kindWindows, kindDOS:
		err = Do(ctx, cfs, "cmd", "/c", "mkdir", name)
		if e := new(Error); errors.As(err, &e) && e.Code == 1 {
			err = nil // The directory already existing is not an error.
		}
	default:
		err = errUnsupportedOS
	}

	return
}
