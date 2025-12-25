package inventory

import (
	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

var DigestNotFound = errors.New("Digest not found")

type State interface {
	String() string
	IterateFiles() func(yield func(digest string, external []string) bool)
	GetFiles(digest string) ([]string, error)
	Err() error
	Equals(state State) bool
	CopyFrom(state State) State
	Check(val validation.Validator, version string, manifestDigests []string, manifestDigestsLower []string) error
}
