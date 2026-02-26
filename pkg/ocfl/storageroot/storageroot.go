package storageroot

import (
	"fmt"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
)

type Initializer interface {
	StorageRoot
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error
	WithStorageRoot(sr StorageRoot) Initializer
	WithFS(objectFS streamfs.FS) Initializer
}

type Loader interface {
	Load() error
	WithStorageRoot(sr StorageRoot) Loader
	WithFS(sourceFS fs.FS) Loader
	WithExtensionFactory(factory extension.Factory) Loader
}

type StorageRoot interface {
	fmt.Stringer
	GetLoader(sourceFS fs.FS, extensionFactor extension.Factory) Loader
	GetInitializer(storageRootFS streamfs.FS) Initializer
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
}
