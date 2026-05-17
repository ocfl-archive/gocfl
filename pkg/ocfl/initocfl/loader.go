// Package initocfl provides functions for loading and initializing OCFL storage roots and objects.
// It acts as a high-level entry point for working with different OCFL versions
// by orchestrating the creation of appropriate factories and managers.
package initocfl

import (
	"context"
	"io"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/ocflerrors"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

type closeFunc func() error

func (f closeFunc) Close() error { return f() } // Muss groß sein!

// LoadStorageRoot detects the OCFL version of the storage root at fsys,
// initializes the extension manager, and loads the storage root structure.
// It returns a storageroot.StorageRoot instance.
func LoadStorageRoot(ctx context.Context, fsys fs.FS, extensionParams map[string]string, logger ocfllogger.OCFLLogger) (storageroot.StorageRoot, io.Closer, error) {
	ver, err := util.GetStorageRootVersion(fsys)
	if err != nil {
		if errors.Is(err, ocflerrors.ErrInvalidContent) {
			logger.ValidationError(validation.E007, "invalid version content in fsys '%v'", fsys)
			ver = version.Default
		} else if errors.Is(err, ocflerrors.ErrVersionNone) {
			logger.ValidationError(validation.E003, "no version in fsys '%v'", fsys)
			ver = version.Default
		} else {
			return nil, nil, errors.Wrapf(err, "getting version from fsys '%v'", fsys)
		}
	}

	extFS, err := fs.Sub(fsys, "extensions")
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot get extensions subfs")
	}
	srExtManager, srExtFactory, err := SetupExtensionManager[storageroot.ExtensionManager](extensionParams, extFS, logger)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot setup extension manager")
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
		return nil, nil, errors.Wrap(err, "cannot load storage root")
	}
	return sr, closeFunc(func() error {
		return errors.WithStack(srExtManager.Terminate())
	}), nil
}

// LoadObject detects the OCFL version of the object at fsys,
// initializes the extension manager, and loads the object structure.
// It returns an object.Object instance.
func LoadObject(ctx context.Context, fsys fs.FS, extensionParams map[string]string, logger ocfllogger.OCFLLogger) (object.Object, io.Closer, error) {
	ver, err := util.GetObjectVersion(fsys)
	if err != nil {
		if errors.Is(err, ocflerrors.ErrInvalidContent) {
			logger.ValidationError(validation.E007, "invalid version content in fsys '%v'", fsys)
			ver = version.Default
		} else if errors.Is(err, ocflerrors.ErrVersionNone) {
			logger.ValidationError(validation.E003, "no version in fsys '%v'", fsys)
			ver = version.Default
		} else {
			return nil, nil, errors.Wrapf(err, "getting version from fsys '%v'", fsys)
		}
	}
	// logger needs to know the version
	logger.WithVersion(ver)

	extFS, err := fs.Sub(fsys, "extensions")
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot get extensions subfs")
	}
	objExtManager, objExtFactory, err := SetupExtensionManager[object.ExtensionManager](extensionParams, extFS, logger)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot setup extension manager")
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
		return nil, nil, errors.Wrap(err, "cannot load object")
	}
	return obj, closeFunc(func() error {
		return errors.WithStack(objExtManager.Terminate())
	}), nil
}
