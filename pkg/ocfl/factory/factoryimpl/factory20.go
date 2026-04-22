package factoryimpl

import (
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactory20(extensionFactory *extensionimpl.Factory, logger ocfllogger.OCFLLogger) factory.Factory {
	return &factory20{
		logger:  logger,
		Factory: NewFactoryBase(version.Version2_0, inventory.InventorySpec2_0, extensionFactory, logger),
	}
}

type factory20 struct {
	factory.Factory
	logger ocfllogger.OCFLLogger
}

var _ factory.Factory = (*factory20)(nil)
