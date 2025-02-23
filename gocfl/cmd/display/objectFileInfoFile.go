package display

import (
	"io/fs"
	"time"
)

func NewObjectFileInfoFile(name string, modTime time.Time, size int64) *ObjectFileInfoFile {
	return &ObjectFileInfoFile{
		name:    name,
		modTime: modTime,
		size:    size,
	}
}

type ObjectFileInfoFile struct {
	name    string
	modTime time.Time
	size    int64
}

func (o *ObjectFileInfoFile) Type() fs.FileMode {
	return 0644
}

func (o *ObjectFileInfoFile) Info() (fs.FileInfo, error) {
	return o, nil
}

func (o *ObjectFileInfoFile) Size() int64 {
	return o.size
}

func (o *ObjectFileInfoFile) Mode() fs.FileMode {
	return 0644
}

func (o *ObjectFileInfoFile) ModTime() time.Time {
	return o.modTime
}

func (o *ObjectFileInfoFile) IsDir() bool {
	return false
}

func (o *ObjectFileInfoFile) Sys() any {
	return nil
}

func (o *ObjectFileInfoFile) Name() string {
	return o.name
}

var _ fs.FileInfo = (*ObjectFileInfoFile)(nil)
var _ fs.DirEntry = (*ObjectFileInfoFile)(nil)
