package types

import (
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/stat"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"

	"io"
	"io/fs"
)

type NamesStruct struct {
	ExternalPaths []string
	InternalPath  string
	ManifestPath  string
}

type Object interface {
	WithFS(fsys fs.FS) Object
	LoadInventory(folder string) (Inventory, error)
	//CreateInventory(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) (Inventory, error)
	StoreInventory(version bool, objectRoot bool) error
	GetInventory() Inventory
	GetInventoryContent() (inventory []byte, checksumString string, err error)
	StoreExtensions() error
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, manager ExtensionManagerCore) error
	Load() error
	StartUpdate(sourceFS fs.FS, msg string, UserName string, UserAddress string, echo bool) (fs.FS, error)
	EndUpdate() error
	BeginArea(area string)
	EndArea() error
	AddFolder(fsys fs.FS, versionFS fs.FS, checkDuplicate bool, area string) error
	AddFile(fsys fs.FS, versionFS fs.FS, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error
	AddData(data []byte, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error
	AddReader(r io.ReadCloser, files []string, area string, noExtensionHook bool, isDir bool) (string, error)
	DeleteFile(virtualFilename string, digest string) error
	RenameFile(virtualFilenameSource, virtualFilenameDest string, digest string) error
	GetID() string
	GetVersion() version.OCFLVersion
	Check() error
	Close() error
	GetFS() fs.FS
	IsModified() bool
	Stat(w io.Writer, statInfo []stat.StatInfo) error
	Extract(fsys fs.FS, version *VersionNumber, withManifest bool, area string) error
	GetMetadata() (*Metadata, error)
	GetAreaPath(area string) (string, error)
	GetExtensionManager() ExtensionManager
	BuildNames(files []string, area string) (*NamesStruct, error)
}
