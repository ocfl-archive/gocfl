package object

import (
	"io"
	"io/fs"
)

type VersionFSMap interface {
	WithBaseFS(baseFS fs.FS) VersionFSMap
	GetVersionFS(version string) (fs.FS, error)
	Close() error
}

var _ io.Closer = (VersionFSMap)(nil)
