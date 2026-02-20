package objectimpl

import (
	"context"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	factorytypes "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

// NewObjectBase creates an empty ObjectBase structure
func NewObjectBase(ctx context.Context, factory factorytypes.Factory, defaultVersion version.OCFLVersion, extensionFactory *extensionimpl.ExtensionFactory, logger ocfllogger.OCFLLogger) *ObjectBase {
	objectBase := &ObjectBase{
		extensionFactory: extensionFactory,
		//extensionManager: extensionManager.(object.ExtensionManager),
		ctx: ctx,
		//		fsys: nil,
		i: factory.NewInventory(ctx).WithWriteable(),
		//versionFolders:     []string{},
		versionInventories: map[string]inventory.Inventory{},
		changed:            false,
		logger:             logger,
		version:            defaultVersion,
		digest:             "",
		echo:               false,
		updateFiles:        []string{},
		area:               "",
		factory:            factory,
	}
	return objectBase
}

type ObjectBase struct {
	//	storageRoot        storageroot.StorageRoot
	extensionFactory *extensionimpl.ExtensionFactory
	extensionManager object.ExtensionManager
	ctx              context.Context
	//fsys             fs.FS
	i inventory.Inventory
	//versionFolders     []string
	versionInventories map[string]inventory.Inventory
	changed            bool
	logger             ocfllogger.OCFLLogger
	version            version.OCFLVersion
	digest             checksum.DigestAlgorithm
	echo               bool
	updateFiles        []string
	area               string
	factory            factorytypes.Factory
}
