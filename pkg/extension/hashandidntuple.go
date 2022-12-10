package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"go.ub.unibas.ch/gocfl/v2/pkg/ocfl"
	"hash"
	"io"
	"strings"
)

const StorageLayoutHashAndIdNTupleName = "0003-hash-and-id-n-tuple-storage-layout"
const StorageLayoutHashAndIdNTupleDescription = "Hashed Truncated N-tuple Trees with Object ID Encapsulating Directory for OCFL Storage Hierarchies"

type StorageLayoutHashAndIdNTuple struct {
	*StorageLayoutHashAndIdNTupleConfig
	hash hash.Hash
	fs   ocfl.OCFLFS
}
type StorageLayoutHashAndIdNTupleConfig struct {
	*ocfl.ExtensionConfig
	DigestAlgorithm string `json:"digestAlgorithm"`
	TupleSize       int    `json:"tupleSize"`
	NumberOfTuples  int    `json:"numberOfTuples"`
}

func NewStorageLayoutHashAndIdNTupleFS(fs ocfl.OCFLFS) (*StorageLayoutHashAndIdNTuple, error) {
	fp, err := fs.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}
	var config = &StorageLayoutHashAndIdNTupleConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewStorageLayoutHashAndIdNTuple(config)
}

func NewStorageLayoutHashAndIdNTuple(config *StorageLayoutHashAndIdNTupleConfig) (*StorageLayoutHashAndIdNTuple, error) {
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
	sl := &StorageLayoutHashAndIdNTuple{StorageLayoutHashAndIdNTupleConfig: config}
	if sl.hash, err = checksum.GetHash(checksum.DigestAlgorithm(config.DigestAlgorithm)); err != nil {
		return nil, errors.Wrapf(err, "invalid hash %s", config.DigestAlgorithm)
	}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.GetName()))
	}

	return sl, nil
}

func (sl *StorageLayoutHashAndIdNTuple) GetName() string {
	return StorageLayoutHashAndIdNTupleName
}

func (sl *StorageLayoutHashAndIdNTuple) SetFS(fs ocfl.OCFLFS) {
	sl.fs = fs
}

func (sl *StorageLayoutHashAndIdNTuple) WriteConfig() error {
	if sl.fs == nil {
		return errors.New("no filesystem set")
	}
	configWriter, err := sl.fs.Create("config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.ExtensionConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
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

func (sl *StorageLayoutHashAndIdNTuple) BuildStorageRootPath(storageRoot ocfl.StorageRoot, id string) (string, error) {
	path := escape(id)
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
		dirparts = append(dirparts, string(digest[i*sl.TupleSize:(i+1)*sl.TupleSize]))
	}
	if len(path) > 100 {
		path = string([]rune(path)[0:100])
		path += "-" + digest
	}
	dirparts = append(dirparts, path)
	return strings.Join(dirparts, "/"), nil
}

func (sl *StorageLayoutHashAndIdNTuple) WriteLayout(fs ocfl.OCFLFS) error {
	configWriter, err := fs.Create("ocfl_layout.json")
	if err != nil {
		return errors.Wrap(err, "cannot open ocfl_layout.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(struct {
		Extension   string `json:"extension"`
		Description string `json:"description"`
	}{
		Extension:   StorageLayoutHashAndIdNTupleName,
		Description: StorageLayoutHashAndIdNTupleDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

// check interface satisfaction
var (
	_ ocfl.Extension                = &StorageLayoutHashAndIdNTuple{}
	_ ocfl.ExtensionStoragerootPath = &StorageLayoutHashAndIdNTuple{}
)
