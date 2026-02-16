package storagerootimpl

import (
	"context"
	"fmt"
	"io/fs"
	"regexp"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/extensionimpl"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/ocflerrors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

var OCFLVersionRegexp = regexp.MustCompile("^0=ocfl_([0-9]+\\.[0-9]+)$")

func newStorageRoot(ctx context.Context, fsys fs.FS, ver version.OCFLVersion, extensionFactory *extensionimpl.ExtensionFactory, extensionManager ExtensionManager, logger zLogger.ZLogger) (storageroot.StorageRoot, error) {
	switch ver {
	case version.Version1_0:
		sr, err := NewStorageRootV1_0(ctx, fsys, extensionFactory, extensionManager, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	case version.Version1_1:
		sr, err := NewStorageRootV1_1(ctx, fsys, extensionFactory, extensionManager, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	case version.Version2_0:
		sr, err := NewStorageRootV2_0(ctx, fsys, extensionFactory, extensionManager, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	default:
		return nil, errors.New(fmt.Sprintf("Storage Root Version %s not supported", ver))
	}
}

func ValidVersion(ver version.OCFLVersion) bool {
	switch ver {
	case version.Version1_0:
		return true
	case version.Version1_1:
		return true
	case version.Version2_0:
		return true
	default:
		return false
	}
}

func CreateStorageRoot(ctx context.Context, fsys fs.FS, ver version.OCFLVersion, extensionFactory *extensionimpl.ExtensionFactory, extensionManager ExtensionManager, digest checksum.DigestAlgorithm, logger zLogger.ZLogger) (storageroot.StorageRoot, error) {
	storageRoot, err := newStorageRoot(ctx, fsys, ver, extensionFactory, extensionManager, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate storage root")
	}

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
	extFSys, err := writefs.Sub(fsys, "extensions")
	if err != nil {
		return nil, errors.Wrap(err, "cannot create sub filesystem 'extensions'")
	}
	extensionManager, err := extensionFactory.CreateExtensions(extFSys, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create extension manager")
	}
	storageRoot, err := newStorageRoot(ctx, fsys, ver, extensionFactory, extensionManager.(ExtensionManager), logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate storage root")
	}

	if err := storageRoot.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return storageRoot, nil
}

func LoadStorageRootRO(ctx context.Context, fsys fs.FS, extensionFactory *extensionimpl.ExtensionFactory, logger zLogger.ZLogger) (storageroot.StorageRoot, error) {
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
	extFSys, err := writefs.Sub(fsys, "extensions")
	if err != nil {
		return nil, errors.Wrap(err, "cannot create sub filesystem 'extensions'")
	}
	extensionManager, err := extensionFactory.CreateExtensions(extFSys, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create extension manager")
	}
	storageRoot, err := newStorageRoot(ctx, fsys, ver, extensionFactory, extensionManager.(ExtensionManager), logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate storage root")
	}

	if err := storageRoot.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return storageRoot, nil
}
