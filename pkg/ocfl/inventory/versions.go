package inventory

import (
	"iter"

	"emperror.dev/errors"
)

var VersionNotFound = errors.New("Version not found")

type Versions interface {
	Operations

	String() string
	Iterate() func(yield func(versionNumber *VersionNumber, version Version) bool)
	GetVersionNumbers() iter.Seq[*VersionNumber]
	Equals(other Versions) bool
	GetVersion(versionNumber *VersionNumber) Version
	IsEmpty() bool
	SetVersion(versionNumber *VersionNumber, ver Version) Versions
	Check(manifestDigests []string) error
	Finalize(inCreation bool) error
	LatestVersionNumber() *VersionNumber
	FileExists(path, digest string) (bool, error)
	Delete(versionNumber *VersionNumber) (bool, error)
	Err() error
	NewVersion(msg string, UserName string, UserAddress string) error
}
