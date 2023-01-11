package s3fs

import (
	"bytes"
	"context"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/op/go-logging"
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
	logger   *logging.Logger
}

func NewFS(endpoint, accessKeyID, secretAccessKey, bucket, region string, useSSL bool, logger *logging.Logger) (*FS, error) {
	var err error
	fs := &FS{
		client:   nil,
		bucket:   bucket,
		region:   region,
		endpoint: endpoint,
		logger:   logger,
	}

	fs.client, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
		Region: region,
		/* Transport: NewDebuggingRoundTripper(
			&http.Transport{},
			logger,
			JustURL,
			URLTiming,
			// CurlCommand,
			RequestHeaders,
			ResponseStatus,
			// ResponseHeaders,
		) ,*/
	})
	if err != nil {
		return nil, errors.Wrap(err, "cannot create s3 client instance")
	}
	return fs, nil
}

func (s3FS *FS) OpenSeeker(name string) (ocfl.FileSeeker, error) {
	s3FS.logger.Debugf("%s - OpenSeeker(%s)", s3FS.String(), name)
	ctx := context.Background()
	name = strings.TrimLeft(name, "./")
	object, err := s3FS.client.GetObject(ctx, s3FS.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open '%s/%s/%s'", s3FS.client.EndpointURL(), s3FS.bucket, name)
	}
	objectInfo, err := object.Stat()
	if err != nil {
		object.Close()
		return nil, fs.ErrNotExist
	}
	if objectInfo.Err != nil {
		object.Close()
		return nil, errors.Wrapf(objectInfo.Err, "error in objectInfo of '%s'", name)
	}
	return &File{
		object,
	}, nil
}

func (s3FS *FS) Open(name string) (fs.File, error) {
	s3FS.logger.Debugf("%s - Open(%s)", s3FS.String(), name)
	return s3FS.OpenSeeker(name)
}

func (s3FS *FS) ReadFile(name string) ([]byte, error) {
	s3FS.logger.Debugf("%s - ReadFile(%s)", s3FS.String(), name)
	fp, err := s3FS.Open(name)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open '%s'", name)
	}
	defer fp.Close()
	data := bytes.NewBuffer(nil)
	if _, err := io.Copy(data, fp); err != nil {
		return nil, errors.Wrapf(err, "cannot read '%s'", name)
	}
	return data.Bytes(), nil
}

func (s3FS *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	s3FS.logger.Debugf("%s - ReadDir(%s)", s3FS.String(), name)
	ctx := context.Background()
	result := []fs.DirEntry{}
	if len(name) > 0 {
		name = strings.TrimRight(name, "/") + "/"
	}
	if name == "./" {
		name = ""
	}
	for objectInfo := range s3FS.client.ListObjects(ctx, s3FS.bucket, minio.ListObjectsOptions{Prefix: name}) {
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

func (s3FS *FS) Create(name string) (io.WriteCloser, error) {
	s3FS.logger.Debugf("%s - Create(%s)", s3FS.String(), name)
	ctx := context.Background()
	wc := NewWriteCloser()
	go func() {
		uploadInfo, err := s3FS.client.PutObject(ctx, s3FS.bucket, name, wc.pr, -1, minio.PutObjectOptions{})
		wc.c <- UploadInfo{
			uploadInfo: uploadInfo,
			err:        err,
		}
	}()
	return wc, nil
}

func (s3FS *FS) Delete(name string) error {
	s3FS.logger.Debugf("%s - Delete(%s)", s3FS.String(), name)
	ctx := context.Background()
	if err := s3FS.client.RemoveObject(ctx, s3FS.bucket, name, minio.RemoveObjectOptions{}); err != nil {
		return errors.Wrapf(err, "cannot remove '%s'", name)
	}
	return nil
}

func (s3FS *FS) SubFSRW(subfolder string) (ocfl.OCFLFS, error) {
	return NewSubFS(s3FS, subfolder)
}

func (s3FS *FS) SubFS(subfolder string) (ocfl.OCFLFSRead, error) {
	return s3FS.SubFSRW(subfolder)
}

func (s3FS *FS) Close() error {
	return nil
}

func (s3FS *FS) Discard() error {
	// cannot reset s3 fs
	return nil
}

func (s3FS *FS) String() string {
	return fmt.Sprintf("%s/%s", s3FS.endpoint, s3FS.bucket)
}

func (s3FS *FS) Rename(src, dest string) error {
	s3FS.logger.Debugf("%s - Rename(%s, %s)", s3FS.String(), src, dest)
	_, err := s3FS.Stat(dest)
	if err != nil {
		if !s3FS.IsNotExist(err) {
			return errors.Wrapf(err, "cannot stat '%s'", dest)
		}
	} else {
		if err := s3FS.Delete(dest); err != nil {
			return errors.Wrapf(err, "cannot delete '%s'", dest)
		}
	}
	// now, dest should not exist...

	srcFP, err := s3FS.Open(src)
	if err != nil {
		return errors.Wrapf(err, "cannot open '%s'", src)
	}
	defer srcFP.Close()
	destFP, err := s3FS.Create(dest)
	if err != nil {
		return errors.Wrapf(err, "cannot create '%s'", dest)
	}
	defer destFP.Close()
	if _, err := io.Copy(destFP, srcFP); err != nil {
		return errors.Wrapf(err, "cannot copy '%s' --> '%s'", src, dest)
	}
	return nil
}

var notFoundStatus = []int{
	http.StatusNotFound,
	// http.StatusForbidden,
	// http.StatusConflict,
	// http.StatusPreconditionFailed,
}

func (s3FS *FS) IsNotExist(err error) bool {
	errResp, ok := err.(minio.ErrorResponse)
	if !ok {
		return false
	}
	return slices.Contains(notFoundStatus, errResp.StatusCode)
}

func (s3FS *FS) WalkDir(root string, fn fs.WalkDirFunc) error {
	s3FS.logger.Debugf("%s - WalkDir(%s)", s3FS.String(), root)
	root = strings.TrimRight(root, "/") + "/"
	ctx := context.Background()
	for objectInfo := range s3FS.client.ListObjects(ctx, s3FS.bucket, minio.ListObjectsOptions{
		Prefix:    root,
		Recursive: true,
	}) {
		if err := fn(objectInfo.Key, DirEntry{&FileInfo{&objectInfo}}, nil); err != nil {
			return errors.Wrapf(err, "error in '%s'", objectInfo.Key)
		}
	}
	return nil
}

func (s3FS *FS) Stat(name string) (fs.FileInfo, error) {
	ctx := context.Background()
	objectInfo, err := s3FS.client.StatObject(ctx, s3FS.bucket, name, minio.StatObjectOptions{})
	if err != nil {
		if s3FS.IsNotExist(err) {
			if s3FS.hasContent(name) {
				return &DummyDir{name: name}, nil
			}
		}
		return nil, errors.Wrapf(err, "cannot stat '%s'", name)
	}
	return &FileInfo{&objectInfo}, nil
}

func (s3FS *FS) hasContent(prefix string) bool {
	s3FS.logger.Debugf("%s - hasContent(%s)", s3FS.String(), prefix)
	ctx, cancel := context.WithCancel(context.Background())
	chanObjectInfo := s3FS.client.ListObjects(ctx, s3FS.bucket, minio.ListObjectsOptions{Prefix: prefix})
	objectInfo, ok := <-chanObjectInfo
	if ok {
		if objectInfo.Err != nil {
			cancel()
			return true
		}
	}
	cancel()
	return ok
}

func (s3FS *FS) HasContent() bool {
	return s3FS.hasContent("")
}

var (
	_ ocfl.OCFLFS = &FS{}
)
