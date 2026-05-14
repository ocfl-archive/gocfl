package factoryimpl

import (
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactoryStorageRoot10(extensionFactory extension.Factory[storageroot.ExtensionManager], logger ocfllogger.OCFLLogger) factory.FactoryStorageRoot {
	return &factoryStorageRoot10{
		FactoryStorageRoot: NewFactoryBaseStorageRoot(version.Version1_0, extensionFactory, logger),
	}
}

type factoryStorageRoot10 struct {
	factory.FactoryStorageRoot
}

var _ factory.FactoryStorageRoot = (*factoryStorageRoot10)(nil)
