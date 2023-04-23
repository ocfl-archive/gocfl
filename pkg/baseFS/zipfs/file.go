//go:build exclude

package zipfs

import (
	"emperror.dev/errors"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"io"
	"io/fs"
)

type File struct {
	*FileInfo
	r io.ReadCloser
}

type readCloser2ReadSeekStat struct {
	*FileInfo
	io.ReadCloser
}

func (rc2rss *readCloser2ReadSeekStat) Stat() (fs.FileInfo, error) {
	return rc2rss.FileInfo, nil
}

func (rc2rss *readCloser2ReadSeekStat) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("seek not implemented")
}

func NewFile(info *FileInfo) (ocfl.FileSeeker, error) {
	var err error

	f := &File{
		FileInfo: info,
	}
	if f.r, err = info.zf.Open(); err != nil {
		return nil, errors.Wrapf(err, "cannot open zip item %s", info.Name())
	}
	return &readCloser2ReadSeekStat{
		FileInfo:   info,
		ReadCloser: f,
	}, nil
}

func (f *File) Stat() (fs.FileInfo, error) { return f.FileInfo, nil }

func (f *File) Read(p []byte) (int, error) {
	num, err := f.r.Read(p)
	if err != nil && err != io.EOF {
		return num, errors.Wrapf(err, "cannot read from %s", f.Name())
	}
	return num, err
}

func (f *File) Close() error {
	if err := f.r.Close(); err != nil {
		return errors.Wrapf(err, "cannot close %s", f.Name())
	}
	return nil
}
