package helper

import (
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"io"
	"io/fs"
	"slices"
)

func NewMultipartFileReader(fsys fs.FS, names []string, digestAlgorithm checksum.DigestAlgorithm) (*MultipartFileReader, error) {
	if len(names) == 0 {
		return nil, errors.New("no files provided for MultipartFileReader")
	}
	mpr := &MultipartFileReader{
		parts:           names,
		digestAlgorithm: digestAlgorithm,
		fsys:            fsys,
		partChecksum:    make(map[string]string),
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
	fp              io.ReadCloser
	currPart        int // current part index
	parts           []string
	fsys            fs.FS
	digestAlgorithm checksum.DigestAlgorithm
	checksumWriter  *checksum.ChecksumWriter
	partChecksum    map[string]string // checksum for each part
}

func (m *MultipartFileReader) GetPartsChecksum() map[string]string {
	return m.partChecksum
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

func (m *MultipartFileReader) nextPart() error {
	var err error
	if err := m.fp.Close(); err != nil {
		return errors.Wrapf(err, "cannot close file %d - %s", m.currPart-1, m.parts[m.currPart-1])
	}
	if err := m.checksumWriter.Close(); err != nil {
		return errors.Wrap(err, "cannot close checksum writer before opening next part")
	}
	checksums, err := m.checksumWriter.GetChecksums()
	if err != nil {
		return errors.Wrapf(err, "cannot get checksums for part %d - %s", m.currPart, m.parts[m.currPart])
	}

	m.fp = nil             // reset file pointer
	m.checksumWriter = nil // reset checksum writer

	if m.currPart >= len(m.parts)-1 {
		return io.EOF // no more parts to read
	}

	var ok bool
	m.partChecksum[m.parts[m.currPart]], ok = checksums[m.digestAlgorithm]
	if !ok {
		return errors.Errorf("checksum for algorithm %s not found for part %d - %s", m.digestAlgorithm, m.currPart, m.parts[m.currPart])
	}
	m.currPart++
	m.checksumWriter, err = checksum.NewChecksumWriter([]checksum.DigestAlgorithm{m.digestAlgorithm})
	if err != nil {
		return errors.Wrapf(err, "cannot create checksum writer for next part %s", m.parts[m.currPart+1])
	}
	if m.fp, err = m.fsys.Open(m.parts[m.currPart]); err != nil {
		return errors.Wrapf(err, "cannot open next file %s", m.parts[m.currPart])
	}
	return nil
}

func (m *MultipartFileReader) Read(p []byte) (n int, err error) {
	if m.fp == nil {
		return 0, io.EOF // no file opened
	}
	n, err = m.fp.Read(p)
	if n > 0 {
		if _, err := m.checksumWriter.Write(p[:n]); err != nil {
			return n, errors.Wrapf(err, "cannot write to checksum writer for part %d - %s", m.currPart, m.parts[m.currPart])
		}
	}
	if err != nil {
		if errors.Is(err, io.EOF) {
			if err := m.nextPart(); err != nil {
				if errors.Is(err, io.EOF) {
					return n, io.EOF // all parts are read
				}
				return n, errors.Wrapf(err, "cannot read next part %d - %s", m.currPart, m.parts[m.currPart])
			}
		} else {
			return n, errors.Wrapf(err, "cannot read from file %d - %s", m.currPart, m.parts[m.currPart])
		}
	}
	return n, nil
}

var _ io.Reader = (*MultipartFileReader)(nil)
