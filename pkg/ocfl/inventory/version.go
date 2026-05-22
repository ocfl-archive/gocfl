package inventory

import (
	"time"
)

// Version represents a single version of an OCFL object.
type Version interface {
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

	// String returns a string representation of the version.
	String() string
	// Equals checks if two versions are equal.
	Equals(other Version) bool
	// Finalize completes the version creation or update.
	Finalize(inCreation bool) error
	// WithCreated sets the creation timestamp.
	WithCreated(t time.Time) Version
	// GetCreated returns the creation timestamp.
	GetCreated() time.Time
	// WithMessage sets the version message.
	WithMessage(msg string) Version
	// GetMessage returns the version message.
	GetMessage() string
	// WithState sets the version state.
	WithState(state State) Version
	// GetState returns the version state.
	GetState() State
	// WithUser sets the version user (author).
	WithUser(user User) Version
	// GetUser returns the version user (author).
	GetUser() User
	// WithVersion sets the version number.
	WithVersion(number *VersionNumber) Version
	// GetVersionNumber returns the version number.
	GetVersionNumber() *VersionNumber
	// Check verifies the version integrity against manifest digests.
	Check(manifestDigests, manifestDigestsLower []string) error
	// Err returns any error encountered during version operations.
	Err() error
	// FileChecksum returns the checksum for a given file path in this version.
	FileChecksum(path string) string
	// InCreation returns true if the version is currently being created.
	InCreation() bool
}
