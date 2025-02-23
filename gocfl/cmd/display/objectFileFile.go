package display

import (
	"emperror.dev/errors"
	"io/fs"
	"time"
)

func NewObjectFileFile(name string, modTime time.Time, fsys *ObjectFS) *ObjectFileFile {
	return &ObjectFileFile{
		name:    name,
		modTime: modTime,
		fsys:    fsys,
	}
}

type ObjectFileFile struct {
	name    string
	modTime time.Time
	fsys    *ObjectFS
}

func (o *ObjectFileFile) ReadDir(n int) ([]fs.DirEntry, error) {
	return nil, errors.Wrapf(fs.ErrInvalid, "%s is a file", o.name)
}

func (o *ObjectFileFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (o *ObjectFileFile) Read(bytes []byte) (int, error) {
	return 0, fs.ErrInvalid
}

func (o *ObjectFileFile) Close() error {
	return nil
}

func (o *ObjectFileFile) Stat() (fs.FileInfo, error) {
	return NewObjectFileInfoFile(o.name, o.modTime, 0), nil
}

var _ fs.File = (*ObjectFileFile)(nil)
var _ fs.ReadDirFile = (*ObjectFileFile)(nil)
