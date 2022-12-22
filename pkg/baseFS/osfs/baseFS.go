package osfs

import (
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/baseFS"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
	"net/url"
	"os"
	"strings"
)

type BaseFS struct {
	logger *logging.Logger
}

func NewBaseFS(logger *logging.Logger) (baseFS.FS, error) {
	return &BaseFS{logger: logger}, nil
}

func (b *BaseFS) SetFSFactory(factory *baseFS.Factory) {
}

func (*BaseFS) Valid(path string) bool {
	if u, err := url.Parse(path); err == nil {
		return u.Scheme == "file"
	}

	return true
}

func (f *BaseFS) GetFSRW(path string) (ocfl.OCFLFS, error) {
	return NewFS(path, f.logger)
}

func (f *BaseFS) GetFS(path string) (ocfl.OCFLFSRead, error) {
	return NewFS(path, f.logger)
}

func (b *BaseFS) Open(path string) (baseFS.ReadSeekCloserStat, error) {
	if strings.HasPrefix(strings.ToLower(path), "file://") {
		path = path[len("file://"):]
	}
	fp, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open '%s'", path)
	}
	return fp, nil
}

func (b *BaseFS) Create(path string) (io.WriteCloser, error) {
	if strings.HasPrefix(strings.ToLower(path), "file://") {
		path = path[len("file://"):]
	}
	fp, err := os.Create(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create '%s'", path)
	}
	return fp, nil
}

var (
	_ baseFS.FS = &BaseFS{}
)
