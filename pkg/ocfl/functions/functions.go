package functions

import (
	"context"
	"fmt"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v4/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/initocfl"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/ocflerrors"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func CreateObject(ctx context.Context, id string, ver version.OCFLVersion, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, extensionFactory extension.Factory[object.ExtensionManager], manager object.ExtensionManager, fsys appendfs.FS, logger ocfllogger.OCFLLogger) (object.Object, error) {
	f := initocfl.NewFactoryObject(ver, extensionFactory, logger)
	obj := f.NewObject(ctx).WithExtensionManager(manager)
	initializer := obj.GetInitializer().WithFS(fsys)

	if err := initializer.Init(id, digest, fixity); err != nil {
		return nil, errors.Wrap(err, "cannot initialize object")
	}

	if id != "" && obj.GetID() != id {
		return nil, fmt.Errorf("id mismatch. '%s' != '%s'", id, obj.GetID())
	}

	return obj, nil
}

func LoadObjectByID(sr storageroot.StorageRoot, extensionFactory extension.Factory[object.ExtensionManager], id string, logger ocfllogger.OCFLLogger) (object.Object, error) {
	folder, err := sr.IdToFolder(id)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object %s", id)
	}
	var ofs fs.FS = sr.GetWriteFS()
	if ofs == nil {
		ofs = sr.GetReadFS()
	}
	fsys, err := writefs.Sub(sr.GetReadFS(), folder)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create subfs for %v / %s", sr.GetReadFS(), folder)
	}
	obj, err := LoadObjectFS(context.Background(), fsys, extensionFactory, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object %s", id)
	}
	return obj, nil
}

func LoadObjectFS(ctx context.Context, objectFS fs.FS, extensionFactory extension.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) (object.Object, error) {
	// get the version of the object
	ver, err := util.GetObjectVersion(objectFS)
	if err != nil {
		if errors.Is(err, ocflerrors.ErrInvalidContent) {
			logger.ValidationError(validation.E007, "invalid version content in fsys '%v'", objectFS)
			ver = version.Default
		}
		if errors.Is(err, ocflerrors.ErrVersionNone) {
			logger.ValidationError(validation.E003, "no version in fsys '%v'", objectFS)
			ver = version.Default
			//return nil, ocflerrors.ErrVersionNone
		} else {
			return nil, errors.Wrapf(err, "getting version from fsys '%v'", objectFS)
		}
	}
	// logger needs to know the new version
	logger.WithVersion(ver)
	fact := initocfl.NewFactoryObject(ver, extensionFactory, logger)
	obj := fact.NewObject(ctx)
	loader := obj.GetLoader().WithFS(objectFS)
	// load the object
	if err := loader.Load(); err != nil {
		return nil, errors.Wrapf(err, "cannot load object from fsys '%v'", objectFS)
	}

	return obj, nil
}

func CheckObject(ctx context.Context, objectFS fs.FS, extensionFactory extension.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) error {
	fmt.Printf("object folder '%v'\n", objectFS)
	obj, err := LoadObjectFS(ctx, objectFS, extensionFactory, logger)
	if err != nil {
		logger.ValidationError(validation.E001, "invalid fsys '%v': %v", objectFS, err)
		return errors.Wrapf(err, "cannot load object from folder '%v'", objectFS)
	}
	checker := obj.GetChecker()
	if err := checker.Check(); err != nil {
		return errors.Wrapf(err, "cannot check object from folder '%v'", objectFS)
	}
	return nil
}

func Extract(ctx context.Context, objectFS fs.FS, destFS appendfs.FS, path string, version *inventory.VersionNumber, withManifest bool, area string, extensionFactory extension.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) error {
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
	o, err = LoadObjectFS(ctx, objFsys, extensionFactory, logger)
	if err != nil {
		return errors.Wrapf(err, "cannot load object '%s'", path)
	}
	extractor := o.GetExtractor().WithFS(objFsys, destFS)
	if err := extractor.Extract(version, withManifest, area); err != nil {
		return errors.Wrapf(err, "cannot extract object '%s'", path)
	}

	logger.Debug().Msgf("extraction done")
	return nil
}

func ExtractMeta(ctx context.Context, fsys fs.FS, path string, extensionFactory extension.Factory[object.ExtensionManager], logger ocfllogger.OCFLLogger) (*inventory.Metadata, error) {
	logger.Debug().Msgf("Extracting object '%s'", path)
	objFsys, err := writefs.Sub(fsys, path)
	obj, err := LoadObjectFS(ctx, objFsys, extensionFactory, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load object '%s'", path)
	}
	defer logger.Debug().Msgf("extraction done")
	extractor := obj.GetExtractor().WithFS(objFsys, nil)
	return extractor.GetMetadata()
}
