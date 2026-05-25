package objectimpl

import (
	"io"
	"io/fs"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewVersionFSMap(logger ocfllogger.OCFLLogger) object.VersionFSMap {
	vfsm := &versionFSMap{
		ma:     map[int]fs.FS{},
		logger: logger.With("task", "versionFSMap"),
	}
	vfsm.loadVersionFS = vfsm._loadVersionFS
	return vfsm
}

type versionFSMap struct {
	ma            map[int]fs.FS
	baseFS        fs.FS
	loadVersionFS func(*inventory.VersionNumber) error
	logger        ocfllogger.OCFLLogger
}

func (v *versionFSMap) WithBaseFS(baseFS fs.FS) object.VersionFSMap {
	v.baseFS = baseFS
	return v
}

func (v *versionFSMap) _loadVersionFS(version *inventory.VersionNumber) error {
	subFS, err := fs.Sub(v.baseFS, version.String())
	if err != nil {
		return errors.Wrapf(err, "failed to create version FS %v/%s", v.baseFS, version)
	}
	v.ma[version.Int()] = subFS
	return nil
}

func (v *versionFSMap) GetVersionFS(version *inventory.VersionNumber) (fs.FS, error) {
	if v.baseFS == nil {
		return nil, errors.New("baseFS not set")
	}
	if version == nil {
		return nil, errors.New("version is nil")
	}
	if version.Int() <= 0 {
		return nil, errors.New("version number cannot be zero or negative")
	}
	if _, ok := v.ma[version.Int()]; !ok {
		if err := v.loadVersionFS(version); err != nil {
			return nil, errors.Wrapf(err, "cannot load version FS for '%s'", version)
		}
	}
	return v.ma[version.Int()], nil
}

func (v *versionFSMap) Close() error {
	var errs []error
	for k, f := range v.ma {
		if closer, ok := f.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, errors.Wrapf(err, "cannot close FS '%d'", k))
			}
		}
	}
	return errors.Combine(errs...)
}

var _ object.VersionFSMap = (*versionFSMap)(nil)
