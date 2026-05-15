package storageroot

import "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"

// ExtensionManager combines all extension interfaces relevant for OCFL storage roots.
type ExtensionManager interface {
	extension.ManagerCore[ExtensionManager]
	ExtensionStorageRootPath
}
