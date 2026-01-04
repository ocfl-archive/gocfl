package factoryimpl

import (
	"context"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	factorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory/inventoryimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object/objectimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func NewFactoryBase(version version.OCFLVersion, spec inventory.InventorySpec, extensionFactory *extension.ExtensionFactory, extensionManager object.ExtensionManager, logger zLogger.ZLogger) factorytypes.Factory {
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
	spec             inventory.InventorySpec
	extensionFactory *extension.ExtensionFactory
	extensionManager object.ExtensionManager
}

func (f *FactoryBase) NewObject(ctx context.Context) object.Object {
	return objectimpl.NewObjectBase(ctx, f, f.version, f.extensionFactory, f.extensionManager, f.logger)
}

func (f *FactoryBase) NewInventory(ctx context.Context) inventory.Inventory {
	return inventoryimpl.NewInventoryBase(ctx, f, f.version, f.spec, f.logger)
}

func (f *FactoryBase) NewFixity(context.Context) inventory.Fixity {
	return inventoryimpl.NewFixityBase(
		append(f.extensionManager.GetFixityDigests(), checksum.DigestSHA256, checksum.DigestSHA512),
		f.logger,
	)
}

func (f *FactoryBase) NewUser(context.Context) inventory.User {
	return inventoryimpl.NewUserBase()
}

func (f *FactoryBase) NewManifest(context.Context) inventory.Manifest {
	return inventoryimpl.NewManifestBase(f.logger)
}

func (f *FactoryBase) NewVersions(ctx context.Context) inventory.Versions {
	return inventoryimpl.NewVersionsBase(ctx, f, f.logger)
}
func (f *FactoryBase) NewVersion(ctx context.Context) inventory.Version {
	return inventoryimpl.NewVersionBase(ctx, f, f.logger)
}

func (f *FactoryBase) NewState(context.Context) inventory.State {
	return inventoryimpl.NewStateBase()
}

var _ factorytypes.Factory = (*FactoryBase)(nil)
