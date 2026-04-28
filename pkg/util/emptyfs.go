package util

import (
	"io/fs"
	"os"
)

// EmptyFS is an empty file system that implements fs.FS.
type EmptyFS struct{}

// Open always returns an error because the file system is empty.
func (EmptyFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	return nil, &fs.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}

// Check if EmptyFS implements fs.FS
var _ fs.FS = EmptyFS{}
