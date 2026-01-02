package command

import (
	"context"
	"testing"

	"lesiw.io/fs"
)

var localizeTests = []struct {
	kind cfsKind
	name string
	work string
	path string
	want string
}{{
	kind: kindGNU,
	name: "GNU: simple file",
	path: "testdir/file.txt",
	want: "testdir/file.txt",
}, {
	kind: kindGNU,
	name: "GNU: directory with trailing slash",
	path: "testdir/",
	want: "testdir/",
}, {
	kind: kindGNU,
	name: "GNU: absolute path",
	path: "/tmp/testdir/file.txt",
	want: "/tmp/testdir/file.txt",
}, {
	kind: kindBSD,
	name: "BSD: simple file",
	path: "testdir/file.txt",
	want: "testdir/file.txt",
}, {
	kind: kindWindows,
	name: "Windows: convert forward to backslash",
	path: "testdir/file.txt",
	want: "testdir\\file.txt",
}, {
	kind: kindWindows,
	name: "Windows: already has backslash - no conversion",
	path: "testdir\\file.txt",
	want: "testdir\\file.txt",
}, {
	kind: kindWindows,
	name: "Windows: directory with trailing slash",
	path: "testdir/",
	want: "testdir\\",
}, {
	kind: kindWindows,
	name: "Windows: mixed slashes - already localized",
	path: "testdir/subdir\\file.txt",
	want: "testdir/subdir\\file.txt",
}, {
	kind: kindDOS,
	name: "DOS: convert forward to backslash",
	path: "testdir/file.txt",
	want: "testdir\\file.txt",
}, {
	kind: kindDOS,
	name: "DOS: already has backslash - no conversion",
	path: "testdir\\file.txt",
	want: "testdir\\file.txt",
}, {
	kind: kindDOS,
	name: "DOS: directory with trailing slash",
	path: "testdir/",
	want: "testdir\\",
}}

func TestLocalize(t *testing.T) {
	var ctx context.Context
	for _, tt := range localizeTests {
		t.Run(tt.name, func(t *testing.T) {
			ctx = t.Context()
			if tt.work != "" {
				ctx = fs.WithWorkDir(ctx, tt.work)
			}

			got, err := localize(ctx, tt.kind, tt.path)
			if err != nil {
				t.Fatalf("localize(%q) error = %v", tt.path, err)
			}

			if got != tt.want {
				t.Errorf("localize(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
