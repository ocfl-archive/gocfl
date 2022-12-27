package s3fs

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/baseFS"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
	"strings"
)

type BaseFS struct {
	endpoint, accessKeyID, secretAccessKey, region string
	useSSL                                         bool
	logger                                         *logging.Logger
}

func NewBaseFS(endpoint, accessKeyID, secretAccessKey, region string, useSSL bool, logger *logging.Logger) (baseFS.FS, error) {
	bFS := &BaseFS{
		endpoint:        endpoint,
		accessKeyID:     accessKeyID,
		secretAccessKey: secretAccessKey,
		region:          region,
		useSSL:          useSSL,
		logger:          logger,
	}
	return bFS, nil
}

func (b *BaseFS) SetFSFactory(factory *baseFS.Factory) {
}

func (b BaseFS) valid(path string) bool {
	return strings.HasPrefix(path, "bucket:")
}

func (b BaseFS) GetFSRW(path string) (ocfl.OCFLFS, error) {
	if !b.valid(path) {
		return nil, baseFS.ErrPathNotSupported
	}

	if !strings.HasPrefix(path, "bucket:") {
		return nil, errors.Errorf("invalid path '%s' (no bucket scheme)", path)
	}
	parts := strings.Split(path[len("bucket:"):], "/")
	if len(parts) == 0 {
		return nil, errors.Errorf("invalid path '%s' (no bucket name)", path)
	}
	f, err := NewFS(b.endpoint, b.accessKeyID, b.secretAccessKey, parts[0], b.region, b.useSSL, b.logger)
	if err != nil {
		return nil, errors.Errorf("cannot instantiate S3FS for '%s'", path)
	}
	if len(parts) > 1 {
		return f.SubFSRW(strings.Join(parts[1:], "/"))
	}
	return f, nil
}

func (b BaseFS) GetFS(path string) (ocfl.OCFLFSRead, error) {
	if !b.valid(path) {
		return nil, baseFS.ErrPathNotSupported
	}

	if !strings.HasPrefix(path, "bucket:") {
		return nil, errors.Errorf("invalid path '%s' (no bucket scheme)", path)
	}
	parts := strings.Split(path[len("bucket:"):], "/")
	if len(parts) == 0 {
		return nil, errors.Errorf("invalid path '%s' (no bucket name)", path)
	}
	f, err := NewFS(b.endpoint, b.accessKeyID, b.secretAccessKey, parts[0], b.region, b.useSSL, b.logger)
	if err != nil {
		return nil, errors.Errorf("cannot instantiate S3FS for '%s'", path)
	}
	if len(parts) > 1 {
		return f.SubFS(strings.Join(parts[1:], "/"))
	}
	return f, nil
}

func (b *BaseFS) Open(path string) (baseFS.ReadSeekCloserStat, error) {
	if !b.valid(path) {
		return nil, baseFS.ErrPathNotSupported
	}

	if !strings.HasPrefix(path, "bucket:") {
		return nil, errors.Errorf("invalid path '%s' (no bucket scheme)", path)
	}
	parts := strings.Split(path[len("bucket:"):], "/")
	if len(parts) < 2 {
		return nil, errors.Errorf("invalid path '%s' (no bucket name)", path)
	}
	fsys, err := b.GetFS(fmt.Sprintf("bucket:%s", strings.Join(parts[0:len(parts)-1], "/")))
	if err != nil {
		return nil, errors.Wrap(err, "cannot get filesystem")
	}
	fname := strings.Join(parts[1:], "/")
	fp, err := fsys.Open(fname)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open '%s'", fname)
	}
	rsc, ok := fp.(baseFS.ReadSeekCloserStat)
	if !ok {
		return nil, errors.Errorf("no FileSeeker for '%s'", path)
	}
	return baseFS.NewGenericReadSeekCloserStat(rsc, func() error {
		return fsys.Close()
	})
}

func (b *BaseFS) Create(path string) (io.WriteCloser, error) {
	if !b.valid(path) {
		return nil, baseFS.ErrPathNotSupported
	}
	parts := strings.Split(path[len("bucket:"):], "/")
	if len(parts) < 2 {
		return nil, errors.Errorf("invalid path '%s' (no bucket name)", path)
	}
	fsys, err := b.GetFSRW(fmt.Sprintf("bucket:%s", strings.Join(parts[0:len(parts)-1], "/")))
	if err != nil {
		return nil, errors.Wrap(err, "cannot get filesystem")
	}
	fname := strings.Join(parts[1:], "/")
	fp, err := fsys.Create(fname)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open '%s'", fname)
	}
	return baseFS.NewGenericWriteCloser(fp, func() error {
		return fsys.Close()
	})
}

func (b *BaseFS) Rename(src, dest string) error {
	// s3 does not support rename functionality...
	return baseFS.ErrPathNotSupported
}

func (b *BaseFS) Delete(path string) error {
	if !b.valid(path) {
		return baseFS.ErrPathNotSupported
	}
	parts := strings.Split(path[len("bucket:"):], "/")
	if len(parts) < 2 {
		return errors.Errorf("invalid path '%s' (no bucket name)", path)
	}
	fsys, err := b.GetFSRW(fmt.Sprintf("bucket:%s", strings.Join(parts[0:len(parts)-1], "/")))
	if err != nil {
		return errors.Wrap(err, "cannot get filesystem")
	}
	defer fsys.Close()
	fname := strings.Join(parts[1:], "/")
	return errors.Wrapf(fsys.Delete(fname), "cannot delete '%s'", path)
}

var (
	_ baseFS.FS = &BaseFS{}
)
