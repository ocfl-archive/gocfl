package object

import (
	"io"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
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
	BeginArea(area string)
	EndArea() error
	BuildNames(files []string, area string) (*NamesStruct, error)
}

type Checker interface {
	Check() error
	WithObject(o Object) Checker
	WithFS(fsys fs.FS) Checker
}

type Initializer interface {
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error
	WithObject(o Object) Initializer
	WithFS(objectFS streamfs.FS) Initializer
}

type Loader interface {
	Load() error
	WithObject(o Object) Loader
	WithFS(sourceFS fs.FS) Loader
	WithExtensionFactory(factory extension.Factory) Loader
}

type Extractor interface {
	Extract(version *inventory.VersionNumber, withManifest bool, area string) error
	WithObject(o Object) Extractor
	WithFS(sourceFS fs.FS, objectFS streamfs.FS) Extractor
}

type Object interface {
	GetChecker(objectFS fs.FS) Checker
	GetExtractor(objectFS fs.FS) Extractor
	GetInitializer(objectFS streamfs.FS) Initializer
	GetLoader(sourceFS fs.FS, extensionFactory extension.Factory) Loader
	WithInventory(inv inventory.Inventory) Object
	WithExtensionManager(manager ExtensionManager) Object
	StartUpdate(objectFS streamfs.FS, msg string, UserName string, UserAddress string, echo bool) (VersionWriter, error)
	GetID() string
	GetInventory() inventory.Inventory
	//GetAreaPath(area string) (string, error)
	Stat(w io.Writer, statInfo []StatInfo) error
	GetMetadata() (*inventory.Metadata, error)
	GetExtensionManager() ExtensionManager
	GetOCFLVersion() version.OCFLVersion
}
