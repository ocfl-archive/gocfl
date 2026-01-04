package storageroot

import (
	extensiontypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/types"
)

type ExtensionManager interface {
	extensiontypes.ExtensionManagerCore
	ExtensionStorageRootPath
}
