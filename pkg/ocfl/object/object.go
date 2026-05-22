// Package object contains the interfaces and abstractions for managing OCFL (Oxford Common File Layout) objects.
//
// Unlike the inventory package, which focuses on data structures and JSON representation,
// the object package handles the high-level orchestration of object operations like
// initialization, loading, updating, extraction, and validation.
package object

import (
	"io"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
)

// NamesStruct contains different path representations of a file.
type NamesStruct struct {
	ExternalPaths []string // Paths outside the OCFL object.
	InternalPath  string   // Path within the OCFL object version state.
	ManifestPath  string   // Path within the OCFL object content directory.
}

// VersionWriter is the interface for adding or modifying files in an OCFL object version.
type VersionWriter interface {
	// AddFolder adds all files from a folder to the version.
	AddFolder(sourceFS fs.FS, checkDuplicate bool, area string) error
	// AddFile adds a single file from a filesystem to the version.
	AddFile(sourceFS fs.FS, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error
	// AddData adds data from a byte slice as a file to the version.
	AddData(data []byte, path string, checkDuplicate bool, area string, noExtensionHook bool, isDir bool) error
	// AddReader adds data from an io.ReadCloser to the version.
	AddReader(r io.ReadCloser, files []string, area string, noExtensionHook bool, isDir bool) (string, error)
	// DeleteFile removes a file from the version state.
	DeleteFile(virtualFilename string, digest string) error
	// RenameFile changes the path of a file in the version state.
	RenameFile(virtualFilenameSource, virtualFilenameDest string, digest string) error
	// Close finalizes the version and writes the inventory.
	Close() error
	// GetID returns the object ID.
	GetID() string
	// GetFS returns the underlying appendfs.FS.
	GetFS() appendfs.FS
	// BeginArea starts a new area (e.g. for extensions).
	BeginArea(area string)
	// EndArea ends the current area.
	EndArea() error
	// BuildNames constructs a NamesStruct for a list of files.
	BuildNames(files []string, area string) (*NamesStruct, error)
	// WithObject sets the object for the VersionWriter.
	WithObject(obj Object) VersionWriter
	// WithEcho sets whether to echo operations.
	WithEcho(echo bool) VersionWriter
	// Init initializes the VersionWriter with user information.
	Init(msg string, name string, address string) error
	// GetObject returns the associated Object.
	GetObject() Object
}

// Checker is the interface for validating an OCFL object.
type Checker interface {
	// Check performs the validation of the object.
	Check() error
	// WithObject sets the object to be checked.
	WithObject(obj Object) Checker
}

// Initializer is the interface for creating a new OCFL object.
type Initializer interface {
	// Init initializes the object with the given ID and algorithms.
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error
	// WithObject sets the object to be initialized.
	WithObject(o Object) Initializer
}

// Loader is the interface for loading an existing OCFL object.
type Loader interface {
	Object
	// Load reads the object's inventory and state.
	Load() error
	// WithObject sets the object to be loaded.
	WithObject(o Object) Loader
	// GetFS returns the underlying filesystem.
	GetFS() fs.FS
	// SetExtensionFactory sets the factory for extension management.
	SetExtensionFactory(factory extension.Factory[ExtensionManager]) Loader
}

// Extractor is the interface for extracting content from an OCFL object.
type Extractor interface {
	// Extract extracts a specific version of the object.
	Extract(version *inventory.VersionNumber, withManifest bool, area string) error
	// WithObject sets the object to extract from.
	WithObject(o Object) Extractor
	// WithDestFS sets the destination filesystem for extraction.
	WithDestFS(destFS appendfs.FS) Extractor
	// GetExtensionFileReader returns a reader for a file within an extension.
	GetExtensionFileReader(extensionName string, path string) (io.ReadCloser, int64, string, error)
	// GetFileReader returns a reader for a file in the object.
	GetFileReader(name string) (io.ReadCloser, int64, string, error)
	// GetMetadata returns the metadata of the object.
	GetMetadata() (*inventory.Metadata, error)
}

// Object is the main interface representing an OCFL object.
type Object interface {
	// GetExtractor returns an Extractor for this object.
	GetExtractor() Extractor
	// GetInitializer returns an Initializer for this object.
	GetInitializer() Initializer
	// GetLoader returns a Loader for this object.
	GetLoader() Loader
	// WithInventory sets the inventory for the object.
	WithInventory(inv inventory.Inventory) Object
	// WithExtensionManager sets the extension manager for the object.
	WithExtensionManager(manager ExtensionManager) Object
	// WithReadFS sets the read-only filesystem for the object.
	WithReadFS(fsys fs.FS) Object
	// WithWriteFS sets the writable filesystem for the object.
	WithWriteFS(fsys appendfs.FS) Object
	// GetReadFS returns the read-only filesystem.
	GetReadFS() fs.FS
	// GetWriteFS returns the writable filesystem.
	GetWriteFS() appendfs.FS
	// StartUpdate begins a new version update.
	StartUpdate(msg string, UserName string, UserAddress string, echo bool) (VersionWriter, error)
	// GetID returns the object ID.
	GetID() string
	// GetInventory returns the current inventory.
	GetInventory() inventory.Inventory
	// Stat writes statistical information about the object.
	Stat(w io.Writer, statInfo []StatInfo) error
	// GetExtensionManager returns the extension manager.
	GetExtensionManager() ExtensionManager
	// GetOCFLVersion returns the OCFL version of the object.
	GetOCFLVersion() version.OCFLVersion
	// GetChecker returns a Checker for this object.
	GetChecker() Checker
	// Close finalizes the object.
	Close() error
}
