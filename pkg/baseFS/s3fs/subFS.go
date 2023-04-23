//go:build exclude

package s3fs

import (
	"fmt"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

type SubFS struct {
	*FS
	pathPrefix string
}

func NewSubFS(fs *FS, pathPrefix string) (*SubFS, error) {
	sfs := &SubFS{
		FS:         fs,
		pathPrefix: strings.TrimRight(filepath.ToSlash(filepath.Clean(pathPrefix)), "/") + "/",
	}
	return sfs, nil
}

func (s3SubFS *SubFS) String() string {
	return fmt.Sprintf("%s/%s", s3SubFS.FS.String(), s3SubFS.pathPrefix)
}

func (s3SubFS *SubFS) OpenSeeker(name string) (ocfl.FileSeeker, error) {
	name = filepath.ToSlash(filepath.Join(s3SubFS.pathPrefix, filepath.Clean(name)))
	return s3SubFS.FS.OpenSeeker(name)
}

func (s3SubFS *SubFS) Open(name string) (fs.File, error) {
	name = filepath.ToSlash(filepath.Join(s3SubFS.pathPrefix, filepath.Clean(name)))
	return s3SubFS.FS.Open(name)
}

func (s3SubFS *SubFS) Create(name string) (io.WriteCloser, error) {
	name = filepath.ToSlash(filepath.Join(s3SubFS.pathPrefix, filepath.Clean(name)))
	return s3SubFS.FS.Create(name)
}

func (s3SubFS *SubFS) Delete(name string) error {
	name = filepath.ToSlash(filepath.Join(s3SubFS.pathPrefix, filepath.Clean(name)))
	return s3SubFS.FS.Delete(name)
}

func (s3SubFS *SubFS) ReadDir(path string) ([]fs.DirEntry, error) {
	path = filepath.ToSlash(filepath.Join(s3SubFS.pathPrefix, filepath.Clean(path)))
	return s3SubFS.FS.ReadDir(path)
}

func (s3SubFS *SubFS) ReadFile(name string) ([]byte, error) {
	name = filepath.ToSlash(filepath.Join(s3SubFS.pathPrefix, filepath.Clean(name)))
	return s3SubFS.FS.ReadFile(name)
}

func (s3SubFS *SubFS) WalkDir(path string, fn fs.WalkDirFunc) error {
	path = filepath.ToSlash(filepath.Join(s3SubFS.pathPrefix, filepath.Clean(path)))
	prefix := strings.TrimRight(s3SubFS.pathPrefix, "/") + "/"
	return s3SubFS.FS.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		return fn(strings.TrimPrefix(path, prefix), d, err)
	})
}

func (s3SubFS *SubFS) Stat(path string) (fs.FileInfo, error) {
	path = filepath.ToSlash(filepath.Join(s3SubFS.pathPrefix, filepath.Clean(path)))
	return s3SubFS.FS.Stat(path)
}

func (s3SubFS *SubFS) SubFSRW(path string) (ocfl.OCFLFS, error) {
	name := filepath.ToSlash(filepath.Join(s3SubFS.pathPrefix, filepath.Clean(path)))
	if name == "." {
		name = ""
	}
	if name == "" {
		return s3SubFS, nil
	}
	return s3SubFS.FS.SubFSRW(name)
}

func (s3SubFS *SubFS) SubFS(path string) (ocfl.OCFLFSRead, error) {
	name := filepath.ToSlash(filepath.Join(s3SubFS.pathPrefix, filepath.Clean(path)))
	if name == "." {
		name = ""
	}
	if name == "" {
		return s3SubFS, nil
	}
	return s3SubFS.FS.SubFS(name)
}
func (s3SubFS *SubFS) HasContent() bool {
	return s3SubFS.FS.hasContent(s3SubFS.pathPrefix)
}

// check interface satisfaction
var (
	_ ocfl.OCFLFS = &SubFS{}
)
