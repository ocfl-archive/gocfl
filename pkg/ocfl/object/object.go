package object

import (
	"io"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

type NamesStruct struct {
	ExternalPaths []string
	InternalPath  string
	ManifestPath  string
}

type VersionWriter interface {
	AddFolder(fsys fs.FS, versionFS fs.FS, checkDuplicate bool, area string) error
	AddFile(fsys fs.FS, versionFS fs.FS, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error
	AddData(data []byte, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error
	AddReader(r io.ReadCloser, files []string, area string, noExtensionHook bool, isDir bool) (string, error)
	DeleteFile(virtualFilename string, digest string) error
	RenameFile(virtualFilenameSource, virtualFilenameDest string, digest string) error
}

type Object interface {
	VersionWriter
	WithFS(fsys fs.FS) Object
	GetFS() fs.FS

	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, manager extension.ExtensionManagerCore) error
	Load() error

	GetID() string
	GetOCFLVersion() version.OCFLVersion
	Close() error
	IsModified() bool

	LoadInventory(folder string) (inventory.Inventory, error)
	//CreateInventory(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) (Inventory, error)
	StoreInventory(version bool, objectRoot bool) error
	GetInventory() inventory.Inventory

	StoreExtensions() error
	StartUpdate(sourceFS fs.FS, msg string, UserName string, UserAddress string, echo bool) (fs.FS, error)
	EndUpdate() error

	BeginArea(area string)
	EndArea() error
	GetAreaPath(area string) (string, error)

	Check() error

	Stat(w io.Writer, statInfo []StatInfo) error
	Extract(fsys fs.FS, version *inventory.VersionNumber, withManifest bool, area string) error
	GetMetadata() (*inventory.Metadata, error)
	GetExtensionManager() ExtensionManager
	BuildNames(files []string, area string) (*NamesStruct, error)
}
