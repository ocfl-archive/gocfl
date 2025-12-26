package inventory

import (
	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

var DigestNotFound = errors.New("Digest not found")
var StateTypeDifferent = errors.New("State type different")

type State interface {
	Operations

	String() string
	IterateFiles() func(yield func(digest string, external []string) bool)
	GetFiles(digest string) ([]string, error)
	Err() error
	Equals(state State) bool
	CopyFrom(state State) error
	Check(val validation.Validation, version string, manifestDigests []string, manifestDigestsLower []string) error
	FileChecksum(path string) string
}
