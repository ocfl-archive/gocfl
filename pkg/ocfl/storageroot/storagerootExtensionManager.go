package storageroot

import (
	extensiontypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
)

type ExtensionManager interface {
	extensiontypes.ExtensionManagerCore
	ExtensionStorageRootPath
}
