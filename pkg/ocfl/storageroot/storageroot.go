package storageroot

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

type StorageRoot interface {
	fmt.Stringer
	WithFS(fsys fs.FS) StorageRoot
	GetFS() fs.FS
	GetDigest() checksum.DigestAlgorithm
	SetDigest(digest checksum.DigestAlgorithm)
	GetFiles() ([]string, error)
	GetFolders() ([]string, error)
	GetObjectFolders() ([]string, error)
	ObjectExists(id string) (bool, error)
	//LoadObjectByFolder(folder string) (object.Object, error)
	//LoadObjectByID(id string) (object.Object, error)
	//CreateObject(id string, ver version.OCFLVersion, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, manager extension.ExtensionManagerCore) (object.Object, error)
	CreateExtension(fsys fs.FS) (extension.Extension, error)
	CreateExtensions(fsys fs.FS, validation validation.Validation) (extension.Extension, error)
	Check() error
	IdToFolder(id string) (folder string, err error)
	//CheckObjectByFolder(objectFolder string) error
	//CheckObjectByID(objectID string) error
	Init(ver version.OCFLVersion, digest checksum.DigestAlgorithm, manager extension.ExtensionManagerCore) error
	Load() error
	IsModified() bool
	SetModified()
	GetVersion() version.OCFLVersion
	Stat(w io.Writer, path string, id string, statInfo []object.StatInfo) error
	//Extract(fsys fs.FS, path, id, version string, withManifest bool, area string) error
	//ExtractMeta(path, id string) (*object.StorageRootMetadata, error)
}
