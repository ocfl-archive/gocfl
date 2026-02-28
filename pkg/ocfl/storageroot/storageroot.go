package storageroot

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
)

type Initializer interface {
	StorageRoot
	Init() error
	WithStorageRoot(sr StorageRoot) Initializer
	WithFS(objectFS streamfs.FS) Initializer
	Close() error
}

type Loader interface {
	Load() error
	WithStorageRoot(sr StorageRoot) Loader
	WithFS(sourceFS fs.FS) Loader
	WithExtensionFactory(factory extension.Factory) Loader
	Close() error
}

type StorageRoot interface {
	fmt.Stringer
	GetLoader(extensionFactor extension.Factory) Loader
	GetInitializer() Initializer
	WithReadFS(sourceFS fs.FS) StorageRoot
	GetReadFS() fs.FS
	WithWriteFS(streamFS streamfs.FS) StorageRoot
	GetWriteFS() streamfs.FS
	//WithFS(fsys fs.FS) StorageRoot
	WithExtensionManager(extensionManager ExtensionManager) StorageRoot
	GetExtensionManager() ExtensionManager
	//GetFS() fs.FS
	GetDigest() checksum.DigestAlgorithm
	SetDigest(digest checksum.DigestAlgorithm)
	GetFolders() ([]string, error)
	GetObjectFolders() ([]string, error)
	ObjectExists(id string) (bool, error)
	Check() error
	IdToFolder(id string) (folder string, err error)
	IsModified() bool
	SetModified()
	GetVersion() version.OCFLVersion
	//Stat(w io.Writer, path string, id string, statInfo []object.StatInfo) error
	GetOCFLVersion() version.OCFLVersion
	Stat(w io.Writer, path string, id string, statInfo []object.StatInfo) error
	WithDigestAlgorithm(digest checksum.DigestAlgorithm) StorageRoot
}
