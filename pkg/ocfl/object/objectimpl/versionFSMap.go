package objectimpl

import (
	"io"
	"io/fs"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewVersionFSMap(logger ocfllogger.OCFLLogger) object.VersionFSMap {
	vfsm := &versionFSMap{
		ma:     map[string]fs.FS{},
		logger: logger.With("task", "versionFSMap"),
	}
	vfsm.loadVersionFS = vfsm._loadVersionFS
	return vfsm
}

type versionFSMap struct {
	ma            map[string]fs.FS
	baseFS        fs.FS
	loadVersionFS func(string) error
	logger        ocfllogger.OCFLLogger
}

func (v *versionFSMap) WithBaseFS(baseFS fs.FS) object.VersionFSMap {
	v.baseFS = baseFS
	return v
}

func (v *versionFSMap) _loadVersionFS(version string) error {
	subFS, err := fs.Sub(v.baseFS, version)
	if err != nil {
		return errors.Wrapf(err, "failed to create version FS %v/%s", v.baseFS, version)
	}
	v.ma[version] = subFS
	return nil
}

func (v *versionFSMap) GetVersionFS(version string) (fs.FS, error) {
	if v.baseFS == nil {
		return nil, errors.New("baseFS not set")
	}
	if _, ok := v.ma[version]; !ok {
		if err := v.loadVersionFS(version); err != nil {
			return nil, errors.Wrapf(err, "cannot load version FS for '%s'", version)
		}
	}
	return v.ma[version], nil
}

func (v *versionFSMap) Close() error {
	var errs []error
	for k, f := range v.ma {
		if closer, ok := f.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, errors.Wrapf(err, "cannot close FS '%s'", k))
			}
		}
	}
	return errors.Combine(errs...)
}

var _ object.VersionFSMap = (*versionFSMap)(nil)
