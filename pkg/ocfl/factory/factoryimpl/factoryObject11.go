package factoryimpl

import (
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactoryObject11(extensionFactory extension.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) factory.FactoryObject {
	return &factoryObject11{
		FactoryObject: NewFactoryBaseObject(version.Version1_1, inventory.InventorySpec1_1, extensionFactory, logger),
	}
}

type factoryObject11 struct {
	factory.FactoryObject
}

var _ factory.FactoryObject = (*factoryObject11)(nil)
