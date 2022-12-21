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
	"strings"
)

type FS struct {
	client   *minio.Client
	bucket   string
	region   string
	endpoint string
}

func NewFS(endpoint, accessKeyID, secretAccessKey, bucket, region string, useSSL bool) (*FS, error) {
	var err error
	fs := &FS{
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
	ctx := context.Background()
	name = strings.TrimLeft(name, "./")
	object, err := s3fs.client.GetObject(ctx, s3fs.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open '%s/%s/%s'", s3fs.client.EndpointURL(), s3fs.bucket, name)
	}
	objectInfo, err := object.Stat()
	if err != nil {
		object.Close()
		return nil, errors.Wrapf(err, "cannot stat open object '%s'", name)
	}
	if objectInfo.Err != nil {
		object.Close()
		return nil, errors.Wrapf(objectInfo.Err, "error in objectInfo of '%s'", name)
	}
	return &File{
		object,
	}, nil
}

func (s3fs *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	ctx := context.Background()
	result := []fs.DirEntry{}
	if len(name) > 0 {
		name = strings.TrimRight(name, "/") + "/"
	}
	if name == "./" {
		name = ""
	}
	for objectInfo := range s3fs.client.ListObjects(ctx, s3fs.bucket, minio.ListObjectsOptions{Prefix: name}) {
		if objectInfo.Err != nil {
			return nil, errors.Wrapf(objectInfo.Err, "cannot read '%s'", name)
		}
		oiHelper := objectInfo
		result = append(result, DirEntry{
			&FileInfo{&oiHelper},
		})
	}
	return result, nil
}

func (s3fs *FS) Create(name string) (io.WriteCloser, error) {
	ctx := context.Background()
	wc := NewWriteCloser()
	go func() {
		uploadInfo, err := s3fs.client.PutObject(ctx, s3fs.bucket, name, wc.pr, -1, minio.PutObjectOptions{})
		wc.c <- UploadInfo{
			uploadInfo: uploadInfo,
			err:        err,
		}
	}()
	return wc, nil
}

func (s3fs *FS) Delete(name string) error {
	ctx := context.Background()
	if err := s3fs.client.RemoveObject(ctx, s3fs.bucket, name, minio.RemoveObjectOptions{}); err != nil {
		return errors.Wrapf(err, "cannot remove '%s'", name)
	}
	return nil
}

func (s3fs *FS) SubFS(subfolder string) (ocfl.OCFLFS, error) {
	return NewSubFS(s3fs, subfolder)
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
	root = strings.TrimRight(root, "/") + "/"
	ctx := context.Background()
	for objectInfo := range s3fs.client.ListObjects(ctx, s3fs.bucket, minio.ListObjectsOptions{
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
	ctx := context.Background()
	objectInfo, err := s3fs.client.StatObject(ctx, s3fs.bucket, name, minio.StatObjectOptions{})
	if err != nil {
		if s3fs.IsNotExist(err) {
			if s3fs.hasContent(name) {
				return &DummyDir{name: name}, nil
			}
		}
		return nil, errors.Wrapf(err, "cannot stat '%s'", name)
	}
	return &FileInfo{&objectInfo}, nil
}

func (s3fs *FS) hasContent(prefix string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	chanObjectInfo := s3fs.client.ListObjects(ctx, s3fs.bucket, minio.ListObjectsOptions{Prefix: prefix})
	objectInfo, ok := <-chanObjectInfo
	if ok {
		if objectInfo.Err != nil {
			return true
		}
	}
	cancel()
	return ok
}

func (s3fs *FS) HasContent() bool {
	return s3fs.hasContent("")
}

var (
	_ ocfl.OCFLFS = &FS{}
)
