package display

import (
	"io/fs"
	"time"
)

func NewObjectFileDir(name string, modTime time.Time, fsys *ObjectFS) *ObjectFileDir {
	return &ObjectFileDir{
		name:    name,
		modTime: modTime,
		fsys:    fsys,
	}
}

type ObjectFileDir struct {
	name    string
	modTime time.Time
	fsys    *ObjectFS
}

func (o *ObjectFileDir) ReadDir(n int) ([]fs.DirEntry, error) {
	files, err := o.fsys.readDir(o.name, n)
	if err != nil {
		return nil, err
	}
	entries := make([]fs.DirEntry, 0, len(files))
	for _, file := range files {
		entries = append(entries, fs.DirEntry(file))
	}
	return entries, nil
}

func (o *ObjectFileDir) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (o *ObjectFileDir) Read(bytes []byte) (int, error) {
	return 0, fs.ErrInvalid
}

func (o *ObjectFileDir) Close() error {
	return nil
}

func (o *ObjectFileDir) Stat() (fs.FileInfo, error) {
	return NewObjectFileInfoDir(o.name, o.modTime), nil
}

var _ fs.File = (*ObjectFileDir)(nil)
var _ fs.ReadDirFile = (*ObjectFileDir)(nil)
