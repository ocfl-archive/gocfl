package factoryimpl

import (
	"context"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory/inventoryimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object/objectimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot/storagerootimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

func NewFactoryBase(version version.OCFLVersion, spec inventory.InventorySpec, extensionFactory *extensionimpl.Factory, logger ocfllogger.OCFLLogger) factory.Factory {
	return &FactoryBase{
		logger:           logger,
		version:          version,
		spec:             spec,
		extensionFactory: extensionFactory,
	}
}

type FactoryBase struct {
	logger           ocfllogger.OCFLLogger
	version          version.OCFLVersion
	spec             inventory.InventorySpec
	extensionFactory *extensionimpl.Factory
}

func (f *FactoryBase) Copy() factory.Factory {
	return &FactoryBase{
		logger:           f.logger,
		version:          f.version,
		spec:             f.spec,
		extensionFactory: f.extensionFactory,
	}
}

func (f *FactoryBase) SetVersion(version.OCFLVersion) error {
	f.logger.Panic().Msgf("cannot change version of fixed version factory")
	return nil
}

func (f *FactoryBase) GetVersion() version.OCFLVersion {
	return f.version
}

func (f *FactoryBase) NewStorageRoot(ctx context.Context) storageroot.StorageRoot {
	return storagerootimpl.NewStorageRootBase(ctx, f, f.version, f.extensionFactory, f.logger)
}

func (f *FactoryBase) NewLoader(ctx context.Context) object.Loader {
	return objectimpl.NewLoader(ctx, f, f.logger)
}

func (f *FactoryBase) NewInitializer(ctx context.Context) object.Initializer {
	return objectimpl.NewInitializer(ctx, f, f.logger)
}

func (f *FactoryBase) NewChecker(ctx context.Context) object.Checker {
	return objectimpl.NewObjectBaseChecker(ctx, f, f.logger)
}

func (f *FactoryBase) NewExtractor(ctx context.Context) object.Extractor {
	return objectimpl.NewExtractor(ctx, f, f.logger)
}

func (f *FactoryBase) NewObject(ctx context.Context) object.Object {
	return objectimpl.NewObjectBase(ctx, f, f.version, f.extensionFactory, f.logger)
}

func (f *FactoryBase) NewInventory(ctx context.Context) inventory.Inventory {
	return inventoryimpl.NewInventoryBase(ctx, f, f.version, f.spec, f.logger)
}

func (f *FactoryBase) NewFixity(ctx context.Context) inventory.Fixity {
	return inventoryimpl.NewFixityBase(
		[]checksum.DigestAlgorithm{checksum.DigestSHA256, checksum.DigestSHA512},
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

var _ factory.Factory = (*FactoryBase)(nil)
