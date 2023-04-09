package osfs

import (
	"emperror.dev/errors"
	"github.com/je4/gocfl/v2/pkg/baseFS"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/op/go-logging"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
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

var winPathRegexp = regexp.MustCompile("^[A-Za-z]:/[^:]*$")

func (*BaseFS) valid(path string) bool {
	path = filepath.ToSlash(path)
	if winPathRegexp.MatchString(path) {
		return true
	}
	if u, err := url.Parse(path); err == nil {
		return u.Scheme == "file" || u.Scheme == ""
	}

	return true
}

func (b *BaseFS) Rename(src, dest string) error {
	if !b.valid(src) {
		return baseFS.ErrPathNotSupported
	}
	if !b.valid(dest) {
		return baseFS.ErrPathNotSupported
	}
	fs, err := b.GetFSRW("/", false)
	if err != nil {
		return errors.Wrap(err, "cannot get fs for '/'")
	}
	defer fs.Close()
	return fs.Rename(src, dest)
}

func (b *BaseFS) GetFSRW(path string, clear bool) (ocfl.OCFLFS, error) {
	if !b.valid(path) {
		return nil, baseFS.ErrPathNotSupported
	}
	return NewFS(path, clear, b.logger)
}

func (b *BaseFS) GetFS(path string) (ocfl.OCFLFSRead, error) {
	if !b.valid(path) {
		return nil, baseFS.ErrPathNotSupported
	}
	return NewFS(path, false, b.logger)
}

func (b *BaseFS) Open(path string) (fs.File, error) {
	if !b.valid(path) {
		return nil, baseFS.ErrPathNotSupported
	}
	if strings.HasPrefix(strings.ToLower(path), "file://") {
		path = path[len("file://"):]
	}
	fp, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fs.ErrNotExist
		}
		return nil, errors.Wrapf(err, "cannot open '%s'", path)
	}
	return fp, nil
}

func (b *BaseFS) Create(path string) (io.WriteCloser, error) {
	if !b.valid(path) {
		return nil, baseFS.ErrPathNotSupported
	}
	if strings.HasPrefix(strings.ToLower(path), "file://") {
		path = path[len("file://"):]
	}
	fp, err := os.Create(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create '%s'", path)
	}
	return fp, nil
}

func (b *BaseFS) Delete(path string) error {
	if !b.valid(path) {
		return baseFS.ErrPathNotSupported
	}
	if strings.HasPrefix(strings.ToLower(path), "file://") {
		path = path[len("file://"):]
	}
	return errors.Wrapf(os.Remove(path), "cannot delete '%s'", path)
}

// check interface satisfaction
var (
	_ fs.FS         = &FS{}
	_ fs.StatFS     = &FS{}
	_ fs.ReadFileFS = &FS{}
	_ fs.ReadDirFS  = &FS{}
	_ fs.SubFS      = &FS{}
)
