package zipfs

import (
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/baseFS"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"io"
	"strings"
)

type BaseFS struct {
	factory *baseFS.Factory
	logger  *logging.Logger
}

func NewBaseFS(logger *logging.Logger) (baseFS.FS, error) {
	return &BaseFS{logger: logger}, nil
}

func (b *BaseFS) SetFSFactory(factory *baseFS.Factory) {
	b.factory = factory
}

func (b *BaseFS) Valid(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".zip")
}

type readSeekerToReaderAt struct {
	readSeeker io.ReadSeeker
}

func (stra *readSeekerToReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if _, err := stra.readSeeker.Seek(off, io.SeekStart); err != nil {
		return 0, errors.Wrapf(err, "cannot seek to offset %v", off)
	}
	return stra.readSeeker.Read(p)
}

func (b *BaseFS) GetFSRW(path string) (ocfl.OCFLFS, error) {
	fp, err := b.factory.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open '%s'", path)
	}
	zipReader, ok := fp.(baseFS.ReadSeekCloserStat)
	if !ok {
		return nil, errors.Errorf("no FileSeeker for '%s'", path)
	}
	fi, err := zipReader.Stat()
	if err != nil {
		zipReader.Close()
		return nil, errors.Wrapf(err, "cannot stat '%s'", path)
	}
	zipReaderAt, ok := zipReader.(io.ReaderAt)
	if !ok {
		zipReaderAt = &readSeekerToReaderAt{readSeeker: zipReader}
	}
	var zipWriter io.WriteCloser
	pathTemp := path + ".tmp"
	zipWriter, err = b.factory.Create(pathTemp)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create '%s'", pathTemp)
	}
	ocfs, err := NewFS(zipReaderAt, fi.Size(), zipWriter, func() error {
		errs := []error{}
		if zipReader != nil {
			errs = append(errs, zipReader.Close())
		}
		if zipWriter != nil {
			errs = append(errs, zipWriter.Close())
		}
		return errors.Combine(errs...)
	}, b.logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create zipfs")
	}
	return ocfs, nil
}

func (b *BaseFS) GetFS(path string) (ocfl.OCFLFSRead, error) {
	fp, err := b.factory.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open '%s'", path)
	}
	zipReader, ok := fp.(baseFS.ReadSeekCloserStat)
	if !ok {
		return nil, errors.Errorf("no FileSeeker for '%s'", path)
	}
	fi, err := zipReader.Stat()
	if err != nil {
		zipReader.Close()
		return nil, errors.Wrapf(err, "cannot stat '%s'", path)
	}
	zipReaderAt, ok := zipReader.(io.ReaderAt)
	if !ok {
		zipReaderAt = &readSeekerToReaderAt{readSeeker: zipReader}
	}
	ocfs, err := NewFS(zipReaderAt, fi.Size(), nil, func() error {
		errs := []error{}
		if zipReader != nil {
			errs = append(errs, zipReader.Close())
		}
		return errors.Combine(errs...)
	}, b.logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create zipfs")
	}
	return ocfs, nil
}

func (b *BaseFS) Open(path string) (baseFS.ReadSeekCloserStat, error) {
	return nil, errors.New("cannot open file inside ZIP")
}

func (b *BaseFS) Create(path string) (io.WriteCloser, error) {
	return nil, errors.New("cannot create file inside ZIP")
}

var (
	_ baseFS.FS = &BaseFS{}
)
