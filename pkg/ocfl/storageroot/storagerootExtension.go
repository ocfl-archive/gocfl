package storageroot

import (
	"io/fs"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/types"
)

type ExtensionStorageRootPath interface {
	types.Extension
	WriteLayout(fsys fs.FS) error
	BuildStorageRootPath(storageRoot types.StorageRoot, id string) (string, error)
}
