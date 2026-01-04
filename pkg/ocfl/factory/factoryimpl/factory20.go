package factoryimpl

import (
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	factorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	inventorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	objecttypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewFactory20(extensionFactory *extension.ExtensionFactory, extensionManager objecttypes.ExtensionManager, logger zLogger.ZLogger) factorytypes.Factory {
	return &factory20{
		logger:  logger,
		Factory: NewFactoryBase(version.Version1_1, inventorytypes.InventorySpec1_1, extensionFactory, extensionManager, logger),
	}
}

type factory20 struct {
	factorytypes.Factory
	logger zLogger.ZLogger
}

var _ factorytypes.Factory = (*factory20)(nil)
