package storagerootimpl

import (
	"io/fs"

	extensiontypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"
)

type ExtensionStorageRootPath interface {
	extensiontypes.Extension
	WriteLayout(fsys fs.FS) error
	BuildStorageRootPath(storageRoot storageroot.StorageRoot, id string) (string, error)
}
