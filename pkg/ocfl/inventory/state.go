package inventory

import (
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
)

var DigestNotFound = errors.New("Digest not found")
var StateTypeDifferent = errors.New("State type different")

type State interface {
	Operations
	AddFile(stateFilename string, digest string) (bool, error)

	String() string
	Iterate() func(yield func(digest string, external []string) bool)
	GetFiles(digest string) ([]string, error)
	Err() error
	Equals(state State) bool
	CopyFrom(state State) error
	Check(version *VersionNumber, manifestDigests []string, manifestDigestsLower []string) error
	FileChecksum(path string) string
	WithDigestAlgorithm(dgst checksum.DigestAlgorithm) State
}
