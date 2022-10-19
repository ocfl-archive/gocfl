package ocfl

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension/storageroot"
	"regexp"
)

type OCFLVersion string

type StorageRoot interface {
	GetFiles() ([]string, error)
	GetFolders() ([]string, error)
	GetObjectFolders() ([]string, error)
	OpenObjectFolder(folder string) (Object, error)
	OpenObject(id string) (Object, error)
	Check() error
}

var OCFLVersionRegexp = regexp.MustCompile("^0=ocfl_([0-9]+\\.[0-9]+)$")

func NewStorageRoot(ctx context.Context, fs OCFLFS, defaultVersion OCFLVersion, defaultStorageLayout storageroot.StorageLayout, logger *logging.Logger) (StorageRoot, error) {
	version, err := getVersion(ctx, fs, ".", "ocfl_")
	if err != nil && err != errVersionNone {
		return nil, errors.WithStack(err)
	}
	if version == "" {
		cnt, err := fs.ReadDir(".")
		if err != nil {
			return nil, errors.Wrap(err, "cannot read storage root directory")
		}
		if len(cnt) > 0 {
			return nil, errors.WithStack(GetValidationError(defaultVersion, E069))
		}
		version = defaultVersion
	}
	switch version {
	case Version1_0:
		sr, err := NewStorageRootV1_0(ctx, fs, defaultStorageLayout, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	case Version1_1:
		sr, err := NewStorageRootV1_1(ctx, fs, defaultStorageLayout, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	default:
		return nil, errors.New(fmt.Sprintf("Storage Root Version %s not supported", version))
	}
}
