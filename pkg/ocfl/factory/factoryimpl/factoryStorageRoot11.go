package factoryimpl

import (
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactoryStorageRoot11(extensionFactory extension.Factory[storageroot.ExtensionManager], logger ocfllogger.OCFLLogger) factory.FactoryStorageRoot {
	return &factoryStorageRoot11{
		FactoryStorageRoot: NewFactoryBaseStorageRoot(version.Version1_1, extensionFactory, logger),
	}
}

type factoryStorageRoot11 struct {
	factory.FactoryStorageRoot
}

var _ factory.FactoryStorageRoot = (*factoryStorageRoot11)(nil)
