package factoryimpl

import (
	"context"

	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot/storagerootimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactoryBaseStorageRoot(version version.OCFLVersion, extensionFactory extension.Factory[storageroot.ExtensionManager], logger ocfllogger.OCFLLogger) factory.FactoryStorageRoot {
	return &FactoryBaseStorageRoot{
		logger:           logger,
		version:          version,
		extensionFactory: extensionFactory,
		config:           make(map[storageroot.ConfigName]any),
	}
}

type FactoryBaseStorageRoot struct {
	logger           ocfllogger.OCFLLogger
	version          version.OCFLVersion
	extensionFactory extension.Factory[storageroot.ExtensionManager]
	config           map[storageroot.ConfigName]any
}

// WithConfig sets the configuration for the factory.
func (f *FactoryBaseStorageRoot) WithConfig(config map[storageroot.ConfigName]any) factory.FactoryStorageRoot {
	f.config = config
	return f
}

func (f *FactoryBaseStorageRoot) WithNewVersion(ocflVersion version.OCFLVersion) factory.FactoryStorageRoot {
	return NewFactoryBaseStorageRoot(ocflVersion, f.extensionFactory, f.logger)
}

func (f *FactoryBaseStorageRoot) Copy() factory.FactoryStorageRoot {
	return &FactoryBaseStorageRoot{
		logger:           f.logger,
		version:          f.version,
		extensionFactory: f.extensionFactory,
		config:           f.config,
	}
}

func (f *FactoryBaseStorageRoot) GetVersion() version.OCFLVersion {
	return f.version
}

func (f *FactoryBaseStorageRoot) NewStorageRoot(ctx context.Context) storageroot.StorageRoot {
	return storagerootimpl.NewStorageRootBase(ctx, f, f.version, f.extensionFactory, f.config[storageroot.StorageRootName], f.logger)
}

func (f *FactoryBaseStorageRoot) NewStorageRootInitializer(ctx context.Context) storageroot.Initializer {
	return storagerootimpl.NewInitializer(ctx, f, f.extensionFactory, f.config[storageroot.InitializerName], f.logger)
}

func (f *FactoryBaseStorageRoot) NewStorageRootLoader(ctx context.Context) storageroot.Loader {
	return storagerootimpl.NewLoader(ctx, f.config[storageroot.LoaderName], f.logger)
}

var _ factory.FactoryStorageRoot = (*FactoryBaseStorageRoot)(nil)
