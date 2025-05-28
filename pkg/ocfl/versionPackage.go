package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
	"io"
)

type VersionPackages interface {
	GetSpec() VersionPackagesSpec
	GetDigestAlgorithm() checksum.DigestAlgorithm
}

type VersionPackageWriter interface {
	addReader(r io.ReadCloser, names *NamesStruct, noExtensionHook bool) (string, error)
	GetObject() *ObjectBase
	GetType() VersionPackageType
	Close() error
	Version() string
}

type VersionPackageReader interface {
	GetObject() *ObjectBase
	GetType() VersionPackageType
	Close() error
}

func newVersionPackage(
	ctx context.Context,
	object Object,
	folder string,
	version OCFLVersion,
	logger zLogger.ZLogger,
	errorFactory *archiveerror.Factory,
) (VersionPackages, error) {
	switch version {
	case Version2_0:
		sr, err := newVersionPackageV2_0(logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	default:
		sr := newVersionPackageBase(logger)
		return sr, nil
	}
}
