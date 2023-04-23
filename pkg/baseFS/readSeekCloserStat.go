//go:build exclude

package baseFS

import (
	"emperror.dev/errors"
	"io"
	"io/fs"
)

type ReadSeekCloserStat interface {
	fs.File
	io.ReadSeekCloser
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
