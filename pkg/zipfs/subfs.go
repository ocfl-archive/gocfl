package zipfs

import (
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
	"path/filepath"
)

type SubFS struct {
	*FS
	pathPrefix string
}

func (zSubFS *SubFS) Close() error {
	return nil
}

func (zSubFS *SubFS) Open(name string) (fs.File, error) {
	name = filepath.ToSlash(filepath.Join(zSubFS.pathPrefix, filepath.Clean(name)))
	return zSubFS.FS.Open(name)
}

func (zSubFS *SubFS) Create(name string) (io.WriteCloser, error) {
	name = filepath.ToSlash(filepath.Join(zSubFS.pathPrefix, filepath.Clean(name)))
	return zSubFS.FS.Create(name)
}

func (zSubFS *SubFS) ReadDir(name string) ([]fs.DirEntry, error) {
	name = filepath.ToSlash(filepath.Join(zSubFS.pathPrefix, filepath.Clean(name)))
	return zSubFS.FS.ReadDir(name)
}

func (zSubFS *SubFS) Stat(name string) (fs.FileInfo, error) {
	name = filepath.ToSlash(filepath.Join(zSubFS.pathPrefix, filepath.Clean(name)))
	return zSubFS.FS.Stat(name)
}

func (zSubFS *SubFS) SubFS(name string) ocfl.OCFLFS {
	if name == "." || name == "" {
		return zSubFS
	}
	return &SubFS{
		FS:         zSubFS.FS,
		pathPrefix: filepath.ToSlash(filepath.Join(zSubFS.pathPrefix, filepath.Clean(name))),
	}
}
