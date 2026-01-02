package inventory

import (
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/types"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewFactory10(extensionFactory *extension.ExtensionFactory, extensionManager extension.ExtensionManager, logger zLogger.ZLogger) types.Factory {
	return &factory10{
		Factory: NewFactoryBase(version.Version1_0, types.InventorySpec1_0, extensionFactory, extensionManager, logger),
	}
}

type factory10 struct {
	types.Factory
}

var _ types.Factory = (*factory10)(nil)
