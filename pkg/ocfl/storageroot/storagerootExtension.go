package storageroot

import (
	extensiontypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
)

type ExtensionStorageRootPath interface {
	extensiontypes.Extension
	WriteLayout(fsys streamfs.FS) error
	BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)
}
