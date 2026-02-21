package object

import (
	"io"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
)

type NamesStruct struct {
	ExternalPaths []string
	InternalPath  string
	ManifestPath  string
}

type VersionWriter interface {
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
}

type Checker interface {
	Check() error
}

type Object interface {
	GetChecker(fsys fs.FS) (Checker, error)
	//VersionWriter
	Load(sourceFS fs.FS) error
	StartUpdate(targetFS streamfs.FS, msg string, UserName string, UserAddress string, echo bool) (VersionWriter, error)
	//WithFS(fsys fs.FS) Object
	//GetFS() fs.FS

	Init(targetFS streamfs.FS, id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm, manager extension.ExtensionManagerCore) error

	GetID() string
	//GetOCFLVersion() version.OCFLVersion
	//IsModified() bool

	//CreateInventory(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) (Inventory, error)
	GetInventory() inventory.Inventory

	GetAreaPath(area string) (string, error)

	Stat(w io.Writer, statInfo []StatInfo) error
	Extract(fsys fs.FS, version *inventory.VersionNumber, withManifest bool, area string) error
	GetMetadata() (*inventory.Metadata, error)
	GetExtensionManager() ExtensionManager
	BuildNames(files []string, area string) (*NamesStruct, error)
}
