package zipfs

import (
	"emperror.dev/errors"
	"github.com/je4/gocfl/v2/pkg/baseFS"
	"github.com/je4/gocfl/v2/pkg/checksum"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/op/go-logging"
	"io"
	"strings"
)

type BaseFS struct {
	factory          *baseFS.Factory
	logger           *logging.Logger
	aes              bool
	aesKey, aesIV    []byte
	digestAlgorithms []checksum.DigestAlgorithm
	noCompression    bool
}

func NewBaseFS(digestAlgorithms []checksum.DigestAlgorithm, noCompression bool, aes bool, aesKey []byte, aesIV []byte, logger *logging.Logger) (baseFS.FS, error) {
	return &BaseFS{
		digestAlgorithms: digestAlgorithms,
		noCompression:    noCompression,
		aes:              aes,
		aesKey:           aesKey,
		aesIV:            aesIV,
		logger:           logger,
	}, nil
}

func (b *BaseFS) SetFSFactory(factory *baseFS.Factory) {
	b.factory = factory
}

func (b *BaseFS) valid(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".zip")
}

type readSeekCloserToCloserAt struct {
	readSeeker io.ReadSeekCloser
}

func (stra *readSeekCloserToCloserAt) ReadAt(p []byte, off int64) (n int, err error) {
	if _, err := stra.readSeeker.Seek(off, io.SeekStart); err != nil {
		return 0, errors.Wrapf(err, "cannot seek to offset %v", off)
	}
	return stra.readSeeker.Read(p)
}

func (stra *readSeekCloserToCloserAt) Close() error {
	return errors.Wrap(stra.readSeeker.Close(), "cannot close")
}

func (b *BaseFS) Rename(src, dest string) error {
	return baseFS.ErrPathNotSupported
}

func (b *BaseFS) Delete(path string) error {
	return baseFS.ErrPathNotSupported
}

func (b *BaseFS) GetFSRW(path string) (ocfl.OCFLFS, error) {
	if !b.valid(path) {
		return nil, baseFS.ErrPathNotSupported
	}
	ocfs, err := NewFS(path, b.factory, b.digestAlgorithms, true, b.noCompression, b.aes, b.aesKey, b.aesIV, b.logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create zipfs")
	}
	return ocfs, nil
}

func (b *BaseFS) GetFS(path string) (ocfl.OCFLFSRead, error) {
	if !b.valid(path) {
		return nil, baseFS.ErrPathNotSupported
	}
	ocfs, err := NewFS(path, b.factory, b.digestAlgorithms, false, b.noCompression, false, nil, nil, b.logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create zipfs")
	}
	return ocfs, nil
}

func (b *BaseFS) Open(path string) (baseFS.ReadSeekCloserStat, error) {
	return nil, baseFS.ErrPathNotSupported
}

func (b *BaseFS) Create(path string) (io.WriteCloser, error) {
	return nil, baseFS.ErrPathNotSupported
}

var (
	_ baseFS.FS = &BaseFS{}
)
