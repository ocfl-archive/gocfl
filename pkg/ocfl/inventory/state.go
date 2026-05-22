package inventory

import (
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
)

var DigestNotFound = errors.New("Digest not found")
var StateTypeDifferent = errors.New("State type different")

// State represents the logical state of files in a specific version.
type State interface {
	// DeleteFile removes a file from the state.
	DeleteFile(stateFilename string) (bool, error)
	// RenameFile changes the logical path of a file.
	RenameFile(oldStateFilename, newStateFilename string) (bool, error)
	// CopyFile copies a file within the state.
	CopyFile(stateFilename, digest string) (bool, error)
	// EchoDelete performs a dry-run or logging of file deletions.
	EchoDelete(existing []string, pathPrefix string) (bool, error)
	// AddFile adds a file to the state.
	AddFile(stateFilename string, digest string) (bool, error)

	// String returns a string representation of the state.
	String() string
	// Iterate returns an iterator for state entries.
	Iterate() func(yield func(digest string, external []string) bool)
	// GetFiles returns the logical file paths for a given digest.
	GetFiles(digest string) ([]string, error)
	// Err returns any error encountered during state operations.
	Err() error
	// Equals checks if two states are equal.
	Equals(state State) bool
	// CopyFrom copies entries from another state.
	CopyFrom(state State) error
	// Check verifies the state integrity.
	Check(version *VersionNumber, manifestDigests []string, manifestDigestsLower []string) error
	// FileChecksum returns the digest for a given logical file path.
	FileChecksum(path string) string
	// WithDigestAlgorithm sets the digest algorithm used by this state.
	WithDigestAlgorithm(dgst checksum.DigestAlgorithm) State
}
