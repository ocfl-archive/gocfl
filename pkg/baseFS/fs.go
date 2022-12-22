package baseFS

import (
	"emperror.dev/errors"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
)

type FS interface {
	SetFSFactory(factory *Factory)
	Valid(path string) bool
	GetFSRW(path string) (ocfl.OCFLFS, error)
	GetFS(path string) (ocfl.OCFLFSRead, error)
	Open(path string) (ReadSeekCloserStat, error)
	Create(path string) (io.WriteCloser, error)
}

type GenericWriteCloser struct {
	io.WriteCloser
	close func() error
}

func NewGenericWriteCloser(wc io.WriteCloser, close func() error) (*GenericWriteCloser, error) {
	return &GenericWriteCloser{
		WriteCloser: wc,
		close:       close,
	}, nil
}

func (gc *GenericWriteCloser) Close() error {
	errs := []error{
		gc.WriteCloser.Close(),
		gc.close(),
	}
	return errors.Combine(errs...)
}

type ReadSeekCloserStat interface {
	io.ReadSeekCloser
	Stat() (fs.FileInfo, error)
}

type GenericReadSeekCloserStat struct {
	ReadSeekCloserStat
	close func() error
}

func NewGenericReadSeekCloserStat(rsc ReadSeekCloserStat, close func() error) (*GenericReadSeekCloserStat, error) {
	return &GenericReadSeekCloserStat{
		ReadSeekCloserStat: rsc,
		close:              close,
	}, nil
}

func (gc *GenericReadSeekCloserStat) Close() error {
	errs := []error{
		gc.ReadSeekCloserStat.Close(),
		gc.close(),
	}
	return errors.Combine(errs...)
}
