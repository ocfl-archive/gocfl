package inventory

import (
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/validation"
)

var DigestAlgNotFound = errors.New("digest algorithm not found")

func NewFixityBase() *FixityBase {
	return &FixityBase{
		fixity: map[checksum.DigestAlgorithm]map[string][]string{},
		err: nil
	}

}

type FixityBase struct {
	fixity map[checksum.DigestAlgorithm]map[string][]string
	err error
}

func (f *FixityBase) IterateFiles(alg checksum.DigestAlgorithm) func(yield func(digest string, external []string) bool) {
	return func(yield func(digest string, internal []string) bool {
		dfiles, ok := f.fixity[alg]
		if !ok {
			return
		}
		for digest, files := range dfiles {
			if !yield(digest, files) {
				return
			}
		}
		return
	})
}

func (f *FixityBase) GetFiles(alg checksum.DigestAlgorithm, digest string) ([]string, error) {
	dfiles, ok := f.fixity[alg]
	if !ok {
		return nil, errors.Wrapf(DigestAlgNotFound, "digest algorithm '%s'", alg)
	}
	files, ok := dfiles[digest]
	if !ok {
		return nil, errors.Wrapf(DigestNotFound, "digest  '%s-%s'", alg, digest)
	}
	return files, nil
}

func (f *FixityBase) Err() error {
	//TODO implement me
	panic("implement me")
}

func (f *FixityBase) Equals(fixity Fixity) bool {
	//TODO implement me
	panic("implement me")
}

func (f *FixityBase) CopyFrom(fixity Fixity) State {
	//TODO implement me
	panic("implement me")
}

func (f *FixityBase) Check(val validation.Validator, version string, manifestDigests []string, manifestDigestsLower []string) error {
	//TODO implement me
	panic("implement me")
}

func (f *FixityBase) String() string {
	//TODO implement me
	panic("implement me")
}

var _ Fixity = (*FixityBase)(nil)
