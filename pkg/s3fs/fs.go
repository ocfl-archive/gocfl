package s3fs

import (
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"golang.org/x/exp/slices"
	"io"
	"io/fs"
	"net/http"
)

type FS struct {
	client   *minio.Client
	bucket   string
	ctx      context.Context
	region   string
	endpoint string
}

func NewFS(endpoint, accessKeyID, secretAccessKey, bucket, region string, useSSL bool) (*FS, error) {
	var err error
	fs := &FS{
		ctx:      context.Background(),
		client:   nil,
		bucket:   bucket,
		region:   region,
		endpoint: endpoint,
	}
	fs.client, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, errors.Wrap(err, "cannot create s3 client instance")
	}
	return fs, nil
}

func (s3fs *FS) Open(name string) (fs.File, error) {
	object, err := s3fs.client.GetObject(s3fs.ctx, s3fs.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open '%s/%s/%s'", s3fs.client.EndpointURL(), s3fs.bucket, name)
	}
	return &File{
		object,
	}, nil
}

func (s3fs *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	result := []fs.DirEntry{}
	for objectInfo := range s3fs.client.ListObjects(s3fs.ctx, s3fs.bucket, minio.ListObjectsOptions{}) {
		result = append(result, DirEntry{
			&FileInfo{&objectInfo},
		})
	}
	return result, nil
}

func (s3fs *FS) Create(name string) (io.WriteCloser, error) {
	wc := NewWriteCloser()
	go func() {
		uploadInfo, err := s3fs.client.PutObject(s3fs.ctx, s3fs.bucket, name, wc.pr, -1, minio.PutObjectOptions{})
		wc.c <- UploadInfo{
			uploadInfo: uploadInfo,
			err:        err,
		}
	}()
	return wc, nil
}

func (s3fs *FS) Delete(name string) error {
	if err := s3fs.client.RemoveObject(s3fs.ctx, s3fs.bucket, name, minio.RemoveObjectOptions{}); err != nil {
		return errors.Wrapf(err, "cannot remove '%s'", name)
	}
	return nil
}

func (s3fs *FS) SubFS(subfolder string) (ocfl.OCFLFS, error) {
	//TODO implement me
	panic("implement me")
}

func (s3fs *FS) Close() error {
	return nil
}

func (s3fs *FS) Discard() error {
	// cannot reset s3 fs
	return nil
}

func (s3fs *FS) String() string {
	return fmt.Sprintf("%s/%s", s3fs.endpoint, s3fs.bucket)
}

var notFoundStatus = []int{
	http.StatusNotFound,
	// http.StatusForbidden,
	// http.StatusConflict,
	// http.StatusPreconditionFailed,
}

func (s3fs *FS) IsNotExist(err error) bool {
	errResp, ok := err.(minio.ErrorResponse)
	if !ok {
		return false
	}
	return slices.Contains(notFoundStatus, errResp.StatusCode)
}

func (s3fs *FS) WalkDir(root string, fn fs.WalkDirFunc) error {
	for objectInfo := range s3fs.client.ListObjects(s3fs.ctx, s3fs.bucket, minio.ListObjectsOptions{
		Prefix:    root,
		Recursive: true,
	}) {
		if err := fn(objectInfo.Key, DirEntry{&FileInfo{&objectInfo}}, nil); err != nil {
			return errors.Wrapf(err, "error in '%s'", objectInfo.Key)
		}
	}
	return nil
}

func (s3fs *FS) Stat(name string) (fs.FileInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (s3fs *FS) HasContent() bool {
	//TODO implement me
	panic("implement me")
}

var (
	_ ocfl.OCFLFS = &FS{}
)
