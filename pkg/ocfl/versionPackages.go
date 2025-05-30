package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
	"io"
	"io/fs"
)

type VersionPackages interface {
	Init(digest checksum.DigestAlgorithm) error
	GetSpec() VersionPackagesSpec
	GetDigestAlgorithm() checksum.DigestAlgorithm
	IsEmpty() bool
	AddVersion(version string, versionType VersionPackagesType, versionTypeVersion string, files map[string]string) error
	GetFS(version string, object Object) (fs.FS, io.Closer, error)
	GetVersion(version string) (v *PackageVersionBase, ok bool)
}

type VersionPackageWriter interface {
	addReader(r io.ReadCloser, names *NamesStruct, noExtensionHook bool) (string, error)
	WriteFile(name string, r io.Reader) (int64, error)
	GetObject() *ObjectBase
	Type() VersionPackagesType
	Close() error
	Version() string
	GetFileDigest() (map[string]string, error)
}

type VersionPackageReader interface {
	GetObject() *ObjectBase
	GetType() VersionPackagesType
	Close() error
}

func newVersionPackage(
	ctx context.Context,
	object Object,
	folder string,
	logger zLogger.ZLogger,
	errorFactory *archiveerror.Factory,
) (VersionPackages, error) {
	switch object.GetVersion() {
	case Version2_0:
		sr, err := newVersionPackageV2_0(object.GetInventory().GetDigestAlgorithm(), logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	default:
		return nil, nil
	}
}
