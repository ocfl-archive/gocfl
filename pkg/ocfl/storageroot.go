package ocfl

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/checksum"
	"github.com/op/go-logging"
	"io"
	"regexp"
)

type OCFLVersion string

type StorageRoot interface {
	GetDigest() checksum.DigestAlgorithm
	SetDigest(digest checksum.DigestAlgorithm)
	GetFiles() ([]string, error)
	GetFolders() ([]string, error)
	GetObjectFolders() ([]string, error)
	ObjectExists(id string) (bool, error)
	LoadObjectByFolder(folder string) (Object, error)
	LoadObjectByID(id string) (Object, error)
	CreateObject(id string, version OCFLVersion, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, defaultExtensions []Extension) (Object, error)
	CreateExtension(fs OCFLFSRead) (Extension, error)
	Check() error
	CheckObjectByFolder(objectFolder string) error
	CheckObjectByID(objectID string) error
	Init(version OCFLVersion, digest checksum.DigestAlgorithm, exts []Extension) error
	Load() error
	IsModified() bool
	setModified()
	GetVersion() OCFLVersion
	Stat(w io.Writer, path string, id string, statInfo []StatInfo) error
	Extract(fs OCFLFS, path, id, version string, withManifest bool) error
}

var OCFLVersionRegexp = regexp.MustCompile("^0=ocfl_([0-9]+\\.[0-9]+)$")

func newStorageRoot(ctx context.Context, fs OCFLFSRead, version OCFLVersion, extensionFactory *ExtensionFactory, logger *logging.Logger) (StorageRoot, error) {
	switch version {
	case Version1_0:
		sr, err := NewStorageRootV1_0(ctx, fs, extensionFactory, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	case Version1_1:
		sr, err := NewStorageRootV1_1(ctx, fs, extensionFactory, logger)
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
	default:
		return false
	}
}

func CreateStorageRoot(ctx context.Context, fs OCFLFS, version OCFLVersion, extensionFactory *ExtensionFactory, defaultExtensions []Extension, digest checksum.DigestAlgorithm, logger *logging.Logger) (StorageRoot, error) {
	storageRoot, err := newStorageRoot(ctx, fs, version, extensionFactory, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate storage root")
	}

	if err := storageRoot.Init(version, digest, defaultExtensions); err != nil {
		return nil, errors.Wrap(err, "cannot initialize storage root")
	}

	return storageRoot, nil
}

func LoadStorageRoot(ctx context.Context, fs OCFLFSRead, extensionFactory *ExtensionFactory, logger *logging.Logger) (StorageRoot, error) {
	version, err := getVersion(ctx, fs, ".", "ocfl_")
	if err != nil && err != errVersionNone {
		return nil, errors.WithStack(err)
	}
	if version == "" {
		if fs.HasContent() {
			err := GetValidationError(Version1_1, E069).AppendDescription("storage root not empty without version information").AppendContext("storage root '%s'", fs)
			addValidationErrors(ctx, err)
			//			return nil, err
		}
		version = Version1_1
	}
	storageRoot, err := newStorageRoot(ctx, fs, version, extensionFactory, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot instantiate storage root")
	}

	if err := storageRoot.Load(); err != nil {
		return nil, errors.Wrap(err, "cannot load storage root")
	}
	return storageRoot, nil
}

func LoadStorageRootRO(ctx context.Context, fs OCFLFSRead, extensionFactory *ExtensionFactory, logger *logging.Logger) (StorageRoot, error) {
	version, err := getVersion(ctx, fs, ".", "ocfl_")
	if err != nil && err != errVersionNone {
		return nil, errors.WithStack(err)
	}
	if version == "" {
		if fs.HasContent() {
			err := GetValidationError(Version1_1, E069).AppendDescription("storage root not empty without version information").AppendContext("storage root '%s'", fs)
			addValidationErrors(ctx, err)
			//			return nil, err
		}
		version = Version1_1
	}
	storageRoot, err := newStorageRoot(ctx, fs, version, extensionFactory, logger)
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
