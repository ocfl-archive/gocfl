package interfaces

import (
	"iter"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
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
	Check(val validation.Validation, manifestDigest []string) error
	Finalize(val validation.Validation, factory Factory, inCreation bool) error
	LatestVersionNumber() *VersionNumber
	FileExists(path, digest string) (bool, error)
	Delete(versionNumber *VersionNumber) (bool, error)
	Err() error
}
