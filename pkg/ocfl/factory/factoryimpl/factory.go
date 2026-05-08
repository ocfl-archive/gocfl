package factoryimpl

import (
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension/extensionimpl"
	factorytypes "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactoryObject(ver version.OCFLVersion, extensionFactory *extensionimpl.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) factorytypes.FactoryObject {
	switch ver {
	case version.Version1_0:
		return NewFactoryObject10(extensionFactory, logger)
	case version.Version1_1:
		return NewFactoryObject11(extensionFactory, logger)
	case version.Version2_0:
		return NewFactoryObject20(extensionFactory, logger)
		// todo: should we do a default??? or add errors
	default:
		return NewFactoryObject11(extensionFactory, logger)
	}
}

func NewFactoryStorageRoot(ver version.OCFLVersion, extensionFactory *extensionimpl.Factory[storageroot.ExtensionManager], logger ocfllogger.OCFLLogger) factorytypes.FactoryStorageRoot {
	switch ver {
	case version.Version1_0:
		return NewFactoryStorageRoot10(extensionFactory, logger)
	case version.Version1_1:
		return NewFactoryStorageRoot11(extensionFactory, logger)
	case version.Version2_0:
		return NewFactoryStorageRoot20(extensionFactory, logger)
		// todo: should we do a default??? or add errors
	default:
		return NewFactoryStorageRoot11(extensionFactory, logger)
	}
}
