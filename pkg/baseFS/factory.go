package baseFS

import (
	"emperror.dev/errors"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
)

type Factory struct {
	fss []FS
}

func NewFactory() (*Factory, error) {
	f := &Factory{fss: []FS{}}
	return f, nil
}

func (f *Factory) Add(fs FS) {
	f.fss = append(f.fss, fs)
}

func (f *Factory) GetFSRW(path string) (ocfl.OCFLFS, error) {
	for _, fsys := range f.fss {
		if fsys.Valid(path) {
			return fsys.GetFSRW(path)
		}
	}
	return nil, errors.Errorf("no filesystem for '%s' found", path)
}

func (f *Factory) GetFS(path string) (ocfl.OCFLFSRead, error) {
	for _, fsys := range f.fss {
		if fsys.Valid(path) {
			return fsys.GetFS(path)
		}
	}
	return nil, errors.Errorf("no filesystem for '%s' found", path)
}

func (f *Factory) Open(path string) (fs.File, error) {
	for _, fsys := range f.fss {
		if fsys.Valid(path) {
			return fsys.Open(path)
		}
	}
	return nil, errors.Errorf("no filesystem for '%s' found", path)
}

func (f *Factory) Create(path string) (io.WriteCloser, error) {
	for _, fsys := range f.fss {
		if fsys.Valid(path) {
			return fsys.Create(path)
		}
	}
	return nil, errors.Errorf("no filesystem for '%s' found", path)
}
