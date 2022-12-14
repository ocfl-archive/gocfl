package s3fs

import (
	"github.com/minio/minio-go/v7"
	"io/fs"
	"time"
)

type FileInfo struct {
	*minio.ObjectInfo
}

func (s3fi FileInfo) Name() string {
	return s3fi.Key
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
	return false
}

func (s3fi FileInfo) Sys() any {
	return nil
}

var _ fs.FileInfo = &FileInfo{}
