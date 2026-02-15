package factoryimpl

import (
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewFactory20(extensionFactory *extensionimpl.ExtensionFactory, logger zLogger.ZLogger) factory.Factory {
	return &factory20{
		Factory: NewFactoryBase(version.Version1_0, inventory.InventorySpec1_0, extensionFactory, logger),
	}
}

type factory20 struct {
	factory.Factory
}

var _ factory.Factory = (*factory20)(nil)
