package appendfs

import (
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v4/pkg/writefs"
)

type FS interface {
	fs.FS
	writefs.MkDirFS
	writefs.CreateFS
}

var ErrNotAppendFS = errors.New("provided filesystem does not implement appendfs.FS interface")

// New returns a new appendfs.FS. It wraps an existing fs.FS to ensure it implements
// the appendfs.FS interface. Primarily used for testing to enforce restricted functionality.
// If fSys does not implement appendfs.FS, appendfs.ErrNotAppendFS is returned.
func New(fSys fs.FS) (FS, error) {
	aFS, ok := fSys.(FS)
	if !ok {
		return nil, ErrNotAppendFS
	}
	return &appendFS{fs: aFS}, nil
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
	return a.fs.Create(path)
}
