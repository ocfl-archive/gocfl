package inventory

import (
	"iter"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
)

var DigestAlgNotFound = errors.New("digest algorithm not found")
var FixityTypeDifferent = errors.New("fixity type different")

// Fixity represents optional additional fixity information for physical files.
type Fixity interface {
	// AddFile adds fixity digests for a physical file.
	AddFile(manifestFilename string, digests map[checksum.DigestAlgorithm]string) (bool, error)
	// WithAlgorithms sets the algorithms to be used for fixity.
	WithAlgorithms(algorithms ...checksum.DigestAlgorithm) Fixity
	// WithAllowedAlgorithms sets the algorithms that are allowed for fixity.
	WithAllowedAlgorithms(algorithms ...checksum.DigestAlgorithm) Fixity
	// String returns a string representation of the fixity information.
	String() string
	// Iterate returns an iterator for fixity entries of a specific algorithm.
	Iterate(alg checksum.DigestAlgorithm) func(yield func(digest string, external []string) bool)
	// GetDigestAlgorithms returns an iterator for all algorithms used in fixity.
	GetDigestAlgorithms() iter.Seq[checksum.DigestAlgorithm]
	// GetFiles returns the physical file paths for a given algorithm and digest.
	GetFiles(alg checksum.DigestAlgorithm, digest string) ([]string, error)
	// Err returns any error encountered during fixity operations.
	Err() error
	// Equals checks if two fixity objects are equal.
	Equals(fixity Fixity) bool
	// CopyFrom copies entries from another fixity object.
	CopyFrom(fixity Fixity) error
	// Check verifies the fixity integrity.
	Check(fileManifest map[checksum.DigestAlgorithm]map[string][]string) error
	// Finalize completes the fixity creation or update.
	Finalize(inCreation bool) error
	// Checksums returns all fixity checksums for a given file path.
	Checksums(s string) map[checksum.DigestAlgorithm]string
}
