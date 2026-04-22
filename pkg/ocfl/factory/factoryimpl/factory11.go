package factoryimpl

import (
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactory11(extensionFactory *extensionimpl.Factory, logger ocfllogger.OCFLLogger) factory.Factory {
	return &factory11{
		logger:  logger,
		Factory: NewFactoryBase(version.Version1_1, inventory.InventorySpec1_1, extensionFactory, logger),
	}
}

type factory11 struct {
	factory.Factory
	logger ocfllogger.OCFLLogger
}

var _ factory.Factory = (*factory11)(nil)
