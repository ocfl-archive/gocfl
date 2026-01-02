package inventory

import (
	"context"

	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/types"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewFactoryBase(version version.OCFLVersion, spec types.InventorySpec, extensionFactory *extension.ExtensionFactory, extensionManager extension.ExtensionManager, logger zLogger.ZLogger) types.Factory {
	return &FactoryBase{
		logger:           logger,
		version:          version,
		spec:             spec,
		extensionFactory: extensionFactory,
		extensionManager: extensionManager,
	}
}

type FactoryBase struct {
	logger           zLogger.ZLogger
	version          version.OCFLVersion
	spec             types.InventorySpec
	extensionFactory *extension.ExtensionFactory
	extensionManager extension.ExtensionManager
}

func (f *FactoryBase) NewObject(ctx context.Context) types.Object {
	return object.NewObjectBase(ctx, f, f.version, f.extensionFactory, f.extensionManager, f.logger)
}

func (f *FactoryBase) NewInventory(ctx context.Context) types.Inventory {
	return NewInventoryBase(ctx, f, f.version, f.spec, f.logger)
}

func (f *FactoryBase) NewFixity(context.Context) types.Fixity {
	return NewFixityBase(f.logger)
}

func (f *FactoryBase) NewUser(context.Context) types.User {
	return NewUserBase()
}

func (f *FactoryBase) NewManifest(context.Context) types.Manifest {
	return NewManifestBase(f.logger)
}

func (f *FactoryBase) NewVersions(context.Context) types.Versions {
	return NewVersionsBase(f, f.logger)
}
func (f *FactoryBase) NewVersion(context.Context) types.Version {
	return NewVersionBase(f, f.logger)
}

func (f *FactoryBase) NewState(context.Context) types.State {
	return NewStateBase()
}

var _ types.Factory = (*FactoryBase)(nil)
