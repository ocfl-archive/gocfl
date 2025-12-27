package inventory

import (
	"iter"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

var VersionNotFound = errors.New("Version not found")

type Versions interface {
	Operations

	String() string
	Iterate() func(yield func(versionString version.OCFLVersion, version Version) bool)
	GetVersionStrings() iter.Seq[version.OCFLVersion]
	Equals(other Versions) bool
	GetVersion(versionString version.OCFLVersion) Version
	IsEmpty() bool
	SetVersion(versionString version.OCFLVersion, ver Version) Versions
	Check(val validation.Validation, manifestDigest []string) error
	Finalize(val validation.Validation, factory Factory, inCreation bool) error
	VersionLessOrEqual(v1, v2 version.OCFLVersion) bool
	LatestVersion() version.OCFLVersion
	FileExists(path, digest string) (bool, error)
	Delete(versionString version.OCFLVersion) (bool, error)
}
