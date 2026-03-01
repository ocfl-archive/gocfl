package storagerootimpl

import (
	"context"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/ocfl-archive/gocfl/v2/docs"
	"github.com/ocfl-archive/gocfl/v2/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

func NewInitializer(ctx context.Context, factory factory.Factory, extensionFactory extension.Factory, logger ocfllogger.OCFLLogger) storageroot.Initializer {
	return &initializer{
		ctx:             ctx,
		factory:         factory,
		extensionFactor: extensionFactory,
		logger:          logger.With("task", "storage root initializer"),
	}
}

type initializer struct {
	storageroot.StorageRoot
	storageRootFS   appendfs.FS
	logger          ocfllogger.OCFLLogger
	ctx             context.Context
	factory         factory.Factory
	extensionFactor extension.Factory
}

func (init *initializer) Close() error {
	return nil
}

func (init *initializer) WithStorageRoot(sr storageroot.StorageRoot) storageroot.Initializer {
	init.StorageRoot = sr
	return init
}

func (init *initializer) WithFS(storageRootFS appendfs.FS) storageroot.Initializer {
	init.storageRootFS = storageRootFS
	return init
}

func (init *initializer) Init() error {

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

	extDocs, err := docs.ExtensionDocs.ReadDir(".")
	if err != nil {
		return errors.Wrap(err, "cannot read extension docs")
	}
	for _, extDoc := range extDocs {
		if extDoc.IsDir() {
			continue
		}
		extDocFileContent, err := docs.ExtensionDocs.ReadFile(extDoc.Name())
		if err != nil {
			return errors.Wrapf(err, "cannot open extension doc %s", extDoc.Name())
		}
		if _, err := writefs.WriteFile(init.storageRootFS, extDoc.Name(), extDocFileContent); err != nil {
			return errors.Wrapf(err, "cannot write extension doc %v/%s", init.storageRootFS, extDoc.Name())
		}
	}
	if err := init.GetExtensionManager().StoreRootLayout(init.storageRootFS); err != nil {
		return errors.Wrap(err, "cannot store ocfl layout")
	}
	return nil

}

var _ storageroot.Initializer = (*initializer)(nil)
