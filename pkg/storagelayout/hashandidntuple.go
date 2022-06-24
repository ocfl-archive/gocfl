package storagelayout

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"gitlab.switch.ch/ub-unibas/gocfl/v2/pkg/checksum"
	"hash"
	"io"
	"strings"
)

const HashAndIdNTupleName = "0003-hash-and-id-n-tuple-storage-layout"

type HashAndIdNTuple struct {
	*HashAndIdNTupleConfig
	hash hash.Hash
}
type HashAndIdNTupleConfig struct {
	*Config
	DigestAlgorithm string `json:"digestAlgorithm"`
	TupleSize       int    `json:"tupleSize"`
	NumberOfTuples  int    `json:"numberOfTuples"`
}

func NewHashAndIdNTuple(config *HashAndIdNTupleConfig) (*HashAndIdNTuple, error) {
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
	sl := &HashAndIdNTuple{HashAndIdNTupleConfig: config}
	if sl.hash, err = checksum.GetHash(checksum.DigestAlgorithm(config.DigestAlgorithm)); err != nil {
		return nil, emperror.Wrapf(err, "invalid hash %s", config.DigestAlgorithm)
	}
	if config.ExtensionName != sl.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.Name()))
	}

	return sl, nil
}

func (sl *HashAndIdNTuple) Name() string {
	return HashAndIdNTupleName
}

func shouldEscape(c rune) bool {
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || c == '-' || c == '_' {
		return false
	}
	// Everything else must be escaped.
	return true
}

func escape(str string) string {
	var result = []byte{}
	for _, c := range []byte(str) {
		if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || c == '-' || c == '_' {
			result = append(result, c)
			continue
		}
		result = append(result, '%')
		result = append(result, fmt.Sprintf("%x", c)...)
	}
	return string(result)
}

func (sl *HashAndIdNTuple) ID2Path(id string) (string, error) {
	path := escape(id)
	sl.hash.Reset()
	if _, err := sl.hash.Write([]byte(id)); err != nil {
		return "", emperror.Wrapf(err, "cannot hash %s", id)
	}
	digestBytes := sl.hash.Sum(nil)
	digest := fmt.Sprintf("%x", digestBytes)
	if len(digest) < sl.TupleSize*sl.NumberOfTuples {
		return "", errors.New(fmt.Sprintf("digest %s to short for %v tuples of %v chars", sl.DigestAlgorithm, sl.NumberOfTuples, sl.TupleSize))
	}
	dirparts := []string{}
	for i := 0; i < sl.NumberOfTuples; i++ {
		dirparts = append(dirparts, string(digest[i*sl.TupleSize:(i+1)*sl.TupleSize]))
	}
	if len(path) > 100 {
		path = string([]rune(path)[0:100])
		path += "-" + digest
	}
	dirparts = append(dirparts, path)
	return strings.Join(dirparts, "/"), nil
}

func (sl *HashAndIdNTuple) WriteConfig(configWriter io.Writer) error {
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.Config); err != nil {
		return emperror.Wrapf(err, "cannot encode config to file")
	}
	return nil
}
