package storageroot

import "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"

type ExtensionManager interface {
	extension.ExtensionManager
	ExtensionStorageRootPath
}
