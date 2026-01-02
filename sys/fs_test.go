package sys

import (
	"fmt"
	"testing"

	"lesiw.io/command"
	"lesiw.io/fs"
	"lesiw.io/fs/fstest"
)

func TestFSCompliance(t *testing.T) {
	// Run compliance tests twice: once with osfs optimization,
	// once without (to ensure command-based FS still works).
	for _, osfs := range []bool{true, false} {
		t.Run(fmt.Sprintf("osfs %v", osfs), func(t *testing.T) {
			oldUseOSFS := useOSFS
			useOSFS = osfs
			t.Cleanup(func() { useOSFS = oldUseOSFS })
			sh := command.Shell(Machine())
			ctx := fs.WithWorkDir(t.Context(), t.TempDir())
			ctx = command.WithEnv(ctx, map[string]string{"TZ": "UTC"})
			fstest.TestFS(ctx, t, command.FS(sh))
		})
	}
}
