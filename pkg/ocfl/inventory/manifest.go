package inventory

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

//var DigestNotFound = errors.New("Digest not found")

type Manifest interface {
	String() string
	IterateFiles() func(yield func(digest string, external []string) bool)
	GetFiles(digest string) ([]string, error)
	Err() error
	Equals(manifest Manifest) bool
	CopyFrom(manifest Manifest) Manifest
	Check(val validation.Validation, version string, manifestDigests []string, manifestDigestsLower []string) error
	Finalize(val validation.Validation, factory Factory, creation bool) error
}
