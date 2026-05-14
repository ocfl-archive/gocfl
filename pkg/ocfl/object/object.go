package object

import (
	"io"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
)

type NamesStruct struct {
	ExternalPaths []string
	InternalPath  string
	ManifestPath  string
}

type VersionWriter interface {
	Object
	AddFolder(sourceFS fs.FS, checkDuplicate bool, area string) error
	AddFile(sourceFS fs.FS, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error
	AddData(data []byte, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error
	AddReader(r io.ReadCloser, files []string, area string, noExtensionHook bool, isDir bool) (string, error)
	DeleteFile(virtualFilename string, digest string) error
	RenameFile(virtualFilenameSource, virtualFilenameDest string, digest string) error
	Close() error
	GetID() string
	GetFS() appendfs.FS
	BeginArea(area string)
	EndArea() error
	BuildNames(files []string, area string) (*NamesStruct, error)
}

type Checker interface {
	Check() error
	SetObject(obj Object) Checker
	SetFS(objectFS fs.FS) Checker
}

type Initializer interface {
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error
	SetObject(o Object) Initializer
	WithFS(objectFS appendfs.FS) Initializer
}

type Loader interface {
	Object
	Load() error
	SetObject(o Object) Loader
	WithFS(sourceFS fs.FS) Loader
	GetFS() fs.FS
	SetExtensionFactory(factory extension.Factory[ExtensionManager]) Loader
}

type Extractor interface {
	Extract(version *inventory.VersionNumber, withManifest bool, area string) error
	SetObject(o Object) Extractor
	WithFS(objectFS fs.FS, destFS appendfs.FS) Extractor
	GetExtensionFileReader(extensionName string, path string) (io.ReadCloser, int64, string, error)
	GetFileReader(name string) (io.ReadCloser, int64, string, error)
	GetMetadata() (*inventory.Metadata, error)
}

type Object interface {
	GetExtractor() Extractor
	GetInitializer() Initializer
	GetLoader() Loader
	WithInventory(inv inventory.Inventory) Object
	WithExtensionManager(manager ExtensionManager) Object
	StartUpdate(objectFS appendfs.FS, msg string, UserName string, UserAddress string, echo bool) (VersionWriter, error)
	GetID() string
	GetInventory() inventory.Inventory
	//GetAreaPath(area string) (string, error)
	Stat(w io.Writer, statInfo []StatInfo) error
	GetExtensionManager() ExtensionManager
	GetOCFLVersion() version.OCFLVersion
	GetChecker(sourceFS fs.FS) Checker
}
