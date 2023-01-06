package baseFS

import (
	"emperror.dev/errors"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"io"
)

var ErrPathNotSupported error = errors.New("path not supported")

func ErrNotSupported(err error) bool {
	return err == ErrPathNotSupported
}

type FS interface {
	SetFSFactory(factory *Factory)
	GetFSRW(path string) (ocfl.OCFLFS, error)
	GetFS(path string) (ocfl.OCFLFSRead, error)
	Open(path string) (ReadSeekCloserStat, error)
	Create(path string) (io.WriteCloser, error)
	Delete(path string) error
	Rename(src, dest string) error
}
