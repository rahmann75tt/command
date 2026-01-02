package command

import (
	"testing"
	"time"

	"lesiw.io/fs"
)

func TestParseGNUStat(t *testing.T) {
	tests := []struct {
		name string
		stat string
		path string
		want *fileInfo
		err  bool
	}{{
		name: "regular file",
		stat: "81a4 1234 1609459200 /path/to/file.txt",
		path: "/path/to/file.txt",
		want: &fileInfo{
			name:  "file.txt",
			size:  1234,
			mode:  0644,
			mtime: time.Unix(1609459200, 0),
			dir:   false,
		},
	}, {
		name: "directory",
		stat: "41ed 4096 1609459200 /path/to/dir",
		path: "/path/to/dir",
		want: &fileInfo{
			name:  "dir",
			size:  4096,
			mode:  0755 | fs.ModeDir,
			mtime: time.Unix(1609459200, 0),
			dir:   true,
		},
	}, {
		name: "invalid output - too few fields",
		stat: "81a4 1234",
		path: "/path/to/file",
		err:  true,
	}, {
		name: "invalid mode",
		stat: "INVALID 1234 1609459200 /path/to/file",
		path: "/path/to/file",
		err:  true,
	}, {
		name: "invalid size",
		stat: "81a4 INVALID 1609459200 /path/to/file",
		path: "/path/to/file",
		err:  true,
	}, {
		name: "invalid time",
		stat: "81a4 1234 INVALID /path/to/file",
		path: "/path/to/file",
		err:  true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGNUStat(tt.stat, tt.path)
			if tt.err {
				if err == nil {
					t.Errorf("parseGNUStat() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseGNUStat() error = %v", err)
			}
			compareFileInfo(t, got, tt.want)
		})
	}
}

func TestParseBSDStat(t *testing.T) {
	tests := []struct {
		name string
		stat string
		path string
		want *fileInfo
		err  bool
	}{{
		name: "regular file",
		stat: "100644 1234 1609459200 /path/to/file.txt",
		path: "/path/to/file.txt",
		want: &fileInfo{
			name:  "file.txt",
			size:  1234,
			mode:  0644,
			mtime: time.Unix(1609459200, 0),
			dir:   false,
		},
	}, {
		name: "directory",
		stat: "40755 4096 1609459200 /path/to/dir",
		path: "/path/to/dir",
		want: &fileInfo{
			name:  "dir",
			size:  4096,
			mode:  0755 | fs.ModeDir,
			mtime: time.Unix(1609459200, 0),
			dir:   true,
		},
	}, {
		name: "invalid output - too few fields",
		stat: "100644 1234",
		path: "/path/to/file",
		err:  true,
	}, {
		name: "invalid mode",
		stat: "INVALID 1234 1609459200 /path/to/file",
		path: "/path/to/file",
		err:  true,
	}, {
		name: "invalid size",
		stat: "100644 INVALID 1609459200 /path/to/file",
		path: "/path/to/file",
		err:  true,
	}, {
		name: "invalid time",
		stat: "100644 1234 INVALID /path/to/file",
		path: "/path/to/file",
		err:  true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBSDStat(tt.stat, tt.path)
			if tt.err {
				if err == nil {
					t.Errorf("parseBSDStat() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseBSDStat() error = %v", err)
			}
			compareFileInfo(t, got, tt.want)
		})
	}
}

func TestParseWindowsStat(t *testing.T) {
	tests := []struct {
		name string
		stat string
		path string
		want *fileInfo
		err  bool
	}{{
		name: "regular file",
		stat: "-a---- 1234 132539328000000000 file.txt",
		path: `C:\path\to\file.txt`,
		want: &fileInfo{
			name:  "file.txt",
			size:  1234,
			mode:  0644,
			mtime: time.Unix(1609459200, 0),
			dir:   false,
		},
	}, {
		name: "directory",
		stat: "d----- 0 132539328000000000 dirname",
		path: `C:\path\to\dirname`,
		want: &fileInfo{
			name:  "dirname",
			size:  0,
			mode:  0755 | fs.ModeDir,
			mtime: time.Unix(1609459200, 0),
			dir:   true,
		},
	}, {
		name: "invalid output - too few fields",
		stat: "-a---- 1234",
		path: `C:\path\to\file`,
		err:  true,
	}, {
		name: "invalid size",
		stat: "-a---- INVALID 132554112000000000 file.txt",
		path: `C:\path\to\file`,
		err:  true,
	}, {
		name: "invalid time",
		stat: "-a---- 1234 INVALID file.txt",
		path: `C:\path\to\file`,
		err:  true,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseWindowsStat(tt.stat, tt.path)
			if tt.err {
				if err == nil {
					t.Errorf("parseWindowsStat() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseWindowsStat() error = %v", err)
			}
			compareFileInfo(t, got, tt.want)
		})
	}
}

// compareFileInfo compares two FileInfo objects for testing.
func compareFileInfo(t *testing.T, got, want fs.FileInfo) {
	t.Helper()

	if got.Name() != want.Name() {
		t.Errorf("Name() = %q, want %q", got.Name(), want.Name())
	}
	if got.Size() != want.Size() {
		t.Errorf("Size() = %d, want %d", got.Size(), want.Size())
	}
	if got.Mode() != want.Mode() {
		t.Errorf("Mode() = %v, want %v", got.Mode(), want.Mode())
	}
	if !got.ModTime().Equal(want.ModTime()) {
		t.Errorf("ModTime() = %v, want %v", got.ModTime(), want.ModTime())
	}
	if got.IsDir() != want.IsDir() {
		t.Errorf("IsDir() = %v, want %v", got.IsDir(), want.IsDir())
	}
}
