package zipfs

import (
	"emperror.dev/errors"
	"io"
	"io/fs"
)

type File struct {
	*FileInfo
	r io.ReadCloser
}

func NewFile(info *FileInfo) (*File, error) {
	var err error

	f := &File{
		FileInfo: info,
	}
	if f.r, err = info.zf.Open(); err != nil {
		return nil, errors.Wrapf(err, "cannot open zip item %s", info.Name())
	}
	return f, nil
}

func (f *File) Stat() (fs.FileInfo, error) { return f.FileInfo, nil }

func (f *File) Read(p []byte) (int, error) {
	num, err := f.r.Read(p)
	if err != nil && err != io.EOF {
		return num, errors.Wrapf(err, "cannot read from %s", f.Name())
	}
	return num, err
}

func (f *File) Close() error {
	if err := f.r.Close(); err != nil {
		return errors.Wrapf(err, "cannot close %s", f.Name())
	}
	return nil
}
