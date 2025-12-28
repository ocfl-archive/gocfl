package inventory

import (
	"iter"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

var VersionNotFound = errors.New("Version not found")

type Versions interface {
	Operations

	String() string
	Iterate() func(yield func(versionString *VersionNumber, version Version) bool)
	GetVersionStrings() iter.Seq[*VersionNumber]
	Equals(other Versions) bool
	GetVersion(versionString *VersionNumber) Version
	IsEmpty() bool
	SetVersion(versionString *VersionNumber, ver Version) Versions
	Check(val validation.Validation, manifestDigest []string) error
	Finalize(val validation.Validation, factory Factory, inCreation bool) error
	VersionLessOrEqual(v1, v2 *VersionNumber) bool
	LatestVersion() *VersionNumber
	FileExists(path, digest string) (bool, error)
	Delete(versionString *VersionNumber) (bool, error)
	AddVersion() error
}
