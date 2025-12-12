package storageroot

import (
	"io/fs"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
)

type ExtensionStorageRootPath interface {
	extension.Extension
	WriteLayout(fsys fs.FS) error
	BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)
}
