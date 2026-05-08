package factoryimpl

import (
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactoryObject10(extensionFactory *extensionimpl.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) factory.FactoryObject {
	return &factoryObject10{
		FactoryObject: NewFactoryBaseObject(version.Version1_0, inventory.InventorySpec1_0, extensionFactory, logger),
	}
}

type factoryObject10 struct {
	factory.FactoryObject
}

var _ factory.FactoryObject = (*factoryObject10)(nil)
