package baseFS

import (
	"emperror.dev/errors"
	"io/fs"
	"strings"
)

type CreateFS func(path string) (fs.FS, error)

type Factory struct {
	fss map[string]CreateFS
}

func NewFactory() (*Factory, error) {
	f := &Factory{fss: map[string]CreateFS{}}
	return f, nil
}

func (f *Factory) Add(fun CreateFS, prefix string) {
	f.fss[prefix] = fun
}

func (f *Factory) Get(path string) (fs.FS, error) {
	if strings.HasSuffix(path, ".zip") {
		create, ok := f.fss["zip://"]
		if !ok {
			return nil, errors.Errorf("%s - zip not supported", path)
		}
		fsys, err := create(path)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create filesystem for '%s'", path)
		}
		return fsys, nil
	}
	for prefix, create := range f.fss {
		if strings.HasPrefix(path, prefix) {
			fsys, err := create(path)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot create filesystem for '%s'", path)
			}
			return fsys, nil
		}
	}
	return nil, errors.Errorf("path %s not supported", path)
}
