package inventory

import (
	"iter"

	"emperror.dev/errors"
)

var VersionNotFound = errors.New("Version not found")

// Versions represents the history of versions in an OCFL object.
type Versions interface {
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

	// String returns a string representation of the versions.
	String() string
	// Iterate returns an iterator for all versions.
	Iterate() func(yield func(versionNumber *VersionNumber, version Version) bool)
	// GetVersionNumbers returns an iterator for all version numbers.
	GetVersionNumbers() iter.Seq[*VersionNumber]
	// Equals checks if two versions collections are equal.
	Equals(other Versions) bool
	// GetVersion returns a specific version by its number.
	GetVersion(versionNumber *VersionNumber) Version
	// IsEmpty returns true if there are no versions.
	IsEmpty() bool
	// SetVersion adds or updates a version in the history.
	SetVersion(versionNumber *VersionNumber, ver Version) Versions
	// Check verifies the integrity of all versions.
	Check(manifestDigests []string) error
	// Finalize completes the creation or update of the versions history.
	Finalize(inCreation bool) error
	// LatestVersionNumber returns the number of the most recent version.
	LatestVersionNumber() *VersionNumber
	// FileExists checks if a file with the given path and digest exists in any version.
	FileExists(path, digest string) (bool, error)
	// Delete removes a specific version from the history.
	Delete(versionNumber *VersionNumber) (bool, error)
	// Err returns any error encountered during operations.
	Err() error
	// NewVersion creates and adds a new version to the history.
	NewVersion(msg string, UserName string, UserAddress string) error
}
