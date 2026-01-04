package storageroot

import (
	"io/fs"

	extensiontypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	inventorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot/storagerootimpl"
)

type ExtensionStorageRootPath interface {
	extensiontypes.Extension
	WriteLayout(fsys fs.FS) error
	BuildStorageRootPath(storageRoot inventorytypes.StorageRoot, id string) (string, error)
}
