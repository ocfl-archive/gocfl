package storageroot

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
)

type Initializer interface {
	StorageRoot
	Init() error
	WithStorageRoot(sr StorageRoot) Initializer
	WithFS(objectFS appendfs.FS) Initializer
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
	IsWriteable() bool
	GetLoader(extensionFactor extension.Factory) Loader
	GetInitializer() Initializer
	WithReadFS(sourceFS fs.FS) StorageRoot
	GetReadFS() fs.FS
	WithWriteFS(appendFS appendfs.FS) StorageRoot
	GetWriteFS() appendfs.FS
	WithExtensionManager(extensionManager ExtensionManager) StorageRoot
	GetExtensionManager() ExtensionManager
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
	CreateObject(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, objectExtensionManager object.ExtensionManager) (object.Object, error)
}
