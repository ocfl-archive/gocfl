package baseFS

import (
	"emperror.dev/errors"
	"io"
)

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
