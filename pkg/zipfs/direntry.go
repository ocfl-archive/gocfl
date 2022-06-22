package zipfs

import "io/fs"

type DirEntry FileInfo

func NewDirEntry(fi *FileInfo) *DirEntry { return (*DirEntry)(fi) }

func (de *DirEntry) Name() string {
	return ((*FileInfo)(de)).Name()
}

func (de *DirEntry) IsDir() bool {
	return ((*FileInfo)(de)).IsDir()
}

func (de *DirEntry) Info() (fs.FileInfo, error) {
	return (*FileInfo)(de), nil
}

func (de *DirEntry) Type() fs.FileMode {
	return ((*FileInfo)(de)).Mode()
}
