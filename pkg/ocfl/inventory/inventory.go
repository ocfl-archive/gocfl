package inventory

import (
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
)

// StateFileCallback is a function type for iterating over state files.
type StateFileCallback func(internal []string, external []string, digest string) error

// Inventory is the main interface for an OCFL inventory.
type Inventory interface {
	// Finalize completes the inventory creation or update.
	Finalize(inCreation bool) error
	// Equals checks if two inventories are equal.
	Equals(i2 Inventory) bool
	// WithID sets the OCFL object ID.
	WithID(id string) Inventory
	// GetID returns the OCFL object ID.
	GetID() string
	// WithDigestAlgorithm sets the primary digest algorithm.
	WithDigestAlgorithm(algorithm checksum.DigestAlgorithm) Inventory
	// GetDigestAlgorithm returns the primary digest algorithm.
	GetDigestAlgorithm() checksum.DigestAlgorithm
	// WithContentDir sets the content directory name.
	WithContentDir(contentDir string) Inventory
	// GetContentDir returns the content directory name.
	GetContentDir() string
	// GetRealContentDir returns the actual content directory path.
	GetRealContentDir() string
	// GetHead returns the current head version number.
	GetHead() *VersionNumber
	// GetSpec returns the OCFL inventory specification version.
	GetSpec() InventorySpec
	// CheckFiles verifies the integrity of files against the manifest.
	CheckFiles(fileManifest map[checksum.DigestAlgorithm]map[string][]string) error
	// Bytes returns the serialized inventory JSON and its checksum.
	Bytes() (inventory []byte, checksumString string, err error)
	// GetOCFLVersion returns the OCFL specification version.
	GetOCFLVersion() version.OCFLVersion

	// AddFile adds a new file to the inventory.
	AddFile(stateFilenames []string, manifestFilename string, checksums map[checksum.DigestAlgorithm]string) error

	// IterateFiles iterates over all files in a specific version.
	IterateFiles(version *VersionNumber, fn StateFileCallback) error

	// WithVersions sets the version history.
	WithVersions(versions Versions) Inventory
	// GetVersions returns the version history.
	GetVersions() Versions
	// WithManifest sets the file manifest.
	WithManifest(manifest Manifest) Inventory
	// GetManifest returns the file manifest.
	GetManifest() Manifest
	// WithFixity sets the fixity information.
	WithFixity(fixity Fixity) Inventory
	// GetFixity returns the fixity information.
	GetFixity() Fixity
	// WithWriteable enables write operations on the inventory.
	WithWriteable() Inventory
	// IsWriteable returns true if the inventory is writeable.
	IsWriteable() bool
	// IsModified returns true if the inventory has been modified.
	IsModified() bool
	// BuildManifestName generates a manifest path for a given state file.
	BuildManifestName(stateFilename string) string
	// BuildManifestNameVersion generates a manifest path for a given state file in a specific version.
	BuildManifestNameVersion(stateFilename string, version *VersionNumber) string
	// AlreadyExists checks if a file with the given name and checksum already exists in the inventory.
	AlreadyExists(stateFilename, checksum string) (bool, error)
	// Clean removes unused entries and temporary data.
	Clean() error
}
