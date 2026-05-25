package objectimpl

import (
	"io/fs"

	"emperror.dev/errors"
	"github.com/ocfl-archive/filesystem/pkg/zipfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

func NewVersionFSMap20(logger ocfllogger.OCFLLogger) object.VersionFSMap {
	vfsm := &versionFSMap20{
		versionFSMap: &versionFSMap{
			ma:     map[string]fs.FS{},
			logger: logger.With("task", "versionFSMap20"),
		},
		logger: logger,
	}
	vfsm.loadVersionFS = vfsm._loadVersionFS
	return vfsm
}

type versionFSMap20 struct {
	*versionFSMap
	logger ocfllogger.OCFLLogger
}

func (v *versionFSMap20) _loadVersionFS(version string) error {
	fi, err := fs.Stat(v.baseFS, version)
	if err == nil && fi.IsDir() {
		return v.versionFSMap._loadVersionFS(version)
	}
	fi, err = fs.Stat(v.baseFS, version+".zip")
	if err == nil && !fi.IsDir() {
		subFS, err := zipfs.NewFSFile(v.baseFS, version+".zip", v.logger.Logger())
		if err != nil {
			return errors.Wrapf(err, "cannot open zip file %s.zip", version)
		}
		v.ma[version] = subFS
		return nil
	}
	return errors.Errorf("folders %s and %s.zip not found", version, version)
}

var _ object.VersionFSMap = (*versionFSMap20)(nil)
