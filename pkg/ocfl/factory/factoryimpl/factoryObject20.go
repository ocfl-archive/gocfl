package factoryimpl

import (
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactoryObject20(extensionFactory extension.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) factory.FactoryObject {
	return &factoryObject20{
		FactoryObject: NewFactoryBaseObject(version.Version2_0, inventory.InventorySpec2_0, extensionFactory, logger),
	}
}

type factoryObject20 struct {
	factory.FactoryObject
}

var _ factory.FactoryObject = (*factoryObject20)(nil)
