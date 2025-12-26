package inventory

import (
	"iter"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

var DigestAlgNotFound = errors.New("digest algorithm not found")
var FixityTypeDifferent = errors.New("fixity type different")

type Fixity interface {
	String() string
	IterateFiles(alg checksum.DigestAlgorithm) func(yield func(digest string, external []string) bool)
	GetDigestAlgorithms() iter.Seq[checksum.DigestAlgorithm]
	GetFiles(alg checksum.DigestAlgorithm, digest string) ([]string, error)
	Err() error
	Equals(val validation.Validation, fixity Fixity) bool
	CopyFrom(fixity Fixity) error
	Check(val validation.Validation, version version.OCFLVersion, fileManifest map[checksum.DigestAlgorithm]map[string][]string) error
	Finalize(inCreation bool) error
}
