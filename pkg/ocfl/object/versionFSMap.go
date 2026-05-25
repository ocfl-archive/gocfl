package object

import (
	"io"
	"io/fs"

	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
)

type VersionFSMap interface {
	WithBaseFS(baseFS fs.FS) VersionFSMap
	GetVersionFS(version *inventory.VersionNumber) (fs.FS, error)
	Close() error
}

var _ io.Closer = (VersionFSMap)(nil)
