package initocfl

import (
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory/factoryimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactoryObject(
	ver version.OCFLVersion,
	extensionFactory extension.Factory[object.ExtensionManager],
	logger ocfllogger.OCFLLogger,
) factory.FactoryObject {
	switch ver {
	case version.Version1_0:
		return factoryimpl.NewFactoryObject10(extensionFactory, logger)
	case version.Version1_1:
		return factoryimpl.NewFactoryObject11(extensionFactory, logger)
	case version.Version2_0:
		return factoryimpl.NewFactoryObject20(extensionFactory, logger)
		// todo: should we do a default??? or add errors
	default:
		return factoryimpl.NewFactoryObject11(extensionFactory, logger)
	}
}

func NewFactoryStorageRoot(
	ver version.OCFLVersion,
	extensionFactory extension.Factory[storageroot.ExtensionManager],
	logger ocfllogger.OCFLLogger,
) factory.FactoryStorageRoot {
	switch ver {
	case version.Version1_0:
		return factoryimpl.NewFactoryStorageRoot10(extensionFactory, logger)
	case version.Version1_1:
		return factoryimpl.NewFactoryStorageRoot11(extensionFactory, logger)
	case version.Version2_0:
		return factoryimpl.NewFactoryStorageRoot20(extensionFactory, logger)
		// todo: should we do a default??? or add errors
	default:
		return factoryimpl.NewFactoryStorageRoot11(extensionFactory, logger)
	}
}
