package storagelayout

import (
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"gitlab.switch.ch/ub-unibas/gocfl/v2/pkg/checksum"
	"hash"
	"strings"
)

type HashedNTuple struct {
	digestAlgorithm checksum.DigestAlgorithm
	tupleSize       int
	numberOfTuples  int
	hash            hash.Hash
	shortObjectRoot bool
}
type HashedNTupleStorage struct {
	Config
	DigestAlgorithm string `json:"digestAlgorithm"`
	TupleSize       int    `json:"tupleSize"`
	NumberOfTuples  int    `json:"numberOfTuples"`
	ShortObjectRoot bool   `json:"shortObjectRoot"`
}

func NewHashedNTuple(config *HashedNTupleStorage) (*HashedNTuple, error) {
	var err error
	if config.NumberOfTuples > 32 {
		config.NumberOfTuples = 32
	}
	if config.TupleSize > 32 {
		config.TupleSize = 32
	}
	if config.TupleSize == 0 || config.NumberOfTuples == 0 {
		config.NumberOfTuples = 0
		config.TupleSize = 0
	}
	sl := &HashedNTuple{
		digestAlgorithm: checksum.DigestAlgorithm(config.DigestAlgorithm),
		tupleSize:       config.TupleSize,
		numberOfTuples:  config.NumberOfTuples,
		shortObjectRoot: config.ShortObjectRoot,
	}
	if sl.hash, err = checksum.GetHash(checksum.DigestAlgorithm(config.DigestAlgorithm)); err != nil {
		return nil, emperror.Wrapf(err, "invalid hash %s", config.DigestAlgorithm)
	}
	if config.ExtensionName != sl.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.Name()))
	}

	return sl, nil
}

func (sl *HashedNTuple) Name() string {
	return "0004-hashed-n-tuple-storage-layout"
}

func (sl *HashedNTuple) ID2Path(id string) (string, error) {
	sl.hash.Reset()
	if _, err := sl.hash.Write([]byte(id)); err != nil {
		return "", emperror.Wrapf(err, "cannot hash %s", id)
	}
	digestBytes := sl.hash.Sum(nil)
	digest := fmt.Sprintf("%x", digestBytes)
	if len(digest) < sl.tupleSize*sl.numberOfTuples {
		return "", errors.New(fmt.Sprintf("digest %s to short for %v tuples of %v chars", sl.digestAlgorithm, sl.numberOfTuples, sl.tupleSize))
	}
	dirparts := []string{}
	for i := 0; i < sl.numberOfTuples; i++ {
		dirparts = append(dirparts, digest[i*sl.tupleSize:(i+1)*sl.tupleSize])
	}
	if sl.shortObjectRoot {
		dirparts = append(dirparts, digest[sl.numberOfTuples*sl.tupleSize:])
	} else {
		dirparts = append(dirparts, digest)
	}
	return strings.Join(dirparts, "/"), nil
}
