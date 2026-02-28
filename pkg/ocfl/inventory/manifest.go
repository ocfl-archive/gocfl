package inventory

import (
	"iter"
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
	Check(csFiles map[string][]string, versionDigests []string) error
	Finalize(creation bool) error
	GetDuplicates(digest string) []string
	GetFilesFlat() iter.Seq[string]
}
