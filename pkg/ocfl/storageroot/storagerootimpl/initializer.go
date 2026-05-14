package storagerootimpl

import (
	"context"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v4/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewInitializer(ctx context.Context, factory factory.FactoryStorageRoot, extensionFactory extension.Factory[storageroot.ExtensionManager], logger ocfllogger.OCFLLogger) storageroot.Initializer {
	return &initializer{
		ctx:              ctx,
		factory:          factory,
		extensionFactory: extensionFactory,
		logger:           logger.With("task", "storage root initializer"),
	}
}

type initializer struct {
	storageroot.StorageRoot
	storageRootFS    appendfs.FS
	logger           ocfllogger.OCFLLogger
	ctx              context.Context
	factory          factory.FactoryStorageRoot
	extensionFactory extension.Factory[storageroot.ExtensionManager]
}

func (init *initializer) Close() error {
	return nil
}

func (init *initializer) SetStorageRoot(sr storageroot.StorageRoot) storageroot.Initializer {
	init.StorageRoot = sr
	return init
}

func (init *initializer) Init() error {
	init.storageRootFS = init.GetWriteFS()
	if init.storageRootFS == nil {
		return errors.New("storageRootFS is nil. initialize StorageRoot with WithWriteFS() first")
	}
	if empty, err := writefs.IsEmpty(init.storageRootFS, ""); err != nil {
		if errors.Is(err, writefs.ErrNotImplemented) {
			init.logger.Warn().Msgf("cannot check whether %v is empty", init.storageRootFS)
		} else {
			return errors.Wrapf(err, "cannot check whether %v is empty", init.storageRootFS)
		}
	} else if !empty {
		return errors.Errorf("cannot create storage root. '%v' is not empty", init.storageRootFS)
	}
	extensionManager := init.GetExtensionManager()
	if extensionManager == nil {
		return errors.New("extension manager is nil. initialize with WithExtensionManager() first")
	}
	objectConformanceDeclaration := "ocfl_" + string(init.factory.GetVersion())
	objectConformanceDeclarationFile := "0=" + objectConformanceDeclaration
	if _, err := writefs.WriteFile(init.storageRootFS, objectConformanceDeclarationFile, []byte(objectConformanceDeclaration+"\n")); err != nil {
		return errors.Wrapf(err, "cannot create '%v/%s'", init.storageRootFS, objectConformanceDeclarationFile)
	}

	/*
		if err := writefs.MkDir(init.storageRootFS, "extensions"); err != nil {
			return errors.Wrapf(err, "cannot create '%v/%s'", init.storageRootFS, "extensions")
		}
	*/
	subFS, err := appendfs.Sub(init.storageRootFS, "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", init.storageRootFS, "extensions")
	}
	if err := init.GetExtensionManager().WriteConfig(subFS); err != nil {
		return errors.Wrapf(err, "cannot write extension config to %v", subFS)
	}

	for name, doc := range init.extensionFactory.GetExtensionDocs() {
		if doc == nil {
			continue
		}
		if _, err := writefs.WriteFile(init.storageRootFS, name+".md", []byte(*doc)); err != nil {
			return errors.Wrapf(err, "cannot write extension doc %v/%s", init.storageRootFS, name)
		}
	}
	if spec, ok := version.Spec[init.GetVersion()]; ok {
		specName := "ocfl_spec_" + init.GetVersion().String() + ".md"
		if _, err := writefs.WriteFile(init.storageRootFS, specName, []byte(spec)); err != nil {
			return errors.Wrapf(err, "cannot write extension doc %v/%s", init.storageRootFS, specName)
		}
	}

	if err := init.GetExtensionManager().StoreRootLayout(init.storageRootFS); err != nil {
		return errors.Wrap(err, "cannot store ocfl layout")
	}
	return nil

}

var _ storageroot.Initializer = (*initializer)(nil)
