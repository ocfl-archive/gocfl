package types

import (
	"iter"

	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

//var DigestNotFound = errors.New("Digest not found")

type Manifest interface {
	AddFile(filename string, digest string) (bool, error)

	String() string
	Iterate() func(yield func(digest string, internal []string) bool)
	GetFiles(digest string) ([]string, error)
	Err() error
	Equals(manifest Manifest) bool
	CopyFrom(manifest Manifest) Manifest
	Check(val validation.Validation, csFiles map[string][]string, versionDigests []string) error
	Finalize(val validation.Validation, factory Factory, creation bool) error
	GetDuplicates(digest string) []string
	GetFilesFlat() iter.Seq[string]
}
