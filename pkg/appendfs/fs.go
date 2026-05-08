package appendfs

import (
	"io/fs"

	"github.com/je4/filesystem/v3/pkg/writefs"
)

type FS interface {
	fs.FS
	writefs.MkDirFS
	writefs.CreateFS
}

func New(fSys FS) FS {
	return &appendFS{fs: fSys}
}

// appendFS is a wrapper for `FS` interface to enforce restricted functionality, primarily for testing purposes.
type appendFS struct {
	fs FS
}

func (a *appendFS) Open(name string) (fs.File, error) {
	return a.fs.Open(name)
}

func (a *appendFS) MkDir(path string) error {
	return a.fs.MkDir(path)
}

func (a *appendFS) Create(path string) (writefs.FileWrite, error) {
	return a.Create(path)
}
