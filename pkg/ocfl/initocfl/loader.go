package initocfl

import (
	"context"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v4/pkg/appendfs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func LoadStorageRoot(ctx context.Context, fsys fs.FS, ver version.OCFLVersion, logger ocfllogger.OCFLLogger) (storageroot.StorageRoot, error) {
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

func LoadObject(ctx context.Context, fsys fs.FS, ver version.OCFLVersion, logger ocfllogger.OCFLLogger) (object.Object, error) {
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
