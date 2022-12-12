// Package ocfl for manipulating and checking Oxford Common Filesystem Layout
// This Oxford Common File Layout (OCFL) specification describes an
// application-independent approach to the storage of digital information in a
// structured, transparent, and predictable manner. It is designed to promote
// long-term object management best practices within digital repositories.
// https://ocfl.io
package ocfl

import (
	"io"
	"io/fs"
)

// Filesystem abstraction for OCFL access
type OCFLFS interface {
	fs.ReadDirFS
	Create(name string) (io.WriteCloser, error)
	Delete(name string) error
	SubFS(subfolder string) (OCFLFS, error)
	Close() error
	Discard() error
	String() string
	IsNotExist(err error) bool
	WalkDir(root string, fn fs.WalkDirFunc) error
	Stat(name string) (fs.FileInfo, error)
	HasContent() bool
}
