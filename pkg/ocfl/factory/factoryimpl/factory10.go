package factoryimpl

import (
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	factorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	inventorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	objecttypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewFactory10(extensionFactory *extension.ExtensionFactory, extensionManager objecttypes.ExtensionManager, logger zLogger.ZLogger) factorytypes.Factory {
	return &factory10{
		Factory: NewFactoryBase(version.Version1_0, inventorytypes.InventorySpec1_0, extensionFactory, extensionManager, logger),
	}
}

type factory10 struct {
	factorytypes.Factory
}

var _ factorytypes.Factory = (*factory10)(nil)
