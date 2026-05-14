package factoryimpl

import (
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactoryStorageRoot20(extensionFactory extension.Factory[storageroot.ExtensionManager], logger ocfllogger.OCFLLogger) factory.FactoryStorageRoot {
	return &factoryStorageRoot20{
		FactoryStorageRoot: NewFactoryBaseStorageRoot(version.Version2_0, extensionFactory, logger),
	}
}

type factoryStorageRoot20 struct {
	factory.FactoryStorageRoot
}

var _ factory.FactoryStorageRoot = (*factoryStorageRoot20)(nil)
