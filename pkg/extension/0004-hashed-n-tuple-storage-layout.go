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

const StorageLayoutHashedNTupleName = "0004-hashed-n-tuple-storage-layout"
const StorageLayoutHashedNTupleDescription = "Hashed N-tuple Storage Layout"

type StorageLayoutHashedNTuple struct {
	*StorageLayoutHashedNTupleConfig
	hash hash.Hash
	fs   ocfl.OCFLFS
}

type StorageLayoutHashedNTupleConfig struct {
	*ocfl.ExtensionConfig
	DigestAlgorithm string `json:"digestAlgorithm"`
	TupleSize       int    `json:"tupleSize"`
	NumberOfTuples  int    `json:"numberOfTuples"`
	ShortObjectRoot bool   `json:"shortObjectRoot"`
}

func NewStorageLayoutHashedNTupleFS(fsys ocfl.OCFLFSRead) (*StorageLayoutHashedNTuple, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}
	var config = &StorageLayoutHashedNTupleConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewStorageLayoutHashedNTuple(config)
}

func NewStorageLayoutHashedNTuple(config *StorageLayoutHashedNTupleConfig) (*StorageLayoutHashedNTuple, error) {
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
	sl := &StorageLayoutHashedNTuple{StorageLayoutHashedNTupleConfig: config}
	if sl.hash, err = checksum.GetHash(checksum.DigestAlgorithm(config.DigestAlgorithm)); err != nil {
		return nil, errors.Wrapf(err, "invalid hash %s", config.DigestAlgorithm)
	}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.GetName()))
	}

	return sl, nil
}

func (sl *StorageLayoutHashedNTuple) GetName() string { return StorageLayoutHashedNTupleName }

func (sl *StorageLayoutHashedNTuple) GetConfigString() string {
	str, _ := json.MarshalIndent(sl.StorageLayoutHashedNTupleConfig, "", "  ")
	return string(str)
}

func (sl *StorageLayoutHashedNTuple) SetFS(fs ocfl.OCFLFS) {
	sl.fs = fs
}

func (sl *StorageLayoutHashedNTuple) SetParams(params map[string]string) error {
	return nil
}

func (sl *StorageLayoutHashedNTuple) WriteConfig() error {
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

func (sl *StorageLayoutHashedNTuple) BuildStorageRootPath(storageRoot ocfl.StorageRoot, id string) (string, error) {
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

func (sl *StorageLayoutHashedNTuple) WriteLayout(fs ocfl.OCFLFS) error {
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
		Extension:   StorageLayoutHashedNTupleName,
		Description: StorageLayoutHashedNTupleDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

// check interface satisfaction
var (
	_ ocfl.Extension                = &StorageLayoutHashedNTuple{}
	_ ocfl.ExtensionStorageRootPath = &StorageLayoutHashedNTuple{}
)
