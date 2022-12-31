package s3fs

import (
	"emperror.dev/errors"
	"github.com/minio/minio-go/v7"
	"io/fs"
)

type File struct {
	*minio.Object
}

func (s3f *File) Read(bytes []byte) (int, error) {
	return s3f.Object.Read(bytes)
}

func (s3f *File) Close() error {
	return s3f.Object.Close()
}

func (s3f *File) Stat() (fs.FileInfo, error) {
	oInfo, err := s3f.Object.Stat()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot stat '%v'", s3f.Object)
	}
	return FileInfo{
		&oInfo,
	}, nil
}

var _ fs.File = &File{}
