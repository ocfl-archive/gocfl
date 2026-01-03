package storageroot

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/types"
)

type ExtensionManager interface {
	types.ExtensionManagerCore
	ExtensionStorageRootPath
}
