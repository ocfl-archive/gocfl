package inventory

import (
	"iter"
)

//var DigestNotFound = errors.New("Digest not found")

// Manifest represents the mapping of digests to physical file paths.
type Manifest interface {
	// AddFile adds a mapping between a physical file and its digest.
	AddFile(filename string, digest string) (bool, error)

	// String returns a string representation of the manifest.
	String() string
	// Iterate returns an iterator for the manifest entries.
	Iterate() func(yield func(digest string, internal []string) bool)
	// GetFiles returns the physical file paths for a given digest.
	GetFiles(digest string) ([]string, error)
	// Err returns any error encountered during manifest operations.
	Err() error
	// Equals checks if two manifests are equal.
	Equals(manifest Manifest) bool
	// CopyFrom copies entries from another manifest.
	CopyFrom(manifest Manifest) Manifest
	// Check verifies the manifest integrity.
	Check(csFiles map[string][]string, versionDigests []string) error
	// Finalize completes the manifest creation or update.
	Finalize(creation bool) error
	// GetDuplicates returns a list of files that have the same digest.
	GetDuplicates(digest string) []string
	// GetFilesFlat returns an iterator for all physical file paths in the manifest.
	GetFilesFlat() iter.Seq[string]
}
