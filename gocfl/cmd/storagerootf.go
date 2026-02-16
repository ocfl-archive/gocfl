package cmd

import (
	"context"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/factory/factoryimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/ocflerrors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot/storagerootimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func CreateStorageRoot(ctx context.Context, fsys fs.FS, ver version.OCFLVersion, extensionFactory *extensionimpl.ExtensionFactory, extensionManager storagerootimpl.ExtensionManager, digest checksum.DigestAlgorithm, logger zLogger.ZLogger) (storageroot.StorageRoot, error) {
	fact := factoryimpl.NewFactory(ver, extensionFactory, logger)
	storageRoot := fact.NewStorageRoot(ctx).WithFS(fsys)

	if err := storageRoot.Init(ver, digest, extensionManager); err != nil {
		return nil, errors.Wrap(err, "cannot initialize storage root")
	}

	return storageRoot, nil
}

func LoadStorageRoot(ctx context.Context, fsys fs.FS, extensionFactory *extensionimpl.ExtensionFactory, logger zLogger.ZLogger) (storageroot.StorageRoot, error) {
	ver, err := util.GetVersion(ctx, fsys, ".", "ocfl_")
	if err != nil && !errors.Is(err, ocflerrors.ErrVersionNone) {
		return nil, errors.WithStack(err)
	}
	if ver == "" {
		dirs, err := fs.ReadDir(fsys, ".")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if len(dirs) > 0 {
			err := validation.GetValidationError(version.Version1_1, validation.E069).AppendDescription("storage root not empty without version information").AppendContext("storage root '%s'", fsys)
			validation.AddValidationErrors(ctx, err)
			//			return nil, err
		}
		ver = version.Version1_1
	}
	fact := factoryimpl.NewFactory(ver, extensionFactory, logger)
	storageRoot := fact.NewStorageRoot(ctx).WithFS(fsys)
	if err := storageRoot.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return storageRoot, nil
}

func LoadStorageRootRO(ctx context.Context, fact factory.Factory, fsys fs.FS, extensionFactory *extensionimpl.ExtensionFactory, logger zLogger.ZLogger) (storageroot.StorageRoot, error) {
	ver, err := util.GetVersion(ctx, fsys, ".", "ocfl_")
	if err != nil && !errors.Is(err, ocflerrors.ErrVersionNone) {
		return nil, errors.WithStack(err)
	}
	if ver == "" {
		dirs, err := fs.ReadDir(fsys, ".")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if len(dirs) > 0 {
			err := validation.GetValidationError(version.Version1_1, validation.E069).AppendDescription("storage root not empty without version information").AppendContext("storage root '%s'", fsys)
			validation.
				AddValidationErrors(ctx, err)
			//			return nil, err
		}
		ver = version.Version1_1
	}
	storageRoot := fact.NewStorageRoot(ctx).WithFS(fsys)
	if err := storageRoot.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return storageRoot, nil
}
