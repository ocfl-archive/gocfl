//go:build exclude

package s3fs

import (
	"github.com/minio/minio-go/v7"
	"io/fs"
	"path/filepath"
	"time"
)

type FileInfo struct {
	*minio.ObjectInfo
}

func (s3fi FileInfo) String() string {
	return s3fi.Key
}

func (s3fi FileInfo) Name() string {
	return filepath.Base(s3fi.Key)
}

func (s3fi FileInfo) Size() int64 {
	return s3fi.ObjectInfo.Size
}

func (s3fi FileInfo) Mode() fs.FileMode {
	return 0
}

func (s3fi FileInfo) ModTime() time.Time {
	return s3fi.LastModified
}

func (s3fi FileInfo) IsDir() bool {
	return s3fi.ObjectInfo.Size == 0 &&
		s3fi.ObjectInfo.StorageClass == ""
	//	return false
}

func (s3fi FileInfo) Sys() any {
	return nil
}

var _ fs.FileInfo = &FileInfo{}

type DummyDir struct {
	name string
}

func (s3fi DummyDir) String() string {
	return s3fi.name
}

func (s3fi DummyDir) Name() string {
	return filepath.Base(s3fi.name)
}

func (s3fi DummyDir) Size() int64 {
	return 0
}

func (s3fi DummyDir) Mode() fs.FileMode {
	return 0
}

func (s3fi DummyDir) ModTime() time.Time {
	return time.Time{}
}

func (s3fi DummyDir) IsDir() bool {
	return true
}

func (s3fi DummyDir) Sys() any {
	return nil
}

var _ fs.FileInfo = &DummyDir{}
