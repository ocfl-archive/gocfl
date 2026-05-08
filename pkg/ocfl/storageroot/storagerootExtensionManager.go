package storageroot

import "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"

type ExtensionManager interface {
	extension.ManagerCore[ExtensionManager]
	ExtensionStorageRootPath
}
