package inventory

import (
	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

var VersionNotFound = errors.New("Version not found")

type Versions interface {
	String() string
	Iterate() func(yield func(versionString string, version Version) bool)
	Equals(other Versions) bool
	GetVersion(versionString string) Version
	IsEmpty() bool
	SetVersion(versionString string, ver Version) Versions
	Check(val validation.Validation, manifestDigest []string) error
	Finalize(val validation.Validation, factory Factory, inCreation bool) error
}
