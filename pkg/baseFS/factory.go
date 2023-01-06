package baseFS

import (
	"emperror.dev/errors"
	"github.com/je4/gocfl/v2/pkg/ocfl"
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
	fs.SetFSFactory(f)
	f.fss = append(f.fss, fs)
}

func (f *Factory) GetFSRW(path string) (ocfl.OCFLFS, error) {
	for _, fsys := range f.fss {
		ret, err := fsys.GetFSRW(path)
		if ErrNotSupported(err) {
			continue
		}
		return ret, err
	}
	return nil, ErrPathNotSupported
}

func (f *Factory) GetFS(path string) (ocfl.OCFLFSRead, error) {
	for _, fsys := range f.fss {
		ret, err := fsys.GetFS(path)
		if ErrNotSupported(err) {
			continue
		}
		return ret, err
	}
	return nil, ErrPathNotSupported
}

func (f *Factory) Open(path string) (fs.File, error) {
	for _, fsys := range f.fss {
		ret, err := fsys.Open(path)
		if ErrNotSupported(err) {
			continue
		}
		return ret, err
	}
	return nil, ErrPathNotSupported
}

func (f *Factory) Create(path string) (io.WriteCloser, error) {
	for _, fsys := range f.fss {
		ret, err := fsys.Create(path)
		if ErrNotSupported(err) {
			continue
		}
		return ret, err
	}
	return nil, ErrPathNotSupported
}

func (f *Factory) Delete(path string) error {
	for _, fsys := range f.fss {
		if err := fsys.Delete(path); err != nil {
			if ErrNotSupported(err) {
				continue
			}
			return errors.Wrapf(err, "cannot delete '%s'", path)
		}
		return nil
	}
	return ErrPathNotSupported
}

func (f *Factory) Rename(src, dest string) error {
	for _, fsys := range f.fss {
		if err := fsys.Rename(src, dest); err != nil {
			if ErrNotSupported(err) {
				continue
			}
			return errors.Wrapf(err, "error renaming '%s' --> '%s'", src, dest)
		}
		break
	}
	srcFP, err := f.Open(src)
	if err != nil {
		return errors.Wrapf(err, "cannot open '%s'", src)
	}
	destFP, err := f.Create(dest)
	if err != nil {
		srcFP.Close()
		return errors.Wrapf(err, "cannot create '%s'", dest)
	}
	defer destFP.Close()
	if _, err := io.Copy(destFP, srcFP); err != nil {
		srcFP.Close()
		return errors.Wrapf(err, "cannot copy '%s' --> '%s'", src, dest)
	}
	if err := srcFP.Close(); err != nil {
		return errors.Wrapf(err, "cannot close '%s'", src)
	}
	if err := f.Delete(src); err != nil {
		return errors.Wrapf(err, "cannot delete '%s'", src)
	}
	return nil
}
