package storageroot

import (
	"io/fs"

	extensiontypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
)

type ExtensionStorageRootPath interface {
	extensiontypes.Extension
	WriteLayout(fsys fs.FS) error
	BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)
}
