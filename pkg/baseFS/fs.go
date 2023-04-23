//go:build exclude

package baseFS

import (
	"emperror.dev/errors"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
)

var ErrPathNotSupported error = errors.New("path not supported")

func ErrNotSupported(err error) bool {
	return err == ErrPathNotSupported
}

type FileWriter interface {
	fs.File
	Write(path string, data []byte) error
}

type RWFS interface {
	fs.FS
	Create(path string) (io.WriteCloser, error)
}

type FS interface {
	// fs.FS
	Open(path string) (fs.File, error)

	SetFSFactory(factory *Factory)
	GetFSRW(path string, clear bool) (ocfl.OCFLFS, error)
	GetFS(path string) (ocfl.OCFLFSRead, error)
	Create(path string) (io.WriteCloser, error)
	Delete(path string) error
	Rename(src, dest string) error
}
