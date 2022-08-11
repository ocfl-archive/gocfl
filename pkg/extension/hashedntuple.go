package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"hash"
	"io"
	"strings"
)

const HashedNTupleName = "0004-hashed-n-tuple-storage-layout"

type HashedNTuple struct {
	*HashedNTupleConfig
	hash hash.Hash
}
type HashedNTupleConfig struct {
	*Config
	DigestAlgorithm string `json:"digestAlgorithm"`
	TupleSize       int    `json:"tupleSize"`
	NumberOfTuples  int    `json:"numberOfTuples"`
	ShortObjectRoot bool   `json:"shortObjectRoot"`
}

func NewHashedNTuple(config *HashedNTupleConfig) (*HashedNTuple, error) {
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
	sl := &HashedNTuple{HashedNTupleConfig: config}
	if sl.hash, err = checksum.GetHash(checksum.DigestAlgorithm(config.DigestAlgorithm)); err != nil {
		return nil, errors.Wrapf(err, "invalid hash %s", config.DigestAlgorithm)
	}
	if config.ExtensionName != sl.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.Name()))
	}

	return sl, nil
}

func (sl *HashedNTuple) Name() string {
	return HashedNTupleName
}

func (sl *HashedNTuple) ID2Path(id string) (string, error) {
	sl.hash.Reset()
	if _, err := sl.hash.Write([]byte(id)); err != nil {
		return "", errors.Wrapf(err, "cannot hash %s", id)
	}
	digestBytes := sl.hash.Sum(nil)
	digest := fmt.Sprintf("%x", digestBytes)
	if len(digest) < sl.TupleSize*sl.NumberOfTuples {
		return "", errors.New(fmt.Sprintf("digest %s to short for %v tuples of %v chars", sl.DigestAlgorithm, sl.NumberOfTuples, sl.TupleSize))
	}
	dirparts := []string{}
	for i := 0; i < sl.NumberOfTuples; i++ {
		dirparts = append(dirparts, digest[i*sl.TupleSize:(i+1)*sl.TupleSize])
	}
	if sl.ShortObjectRoot {
		dirparts = append(dirparts, digest[sl.NumberOfTuples*sl.TupleSize:])
	} else {
		dirparts = append(dirparts, digest)
	}
	return strings.Join(dirparts, "/"), nil
}

func (sl *HashedNTuple) WriteConfig(configWriter io.Writer) error {
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.Config); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}
