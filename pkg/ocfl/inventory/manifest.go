package inventory

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

//var DigestNotFound = errors.New("Digest not found")

type Manifest interface {
	AddFile(filename string, digest string) (bool, error)

	String() string
	IterateFiles() func(yield func(digest string, external []string) bool)
	GetFiles(digest string) ([]string, error)
	Err() error
	Equals(manifest Manifest) bool
	CopyFrom(manifest Manifest) Manifest
	Check(val validation.Validation, version version.OCFLVersion, csFiles map[string][]string) error
	Finalize(val validation.Validation, factory Factory, creation bool) error
	GetDuplicates(digest string) []string
}
