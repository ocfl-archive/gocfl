//go:build exclude

package zipfs

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
		pathPrefix: filepath.ToSlash(filepath.Clean(pathPrefix)),
	}
	return sfs, nil
}

func (zipSubFS *SubFS) String() string {
	return fmt.Sprintf("zipfs://%s", zipSubFS.pathPrefix)
}

func (zipSubFS *SubFS) OpenSeeker(name string) (ocfl.FileSeeker, error) {
	name = filepath.ToSlash(filepath.Join(zipSubFS.pathPrefix, filepath.Clean(name)))
	return zipSubFS.FS.OpenSeeker(name)
}

func (zipSubFS *SubFS) Open(name string) (fs.File, error) {
	name = filepath.ToSlash(filepath.Join(zipSubFS.pathPrefix, filepath.Clean(name)))
	return zipSubFS.FS.Open(name)
}

func (zipSubFS *SubFS) ReadFile(name string) ([]byte, error) {
	name = filepath.ToSlash(filepath.Join(zipSubFS.pathPrefix, filepath.Clean(name)))
	return zipSubFS.FS.ReadFile(name)
}

func (zipSubFS *SubFS) Create(name string) (io.WriteCloser, error) {
	name = filepath.ToSlash(filepath.Join(zipSubFS.pathPrefix, filepath.Clean(name)))
	return zipSubFS.FS.Create(name)
}

func (zipSubFS *SubFS) Delete(name string) error {
	name = filepath.ToSlash(filepath.Join(zipSubFS.pathPrefix, filepath.Clean(name)))
	return zipSubFS.FS.Delete(name)
}

func (zipSubFS *SubFS) ReadDir(path string) ([]fs.DirEntry, error) {
	path = filepath.ToSlash(filepath.Join(zipSubFS.pathPrefix, filepath.Clean(path)))
	return zipSubFS.FS.ReadDir(path)
}

func (zipSubFS *SubFS) WalkDir(path string, fn fs.WalkDirFunc) error {
	path = filepath.ToSlash(filepath.Join(zipSubFS.pathPrefix, filepath.Clean(path)))
	prefix := zipSubFS.pathPrefix + "/"
	return zipSubFS.FS.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		newPath := strings.TrimPrefix(path, prefix)
		return fn(newPath, d, err)
	})
}

func (zipSubFS *SubFS) Stat(path string) (fs.FileInfo, error) {
	path = filepath.ToSlash(filepath.Join(zipSubFS.pathPrefix, filepath.Clean(path)))
	return zipSubFS.FS.Stat(path)
}

func (zipSubFS *SubFS) SubFSRW(path string) (ocfl.OCFLFS, error) {
	name := filepath.ToSlash(filepath.Join(zipSubFS.pathPrefix, filepath.Clean(path)))
	if name == "." {
		name = ""
	}
	if name == "" {
		return zipSubFS, nil
	}
	return zipSubFS.FS.SubFSRW(name)
}

func (zipSubFS *SubFS) SubFS(path string) (ocfl.OCFLFSRead, error) {
	name := filepath.ToSlash(filepath.Join(zipSubFS.pathPrefix, filepath.Clean(path)))
	if name == "." {
		name = ""
	}
	if name == "" {
		return zipSubFS, nil
	}
	return zipSubFS.FS.SubFS(name)
}

// check interface satisfaction
var (
	_ ocfl.OCFLFS = &SubFS{}
)
