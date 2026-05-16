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

// InitStorageRoot initializes a new OCFL storage root with the given version at fsys.
func InitStorageRoot(ctx context.Context, fsys appendfs.FS, extensionConfigFS fs.FS, ver version.OCFLVersion, digest checksum.DigestAlgorithm, params map[string]string, logger ocfllogger.OCFLLogger) (storageroot.StorageRoot, error) {
	extManager, extFactory, err := SetupExtensionManager[storageroot.ExtensionManager](params, extensionConfigFS, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot setup extension manager")
	}
	srFactory := NewFactoryStorageRoot(ver, extFactory, logger)
	sr := srFactory.NewStorageRoot(ctx).
		WithWriteFS(fsys).
		WithExtensionManager(extManager).
		WithDigestAlgorithm(digest)

	initializer := sr.GetInitializer()
	if err := initializer.Init(); err != nil {
		return nil, errors.Wrap(err, "cannot initialize storage root")
	}
	return sr, nil
}

// InitObject initializes a new OCFL object with the given version, id and digest algorithm at fsys.
func InitObject(ctx context.Context, fsys appendfs.FS, extensionConfigFS fs.FS, ver version.OCFLVersion, id string, digest checksum.DigestAlgorithm, extensionParams map[string]string, logger ocfllogger.OCFLLogger) (object.Object, error) {
	objExtManager, objExtFactory, err := SetupExtensionManager[object.ExtensionManager](extensionParams, extensionConfigFS, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot setup extension manager")
	}
	objFactory := NewFactoryObject(ver, objExtFactory, logger)
	obj := objFactory.NewObject(ctx).
		WithExtensionManager(objExtManager).
		WithWriteFS(fsys)

	initializer := obj.GetInitializer()
	if err := initializer.Init(id, digest, nil); err != nil {
		return nil, errors.Wrap(err, "cannot initialize object")
	}
	return obj, nil
}
