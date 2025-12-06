package display

import (
	"io/fs"
	"time"
)

func NewObjectFileInfoDir(name string, modTime time.Time) *ObjectFileInfoDir {
	return &ObjectFileInfoDir{
		name:    name,
		modTime: modTime,
	}
}

type ObjectFileInfoDir struct {
	name    string
	modTime time.Time
}

func (o *ObjectFileInfoDir) Type() fs.FileMode {
	return fs.ModeDir
}

func (o *ObjectFileInfoDir) Info() (fs.FileInfo, error) {
	return o, nil
}

func (o *ObjectFileInfoDir) Size() int64 {
	return 0
}

func (o *ObjectFileInfoDir) Mode() fs.FileMode {
	return fs.ModeDir
}

func (o *ObjectFileInfoDir) ModTime() time.Time {
	return o.modTime
}

func (o *ObjectFileInfoDir) IsDir() bool {
	return true
}

func (o *ObjectFileInfoDir) Sys() any {
	return nil
}

func (o *ObjectFileInfoDir) Name() string {
	return o.name
}

var _ fs.FileInfo = (*ObjectFileInfoDir)(nil)
