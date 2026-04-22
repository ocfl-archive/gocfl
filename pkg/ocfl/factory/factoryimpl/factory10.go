package factoryimpl

import (
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactory10(extensionFactory *extensionimpl.Factory, logger ocfllogger.OCFLLogger) factory.Factory {
	return &factory10{
		Factory: NewFactoryBase(version.Version1_0, inventory.InventorySpec1_0, extensionFactory, logger),
	}
}

type factory10 struct {
	factory.Factory
}

var _ factory.Factory = (*factory10)(nil)
