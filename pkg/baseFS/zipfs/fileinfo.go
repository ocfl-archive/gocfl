//go:build exclude

package zipfs

import (
	"archive/zip"
	"io/fs"
	"strings"
	"time"
)

type FileInfo struct {
	zf      *zip.File
	dirName string
}

func NewFileInfoFile(zf *zip.File) (*FileInfo, error) {
	fi := &FileInfo{zf: zf}
	return fi, nil
}

func NewFileInfoDir(dirName string) (*FileInfo, error) {
	fi := &FileInfo{dirName: dirName}
	return fi, nil
}

func (fi *FileInfo) Name() string {
	var parts []string
	if fi.zf != nil {
		parts = strings.Split(fi.zf.Name, "/")
	} else {
		parts = strings.Split(fi.dirName, "/")
	}
	return parts[len(parts)-1]
}

func (fi *FileInfo) Size() int64 {
	if fi.zf != nil {
		return int64(fi.zf.UncompressedSize64)
	} else {
		return 0
	}
}

func (fi *FileInfo) Mode() fs.FileMode {
	if fi.zf != nil {
		return fi.zf.Mode()
	} else {
		return 0777
	}
}

func (fi *FileInfo) ModTime() time.Time {
	if fi.zf != nil {
		return fi.zf.Modified
	} else {
		return time.Time{}
	}
}

func (fi *FileInfo) IsDir() bool {
	return fi.zf == nil
}

func (fi *FileInfo) Sys() any { return nil }
