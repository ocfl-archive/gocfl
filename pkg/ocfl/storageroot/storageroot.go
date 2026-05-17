// Package storageroot contains the interfaces and abstractions for managing OCFL (Oxford Common File Layout) storage roots.
//
// A storage root is the top-level directory in an OCFL filesystem, containing OCFL objects
// and root-level configuration like the storage layout. This package provides interfaces
// for initializing, loading, and interacting with storage roots.
package storageroot

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/filesystem/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/object"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
)

// Initializer is the interface for creating a new OCFL storage root.
type Initializer interface {
	StorageRoot
	// Init initializes the storage root.
	Init() error
	// SetStorageRoot sets the storage root for the Initializer.
	SetStorageRoot(sr StorageRoot) Initializer
	// Close finalizes the initialization.
	Close() error
}

// Loader is the interface for loading an existing OCFL storage root.
type Loader interface {
	// Load reads the storage root configuration and state.
	Load() error
	// SetStorageRoot sets the storage root for the Loader.
	SetStorageRoot(sr StorageRoot) Loader
	// SetExtensionFactory sets the factory for extension management.
	SetExtensionFactory(factory extension.Factory[ExtensionManager]) Loader
	// Close finalizes the loading process.
	Close() error
}

// StorageRoot is the main interface representing an OCFL storage root.
type StorageRoot interface {
	fmt.Stringer
	// IsWriteable returns true if the storage root is writable.
	IsWriteable() bool
	// GetLoader returns a Loader for this storage root.
	GetLoader() Loader
	// GetInitializer returns an Initializer for this storage root.
	GetInitializer() Initializer
	// WithReadFS sets the read-only filesystem for the storage root.
	WithReadFS(sourceFS fs.FS) StorageRoot
	// GetReadFS returns the read-only filesystem.
	GetReadFS() fs.FS
	// WithWriteFS sets the writable filesystem for the storage root.
	WithWriteFS(appendFS appendfs.FS) StorageRoot
	// GetWriteFS returns the writable filesystem.
	GetWriteFS() appendfs.FS
	// WithExtensionManager sets the extension manager for the storage root.
	WithExtensionManager(extensionManager ExtensionManager) StorageRoot
	// GetExtensionManager returns the extension manager.
	GetExtensionManager() ExtensionManager
	// GetDigest returns the default digest algorithm for the storage root.
	GetDigest() checksum.DigestAlgorithm
	// SetDigest sets the default digest algorithm for the storage root.
	SetDigest(digest checksum.DigestAlgorithm)
	// GetFolders returns all folders in the storage root.
	GetFolders() ([]string, error)
	// GetObjectFolders returns all folders containing OCFL objects.
	GetObjectFolders() ([]string, error)
	// ObjectExists checks if an object with the given ID exists in the storage root.
	ObjectExists(id string) (bool, error)
	// Check performs a validation check on the storage root.
	Check() error
	// IdToFolder maps an object ID to its storage folder path using the storage layout.
	IdToFolder(id string) (folder string, err error)
	// IsModified returns true if the storage root has been modified.
	IsModified() bool
	// SetModified marks the storage root as modified.
	SetModified()
	// GetVersion returns the OCFL version of the storage root.
	GetVersion() version.OCFLVersion
	// GetOCFLVersion returns the OCFL version of the storage root.
	GetOCFLVersion() version.OCFLVersion
	// Stat writes statistical information about the storage root and its objects.
	Stat(w io.Writer, path string, id string, statInfo []object.StatInfo) error
	// WithDigestAlgorithm sets the default digest algorithm and returns the storage root.
	WithDigestAlgorithm(digest checksum.DigestAlgorithm) StorageRoot
}
