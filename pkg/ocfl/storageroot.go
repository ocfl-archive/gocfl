package ocfl

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/ocflerrors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/util"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"

	"io"
	"io/fs"
	"regexp"
)

type StorageRoot interface {
	fmt.Stringer
	GetDigest() checksum.DigestAlgorithm
	SetDigest(digest checksum.DigestAlgorithm)
	GetFiles() ([]string, error)
	GetFolders() ([]string, error)
	GetObjectFolders() ([]string, error)
	ObjectExists(id string) (bool, error)
	LoadObjectByFolder(folder string) (Object, error)
	LoadObjectByID(id string) (Object, error)
	CreateObject(id string, ver version.OCFLVersion, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, manager ExtensionManager) (Object, error)
	CreateExtension(fsys fs.FS) (Extension, error)
	CreateExtensions(fsys fs.FS, validation validation.Validation) (ExtensionManager, error)
	Check() error
	CheckObjectByFolder(objectFolder string) error
	CheckObjectByID(objectID string) error
	Init(ver version.OCFLVersion, digest checksum.DigestAlgorithm, manager ExtensionManager) error
	Load() error
	IsModified() bool
	setModified()
	GetVersion() version.OCFLVersion
	Stat(w io.Writer, path string, id string, statInfo []StatInfo) error
	Extract(fsys fs.FS, path, id, version string, withManifest bool, area string) error
	ExtractMeta(path, id string) (*StorageRootMetadata, error)
}

var OCFLVersionRegexp = regexp.MustCompile("^0=ocfl_([0-9]+\\.[0-9]+)$")

func newStorageRoot(ctx context.Context, fsys fs.FS, ver version.OCFLVersion, extensionFactory *ExtensionFactory, extensionManager ExtensionManager, logger zLogger.ZLogger) (StorageRoot, error) {
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

func CreateStorageRoot(ctx context.Context, fsys fs.FS, ver version.OCFLVersion, extensionFactory *ExtensionFactory, extensionManager ExtensionManager, digest checksum.DigestAlgorithm, logger zLogger.ZLogger) (StorageRoot, error) {
	storageRoot, err := newStorageRoot(ctx, fsys, ver, extensionFactory, extensionManager, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate storage root")
	}

	if err := storageRoot.Init(ver, digest, extensionManager); err != nil {
		return nil, errors.Wrap(err, "cannot initialize storage root")
	}

	return storageRoot, nil
}

func LoadStorageRoot(ctx context.Context, fsys fs.FS, extensionFactory *ExtensionFactory, logger zLogger.ZLogger) (StorageRoot, error) {
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
	storageRoot, err := newStorageRoot(ctx, fsys, ver, extensionFactory, extensionManager, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate storage root")
	}

	if err := storageRoot.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return storageRoot, nil
}

func LoadStorageRootRO(ctx context.Context, fsys fs.FS, extensionFactory *ExtensionFactory, logger zLogger.ZLogger) (StorageRoot, error) {
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
	storageRoot, err := newStorageRoot(ctx, fsys, ver, extensionFactory, extensionManager, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate storage root")
	}

	if err := storageRoot.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return storageRoot, nil
}

var (
	_ StorageRoot = &StorageRootV1_0{}
	_ StorageRoot = &StorageRootV1_1{}
)
