package inventory

import (
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

type Fixity interface {
	String() string
	IterateFiles() func(yield func(digest string, external []string) bool)
	GetFiles(alg checksum.DigestAlgorithm, digest string) ([]string, error)
	Err() error
	Equals(fixity Fixity) bool
	CopyFrom(fixity Fixity) State
	Check(val validation.Validator, version string, manifestDigests []string, manifestDigestsLower []string) error
}
