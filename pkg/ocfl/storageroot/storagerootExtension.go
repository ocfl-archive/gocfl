package storageroot

import (
	"github.com/je4/filesystem/v4/pkg/appendfs"
	extensiontypes "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
)

type ExtensionStorageRootPath interface {
	extensiontypes.Extension
	WriteLayout(fsys appendfs.FS) error
	BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)
}
