package ocfl

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"io"
	"io/fs"
	"regexp"
)

type OCFLVersion string

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
	CreateObject(id string, version OCFLVersion, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, manager ExtensionManager) (Object, error)
	CreateExtension(fsys fs.FS) (Extension, error)
	CreateExtensions(fsys fs.FS, validation Validation) (ExtensionManager, error)
	Check() error
	CheckObjectByFolder(objectFolder string) error
	CheckObjectByID(objectID string) error
	Init(version OCFLVersion, digest checksum.DigestAlgorithm, manager ExtensionManager) error
	Load() error
	IsModified() bool
	setModified()
	GetVersion() OCFLVersion
	Stat(w io.Writer, path string, id string, statInfo []StatInfo) error
	Extract(fsys fs.FS, path, id, version string, withManifest bool, area string) error
	ExtractMeta(path, id string) (*StorageRootMetadata, error)
}

var OCFLVersionRegexp = regexp.MustCompile("^0=ocfl_([0-9]+\\.[0-9]+)$")

func newStorageRoot(ctx context.Context, fsys fs.FS, version OCFLVersion, extensionFactory *ExtensionFactory, extensionManager ExtensionManager, logger zLogger.ZLogger) (StorageRoot, error) {
	switch version {
	case Version1_0:
		sr, err := NewStorageRootV1_0(ctx, fsys, extensionFactory, extensionManager, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	case Version1_1:
		sr, err := NewStorageRootV1_1(ctx, fsys, extensionFactory, extensionManager, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	case Version2_0:
		sr, err := NewStorageRootV2_0(ctx, fsys, extensionFactory, extensionManager, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	default:
		return nil, errors.New(fmt.Sprintf("Storage Root Version %s not supported", version))
	}
}

func ValidVersion(version OCFLVersion) bool {
	switch version {
	case Version1_0:
		return true
	case Version1_1:
		return true
	case Version2_0:
		return true
	default:
		return false
	}
}

func CreateStorageRoot(ctx context.Context, fsys fs.FS, version OCFLVersion, extensionFactory *ExtensionFactory, extensionManager ExtensionManager, digest checksum.DigestAlgorithm, logger zLogger.ZLogger) (StorageRoot, error) {
	storageRoot, err := newStorageRoot(ctx, fsys, version, extensionFactory, extensionManager, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate storage root")
	}

	if err := storageRoot.Init(version, digest, extensionManager); err != nil {
		return nil, errors.Wrap(err, "cannot initialize storage root")
	}

	return storageRoot, nil
}

func LoadStorageRoot(ctx context.Context, fsys fs.FS, extensionFactory *ExtensionFactory, logger zLogger.ZLogger) (StorageRoot, error) {
	version, err := getVersion(ctx, fsys, ".", "ocfl_")
	if err != nil && !errors.Is(err, errVersionNone) {
		return nil, errors.WithStack(err)
	}
	if version == "" {
		dirs, err := fs.ReadDir(fsys, ".")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if len(dirs) > 0 {
			err := GetValidationError(Version1_1, E069).AppendDescription("storage root not empty without version information").AppendContext("storage root '%s'", fsys)
			addValidationErrors(ctx, err)
			//			return nil, err
		}
		version = Version1_1
	}
	extFSys, err := fs.Sub(fsys, "extensions")
	if err != nil {
		return nil, errors.Wrap(err, "cannot create sub filesystem 'extensions'")
	}
	extensionManager, err := extensionFactory.CreateExtensions(extFSys, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create extension manager")
	}
	storageRoot, err := newStorageRoot(ctx, fsys, version, extensionFactory, extensionManager, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate storage root")
	}

	if err := storageRoot.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return storageRoot, nil
}

func LoadStorageRootRO(ctx context.Context, fsys fs.FS, extensionFactory *ExtensionFactory, logger zLogger.ZLogger) (StorageRoot, error) {
	version, err := getVersion(ctx, fsys, ".", "ocfl_")
	if err != nil && err != errVersionNone {
		return nil, errors.WithStack(err)
	}
	if version == "" {
		dirs, err := fs.ReadDir(fsys, ".")
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if len(dirs) > 0 {
			err := GetValidationError(Version1_1, E069).AppendDescription("storage root not empty without version information").AppendContext("storage root '%s'", fsys)
			addValidationErrors(ctx, err)
			//			return nil, err
		}
		version = Version1_1
	}
	extFSys, err := fs.Sub(fsys, "extensions")
	if err != nil {
		return nil, errors.Wrap(err, "cannot create sub filesystem 'extensions'")
	}
	extensionManager, err := extensionFactory.CreateExtensions(extFSys, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create extension manager")
	}
	storageRoot, err := newStorageRoot(ctx, fsys, version, extensionFactory, extensionManager, logger)
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
