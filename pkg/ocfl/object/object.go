package object

import (
	"io"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

// NamessStruct provides a mechanism to record the logical and virtual
// state of paths within an OCFL object.
type NamesStruct struct {
	ExternalPaths []string
	InternalPath  string
	ManifestPath  string
}

// VersionWriter is the interface that defines read/write operations
// across different versions of an OCFL object.
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

// Checker is the interface that defines rudimentary validation
// operations that should be availabnle to baseline OCFL objects
// in memory.
type Checker interface {
	Check() error
	WithObject(obj Object) Checker
	WithFS(objectFS fs.FS) Checker
}

// Initializer is the interface that defines operations for initializing
// an OCFL object.
type Initializer interface {
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error
	WithObject(o Object) Initializer
	WithFS(objectFS appendfs.FS) Initializer
}

// Loader is the interface that defines methods for loading an OCFL
// object into memory including its extensions and inventories if
// available to a version.
type Loader interface {
	Object
	Load() error
	WithObject(o Object) Loader
	WithFS(sourceFS fs.FS) Loader
	GetFS() fs.FS
	WithExtensionFactory(factory extension.Factory) Loader
}

// Extractor is the interface that defines methods for extracting
// information from an OCFL object.
type Extractor interface {
	Extract(version *inventory.VersionNumber, withManifest bool, area string) error
	WithObject(o Object) Extractor
	WithFS(objectFS fs.FS, destFS appendfs.FS) Extractor
	GetExtensionFileReader(extensionName string, path string) (io.ReadCloser, int64, string, error)
	GetFileReader(name string) (io.ReadCloser, int64, string, error)
	GetMetadata() (*inventory.Metadata, error)
}

// Object is the interface that defines the core characteristics and
// operations available to an OCFL object.
type Object interface {
	GetExtractor(objectFS fs.FS, destFS appendfs.FS) Extractor
	GetInitializer(objectFS appendfs.FS) Initializer
	GetLoader(sourceFS fs.FS, extensionFactory extension.Factory) Loader
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
