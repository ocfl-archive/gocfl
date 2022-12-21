package s3fs

import "io/fs"

type DirEntry struct{ *FileInfo }

func (s DirEntry) Type() fs.FileMode {
	return s.Mode()
}

func (s DirEntry) Info() (fs.FileInfo, error) {
	return s.FileInfo, nil
}

var _ fs.DirEntry = &DirEntry{}
