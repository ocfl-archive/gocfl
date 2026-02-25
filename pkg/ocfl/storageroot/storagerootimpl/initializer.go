package storagerootimpl

import (
	"context"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/docs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
)

func NewInitializer(ctx context.Context, factory factory.Factory, logger ocfllogger.OCFLLogger) object.Initializer {
	return &initializer{
		ctx:     ctx,
		factory: factory,
		logger:  logger.With("task", "storage root initializer"),
	}
}

type initializer struct {
	object.Object
	storageRootFS streamfs.FS
	logger        ocfllogger.OCFLLogger
	ctx           context.Context
	factory       factory.Factory
}

func (initializer *initializer) WithObject(o object.Object) object.Initializer {
	initializer.Object = o
	return initializer
}

func (initializer *initializer) WithFS(objectFS streamfs.FS) object.Initializer {
	initializer.storageRootFS = objectFS
	return initializer
}

func (initializer *initializer) Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error {
	initializer.logger.Debug().Msgf("%s", id)

	objectConformanceDeclaration := "ocfl_" + string(initializer.factory.GetVersion())
	objectConformanceDeclarationFile := "0=" + objectConformanceDeclaration
	if _, err := writefs.WriteFile(initializer.storageRootFS, objectConformanceDeclarationFile, []byte(objectConformanceDeclaration+"\n")); err != nil {
		return errors.Wrapf(err, "cannot create '%v/%s'", initializer.storageRootFS, objectConformanceDeclarationFile)
	}

	/*
		if err := writefs.MkDir(initializer.storageRootFS, "extensions"); err != nil {
			return errors.Wrapf(err, "cannot create '%v/%s'", initializer.storageRootFS, "extensions")
		}
	*/
	subFS, err := streamfs.Sub(initializer.storageRootFS, "extensions")
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs of %v for folder '%s'", initializer.storageRootFS, "extensions")
	}
	if err := initializer.GetExtensionManager().WriteConfig(subFS); err != nil {
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
		if _, err := writefs.WriteFile(initializer.storageRootFS, extDoc.Name(), extDocFileContent); err != nil {
			return errors.Wrapf(err, "cannot write extension doc %v/%s", initializer.storageRootFS, extDoc.Name())
		}
	}
	if err := initializer.GetExtensionManager().StoreRootLayout(initializer.storageRootFS); err != nil {
		return errors.Wrap(err, "cannot store ocfl layout")
	}
	return nil

}

var _ object.Initializer = (*initializer)(nil)
