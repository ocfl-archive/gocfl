package ocfl

import (
	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/zipfs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"io"
	"io/fs"
	"slices"
	"strings"
)

func NewMultipartFileReader(fsys fs.FS, names []string, digestAlgorithm checksum.DigestAlgorithm) (*MultipartFileReader, error) {
	if len(names) == 0 {
		return nil, errors.New("no files provided for MultipartFileReader")
	}
	mpr := &MultipartFileReader{
		parts: []string{},
		fsys:  fsys,
	}
	slices.Sort(names)
	var err error
	if mpr.fp, err = fsys.Open(names[0]); err != nil {
		return nil, errors.Wrapf(err, "cannot open first file %s", names[0])
	}
	if mpr.checksumWriter, err = checksum.NewChecksumWriter([]checksum.DigestAlgorithm{digestAlgorithm}); err != nil {
		mpr.fp.Close()
		return nil, errors.Wrapf(err, "cannot create checksum writer for first file %s", names[0])
	}
	return mpr, nil
}

type MultipartFileReader struct {
	fp             io.ReadCloser
	currPart       int // current part index
	parts          []string
	fsys           fs.FS
	checksumWriter *checksum.ChecksumWriter
}

func (m *MultipartFileReader) GetChecksums() (map[checksum.DigestAlgorithm]string, error) {
	checksums, err := m.checksumWriter.GetChecksums()
	if err != nil {
		return nil, errors.Wrap(err, "cannot get checksums from checksum writer")
	}
	return checksums, nil
}

func (m *MultipartFileReader) Close() error {
	var errs = []error{}
	if err := m.fp.Close(); err != nil {
		errs = append(errs, errors.Wrapf(err, "cannot close file %d - %s", m.currPart, m.parts[m.currPart]))
	}
	if err := m.checksumWriter.Close(); err != nil {
		errs = append(errs, errors.Wrap(err, "cannot close checksum writer"))
	}
	return errors.Combine(errs...)
}

func (m *MultipartFileReader) Read(p []byte) (n int, err error) {
	n, err = m.fp.Read(p)
	if err != nil {
		if errors.Is(err, io.EOF) {
			if m.currPart < len(m.parts)-1 {
				m.currPart++
				if m.fp, err = m.fsys.Open(m.parts[m.currPart]); err != nil {
					return 0, errors.Wrapf(err, "cannot open next file %s", m.parts[m.currPart])
				}
				return n, nil // continue reading from the next part
			} else {
				return n, io.EOF // all parts are read
			}
		}
		return n, errors.Wrapf(err, "cannot read from file %d - %s", m.currPart, m.parts[m.currPart])
	}
	return n, nil
}

func NewMultiZIPFS(fsys fs.FS, names []string, logger zLogger.ZLogger) (fs.FS, io.Closer, error) {
	if len(names) == 0 {
		return nil, nil, errors.New("no files provided for ZIPFSReaderAt")
	}
	mpr, err := NewMultipartFileReaderAt(fsys, names)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot create MultipartFileReaderAt")
	}
	zfs, err := zipfs.NewFS(mpr, mpr.GetSize(), strings.Join(names, ", "), logger)
	if err != nil {
		mpr.Close()
		return nil, nil, errors.Wrap(err, "cannot create ZIPFS from MultipartFileReaderAt")
	}
	return zfs, mpr, nil
}

var _ readAtSeekCloser = (*MultipartFileReaderAt)(nil)
