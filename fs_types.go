package command

import (
	"time"

	"lesiw.io/fs"
)

var _ fs.FileInfo = (*fileInfo)(nil)

type fileInfo struct {
	name  string
	size  int64
	mode  fs.Mode
	mtime time.Time
	dir   bool
}

func (fi fileInfo) Name() string       { return fi.name }
func (fi fileInfo) Size() int64        { return fi.size }
func (fi fileInfo) Mode() fs.Mode      { return fi.mode }
func (fi fileInfo) ModTime() time.Time { return fi.mtime }
func (fi fileInfo) IsDir() bool        { return fi.dir }
func (fi fileInfo) Sys() any           { return nil }

var _ fs.DirEntry = (*dirEntry)(nil)

type dirEntry struct {
	name string
	dir  bool
	mode fs.Mode
	info fs.FileInfo
	path string
}

func (de dirEntry) Name() string               { return de.name }
func (de dirEntry) IsDir() bool                { return de.dir }
func (de dirEntry) Type() fs.Mode              { return de.mode }
func (de dirEntry) Info() (fs.FileInfo, error) { return de.info, nil }
func (de dirEntry) Path() string               { return de.path }
