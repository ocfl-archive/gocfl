// Package ocfl for manipulating and checking Oxford Common Filesystem Layout
// This Oxford Common File Layout (OCFL) specification describes an
// application-independent approach to the storage of digital information in a
// structured, transparent, and predictable manner. It is designed to promote
// long-term object management best practices within digital repositories.
// https://ocfl.io
package ocfl

/*
type FileSeeker interface {
	io.Seeker
	fs.File
	//Stat() (fs.FileInfo, error)
}

type CloserAt interface {
	io.ReaderAt
	io.Closer
}

type OCFLFSRead interface {
	String() string
	OpenSeeker(name string) (FileSeeker, error)
	Open(name string) (fs.File, error)
	Stat(name string) (fs.FileInfo, error)
	ReadFile(name string) ([]byte, error)
	Close() error
	IsNotExist(err error) bool
	WalkDir(root string, fn fs.WalkDirFunc) error
	ReadDir(name string) ([]fs.DirEntry, error)
	HasContent() bool
	SubFS(subfolder string) (OCFLFSRead, error)
}

// Filesystem abstraction for OCFL access
type OCFLFS interface {
	OCFLFSRead
	Create(name string) (io.WriteCloser, error)
	Delete(name string) error
	Discard() error
	SubFSRW(subfolder string) (OCFLFS, error)
	Rename(src, dest string) error
}
*/
