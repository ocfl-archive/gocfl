package inventory

import (
	"iter"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
)

var DigestAlgNotFound = errors.New("digest algorithm not found")
var FixityTypeDifferent = errors.New("fixity type different")

type Fixity interface {
	AddFile(manifestFilename string, digests map[checksum.DigestAlgorithm]string) (bool, error)
	WithAlgorithms(algorithms ...checksum.DigestAlgorithm) Fixity
	String() string
	Iterate(alg checksum.DigestAlgorithm) func(yield func(digest string, external []string) bool)
	GetDigestAlgorithms() iter.Seq[checksum.DigestAlgorithm]
	GetFiles(alg checksum.DigestAlgorithm, digest string) ([]string, error)
	Err() error
	Equals(fixity Fixity) bool
	CopyFrom(fixity Fixity) error
	Check(fileManifest map[checksum.DigestAlgorithm]map[string][]string) error
	Finalize(inCreation bool) error
	Checksums(s string) map[checksum.DigestAlgorithm]string
}
