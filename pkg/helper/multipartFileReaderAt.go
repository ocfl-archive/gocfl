package helper

import (
	"emperror.dev/errors"
	"io"
	"io/fs"
	"slices"
)

type readAtSeekCloser interface {
	io.ReaderAt
	io.Reader
	io.Closer
	io.Seeker
}
type readSeekCloser interface {
	io.Reader
	io.Closer
	io.Seeker
}
type filePart struct {
	fp      readSeekCloser
	size    int64
	name    string
	currPos int64 // current position in the part
}

func NewMultipartFileReaderAt(fsys fs.FS, names []string) (*MultipartFileReaderAt, error) {
	mpr := &MultipartFileReaderAt{
		parts: make([]*filePart, 0, len(names)),
	}
	slices.Sort(names)
	for _, name := range names {
		fp, err := fsys.Open(name)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot open file %s", name)
		}
		fi, err := fp.Stat()
		if err != nil {
			fp.Close()
			mpr.Close()
			return nil, errors.Wrapf(err, "cannot stat file %s", name)
		}
		if fi.IsDir() {
			fp.Close()
			mpr.Close()
			return nil, errors.Errorf("file %s is a directory", name)
		}
		rsc, ok := fp.(readSeekCloser)
		if !ok {
			fp.Close()
			mpr.Close()
			return nil, errors.Errorf("file %s does not implement readSeekCloser interface", name)
		}
		mpr.parts = append(mpr.parts, &filePart{
			fp:   rsc,
			size: fi.Size(),
			name: name,
		})
		mpr.size += fi.Size()
	}
	return mpr, nil
}

type MultipartFileReaderAt struct {
	parts   []*filePart
	currPos int64 // current position in the multipart file reader
	size    int64 // total size of all parts
}

func (m *MultipartFileReaderAt) GetSize() int64 {
	return m.size
}

func (m *MultipartFileReaderAt) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		m.currPos = offset
	case io.SeekCurrent:
		m.currPos += offset
	case io.SeekEnd:
		m.currPos = m.size + offset
	default:
		return 0, errors.Errorf("invalid whence value: %d", whence)
	}
	if m.currPos < 0 {
		return 0, errors.Errorf("negative position: %d", m.currPos)
	}
	return m.currPos, nil
}

func (m *MultipartFileReaderAt) Close() error {
	var errs = []error{}
	for _, part := range m.parts {
		if err := part.fp.Close(); err != nil {
			errs = append(errs, errors.Wrapf(err, "cannot close part %s", part.name))
		}
	}
	m.parts = nil
	return errors.Combine(errs...)
}

func (m *MultipartFileReaderAt) Read(p []byte) (n int, err error) {
	if len(m.parts) == 0 {
		return 0, io.EOF
	}
	currPos := m.currPos
	for _, part := range m.parts {
		if currPos < part.size {
			if currPos != part.currPos {
				if part.currPos, err = part.fp.Seek(currPos, io.SeekStart); err != nil {
					return 0, errors.Wrapf(err, "cannot seek in part %s", part.name)
				}
			}
			n, err = part.fp.Read(p)
			if err != nil && !errors.Is(err, io.EOF) {
				return n, errors.Wrapf(err, "cannot read from part %s", part.name)
			}
			currPos += int64(n)
			part.currPos += int64(n) // update the current position in the part
			m.currPos += int64(n)
			if n > 0 {
				if m.currPos >= m.size {
					return n, io.EOF // return the number of bytes read
				} else {
					return n, nil // return the number of bytes read
				}
			}
		} else {
			currPos -= part.size // move to the next part
		}
	}

	return 0, io.EOF // all parts are read
}

func (m *MultipartFileReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if _, err := m.Seek(off, io.SeekStart); err != nil {
		return 0, errors.Wrapf(err, "cannot seek to offset %d", off)
	}
	return m.Read(p)
}

var _ readAtSeekCloser = (*MultipartFileReaderAt)(nil)
