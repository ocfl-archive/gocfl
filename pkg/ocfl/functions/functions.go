package functions

import (
	"context"
	"fmt"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory/factoryimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/ocflerrors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

func CreateObject(
	ctx context.Context,
	id string,
	ver version.OCFLVersion,
	digest checksum.DigestAlgorithm,
	fixity []checksum.DigestAlgorithm,
	extensionFactory *extensionimpl.Factory,
	manager object.ExtensionManager,
	fsys appendfs.FS,
	logger ocfllogger.OCFLLogger,
) (object.Object, error) {
	f := factoryimpl.NewFactory(ver, extensionFactory, logger)
	obj := f.NewObject(ctx).WithExtensionManager(manager)
	initializer := obj.GetInitializer(fsys)

	if err := initializer.Init(id, digest, fixity); err != nil {
		return nil, errors.Wrap(err, "cannot initialize object")
	}

	if id != "" && obj.GetID() != id {
		return nil, fmt.Errorf("id mismatch. '%s' != '%s'", id, obj.GetID())
	}

	return obj, nil
}

func LoadObject(
	ctx context.Context,
	sourceFS fs.FS,
	extensionFactory *extensionimpl.Factory,
	logger ocfllogger.OCFLLogger,
) (object.Object, error) {
	// get the version of the object
	ver, err := util.GetObjectVersion(sourceFS)
	if err != nil {
		if errors.Is(err, ocflerrors.ErrVersionNone) {
			logger.ValidationError(validation.E003, "no version in fsys '%v'", sourceFS)
			return nil, ocflerrors.ErrVersionNone
		} else {
			return nil, errors.Wrapf(err, "getting version from fsys '%v'", sourceFS)
		}
	}
	// logger needs to know the new version
	logger.WithVersion(ver)
	fact := factoryimpl.NewFactory(ver, extensionFactory, logger)
	obj := fact.NewObject(ctx)
	loader := obj.GetLoader(sourceFS, extensionFactory)
	// load the object
	if err := loader.Load(); err != nil {
		return nil, errors.Wrapf(err, "cannot load object from fsys '%v'", sourceFS)
	}

	return obj, nil
}

func CheckObject(ctx context.Context, objectFS fs.FS, extensionFactory *extensionimpl.Factory, logger ocfllogger.OCFLLogger) error {
	fmt.Printf("object folder '%v'\n", objectFS)
	obj, err := LoadObject(ctx, objectFS, extensionFactory, logger)
	if err != nil {
		logger.ValidationError(validation.E001, "invalid fsys '%v': %v", objectFS, err)
		return errors.Wrapf(err, "cannot load object from folder '%v'", objectFS)
	}
	checker := obj.GetChecker(objectFS)
	if err := checker.Check(); err != nil {
		return errors.Wrapf(err, "cannot check object from folder '%v'", objectFS)
	}
	return nil
}

func Extract(ctx context.Context, objectFS fs.FS, destFS appendfs.FS, path string, version *inventory.VersionNumber, withManifest bool, area string, extensionFactory *extensionimpl.Factory, logger ocfllogger.OCFLLogger) error {
	if !version.IsValid() {
		version = inventory.NewVersionNumber().WithLatest()
	}

	logger.Debug().Msgf("Extracting object '%s' with version '%s'", path, version)
	var o object.Object
	var err error
	objFsys, err := writefs.Sub(objectFS, path)
	if err != nil {
		return errors.Wrapf(err, "cannot create subfs  '%v' / %s", objectFS, path)
	}
	o, err = LoadObject(ctx, objFsys, extensionFactory, logger)
	if err != nil {
		return errors.Wrapf(err, "cannot load object '%s'", path)
	}
	extractor := o.GetExtractor(objFsys, destFS)
	if err := extractor.Extract(version, withManifest, area); err != nil {
		return errors.Wrapf(err, "cannot extract object '%s'", path)
	}

	logger.Debug().Msgf("extraction done")
	return nil
}

func ExtractMeta(ctx context.Context, fsys fs.FS, path string, extensionFactory *extensionimpl.Factory, logger ocfllogger.OCFLLogger) (*inventory.Metadata, error) {
	logger.Debug().Msgf("Extracting object '%s'", path)
	objFsys, err := writefs.Sub(fsys, path)
	obj, err := LoadObject(ctx, objFsys, extensionFactory, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object '%s'", path)
	}
	defer logger.Debug().Msgf("extraction done")
	extractor := obj.GetExtractor(objFsys, nil)
	return extractor.GetMetadata()
}
