package storageroot

import (
	"github.com/je4/filesystem/v4/pkg/appendfs"
	extensiontypes "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
)

// ExtensionStorageRootPath is an interface for extensions that can build storage root paths (layouts).
type ExtensionStorageRootPath interface {
	extensiontypes.Extension
	// WriteLayout writes the layout configuration to the storage root.
	WriteLayout(fsys appendfs.FS) error
	// BuildStorageRootPath builds a storage root path from an object ID.
	BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)
}
