package factory

import (
	"context"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/types"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewFactoryBase(version version.OCFLVersion, spec types.InventorySpec, extensionFactory *extension.ExtensionFactory, extensionManager types.ExtensionManager, logger zLogger.ZLogger) types.Factory {
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
	extensionManager types.ExtensionManager
}

func (f *FactoryBase) NewObject(ctx context.Context) types.Object {
	return object.NewObjectBase(ctx, f, f.version, f.extensionFactory, f.extensionManager, f.logger)
}

func (f *FactoryBase) NewInventory(ctx context.Context) types.Inventory {
	return inventory.NewInventoryBase(ctx, f, f.version, f.spec, f.logger)
}

func (f *FactoryBase) NewFixity(context.Context) types.Fixity {
	return inventory.NewFixityBase(
		append(f.extensionManager.GetFixityDigests(), checksum.DigestSHA256, checksum.DigestSHA512),
		f.logger,
	)
}

func (f *FactoryBase) NewUser(context.Context) types.User {
	return inventory.NewUserBase()
}

func (f *FactoryBase) NewManifest(context.Context) types.Manifest {
	return inventory.NewManifestBase(f.logger)
}

func (f *FactoryBase) NewVersions(context.Context) types.Versions {
	return inventory.NewVersionsBase(f, f.logger)
}
func (f *FactoryBase) NewVersion(context.Context) types.Version {
	return inventory.NewVersionBase(f, f.logger)
}

func (f *FactoryBase) NewState(context.Context) types.State {
	return inventory.NewStateBase()
}

var _ types.Factory = (*FactoryBase)(nil)
