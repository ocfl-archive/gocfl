package factoryimpl

import (
	"context"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory/inventoryimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object/objectimpl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewFactoryBaseObject(version version.OCFLVersion, spec inventory.InventorySpec, extensionFactory extension.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) factory.FactoryObject {
	return &FactoryBaseObject{
		logger:           logger,
		version:          version,
		spec:             spec,
		extensionFactory: extensionFactory,
	}
}

type FactoryBaseObject struct {
	logger           ocfllogger.OCFLLogger
	version          version.OCFLVersion
	spec             inventory.InventorySpec
	extensionFactory extension.Factory[object.ExtensionManager]
}

func (f *FactoryBaseObject) WithNewVersion(ocflVersion version.OCFLVersion) factory.FactoryObject {
	return NewFactoryBaseObject(ocflVersion, f.spec, f.extensionFactory, f.logger)
}

func (f *FactoryBaseObject) Copy() factory.FactoryObject {
	return &FactoryBaseObject{
		logger:           f.logger,
		version:          f.version,
		spec:             f.spec,
		extensionFactory: f.extensionFactory,
	}
}

func (f *FactoryBaseObject) GetVersion() version.OCFLVersion {
	return f.version
}

func (f *FactoryBaseObject) NewLoader(ctx context.Context) object.Loader {
	return objectimpl.NewLoader(ctx, f, f.logger)
}

func (f *FactoryBaseObject) NewInitializer(ctx context.Context) object.Initializer {
	return objectimpl.NewInitializer(ctx, f, f.logger)
}

func (f *FactoryBaseObject) NewChecker(ctx context.Context) object.Checker {
	return objectimpl.NewObjectBaseChecker(ctx, f, f.logger)
}

func (f *FactoryBaseObject) NewExtractor(ctx context.Context) object.Extractor {
	return objectimpl.NewExtractor(ctx, f, f.logger)
}

func (f *FactoryBaseObject) NewObject(ctx context.Context) object.Object {
	return objectimpl.NewObjectBase(ctx, f, f.version, f.extensionFactory, f.logger)
}

func (f *FactoryBaseObject) NewInventory(ctx context.Context) inventory.Inventory {
	return inventoryimpl.NewInventoryBase(ctx, f, f.version, f.spec, f.logger)
}

func (f *FactoryBaseObject) NewFixity(ctx context.Context) inventory.Fixity {
	return inventoryimpl.NewFixityBase(
		[]checksum.DigestAlgorithm{checksum.DigestSHA256, checksum.DigestSHA512},
		f.logger,
	)
}

func (f *FactoryBaseObject) NewUser(context.Context) inventory.User {
	return inventoryimpl.NewUserBase(f.logger)
}

func (f *FactoryBaseObject) NewManifest(context.Context) inventory.Manifest {
	return inventoryimpl.NewManifestBase(f.logger)
}

func (f *FactoryBaseObject) NewVersions(ctx context.Context) inventory.Versions {
	return inventoryimpl.NewVersionsBase(ctx, f, f.logger)
}
func (f *FactoryBaseObject) NewVersion(ctx context.Context) inventory.Version {
	return inventoryimpl.NewVersionBase(ctx, f, f.logger)
}

func (f *FactoryBaseObject) NewState(context.Context) inventory.State {
	return inventoryimpl.NewStateBase(f.logger)
}

var _ factory.FactoryObject = (*FactoryBaseObject)(nil)
