package initocfl

import (
	"context"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v4/pkg/appendfs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/ocflerrors"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func LoadStorageRoot(ctx context.Context, fsys fs.FS, logger ocfllogger.OCFLLogger) (storageroot.StorageRoot, error) {
	ver, err := util.GetStorageRootVersion(fsys)
	if err != nil {
		if errors.Is(err, ocflerrors.ErrInvalidContent) {
			logger.ValidationError(validation.E007, "invalid version content in fsys '%v'", fsys)
			ver = version.Default
		} else if errors.Is(err, ocflerrors.ErrVersionNone) {
			logger.ValidationError(validation.E003, "no version in fsys '%v'", fsys)
			ver = version.Default
		} else {
			return nil, errors.Wrapf(err, "getting version from fsys '%v'", fsys)
		}
	}

	extFS, err := fs.Sub(fsys, "extensions")
	if err != nil {
		return nil, errors.Wrap(err, "cannot get extensions subfs")
	}
	_, srExtFactory, err := SetupExtensionManager[storageroot.ExtensionManager](nil, extFS, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot setup extension manager")
	}
	srFactory := NewFactoryStorageRoot(ver, srExtFactory, logger)
	sr := srFactory.NewStorageRoot(ctx).
		WithReadFS(fsys).
		WithDigestAlgorithm(checksum.DigestSHA512)

	if writeFS, ok := fsys.(appendfs.FS); ok {
		sr = sr.WithWriteFS(writeFS)
	}

	loader := sr.GetLoader()
	if err := loader.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return sr, nil
}

func LoadObject(ctx context.Context, fsys fs.FS, logger ocfllogger.OCFLLogger) (object.Object, error) {
	ver, err := util.GetObjectVersion(fsys)
	if err != nil {
		if errors.Is(err, ocflerrors.ErrInvalidContent) {
			logger.ValidationError(validation.E007, "invalid version content in fsys '%v'", fsys)
			ver = version.Default
		} else if errors.Is(err, ocflerrors.ErrVersionNone) {
			logger.ValidationError(validation.E003, "no version in fsys '%v'", fsys)
			ver = version.Default
		} else {
			return nil, errors.Wrapf(err, "getting version from fsys '%v'", fsys)
		}
	}
	// logger needs to know the version
	logger.WithVersion(ver)

	extFS, err := fs.Sub(fsys, "extensions")
	if err != nil {
		return nil, errors.Wrap(err, "cannot get extensions subfs")
	}
	objExtManager, objExtFactory, err := SetupExtensionManager[object.ExtensionManager](nil, extFS, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot setup extension manager")
	}
	objFactory := NewFactoryObject(ver, objExtFactory, logger)
	obj := objFactory.NewObject(ctx).
		WithExtensionManager(objExtManager).
		WithReadFS(fsys)

	if writeFS, ok := fsys.(appendfs.FS); ok {
		obj = obj.WithWriteFS(writeFS)
	}

	loader := obj.GetLoader()
	if err := loader.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load object")
	}
	return obj, nil
}
