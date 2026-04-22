package extension

import (
	"encoding/json"
	"fmt"
	"hash"
	"io/fs"
	"strings"

	"emperror.dev/errors"
	"github.com/je4/filesystem/v3/pkg/writefs"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	extensiontypes "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

const StorageLayoutHashedNTupleName = "0004-hashed-n-tuple-storage-layout"
const StorageLayoutHashedNTupleDescription = "Hashed N-tuple Storage Layout"

func init() {
	extension.RegisterExtension(StorageLayoutHashedNTupleName, NewStorageLayoutHashedNTuple, nil)
}

func NewStorageLayoutHashedNTuple() (extensiontypes.Extension, error) {
	config := &StorageLayoutHashedNTupleConfig{
		ExtensionConfig: &extensiontypes.ExtensionConfig{ExtensionName: StorageLayoutHashedNTupleName},
		DigestAlgorithm: string(checksum.DigestSHA512),
		TupleSize:       0,
		NumberOfTuples:  0,
		ShortObjectRoot: false,
	}
	sl := &StorageLayoutHashedNTuple{StorageLayoutHashedNTupleConfig: config}
	var err error
	if sl.hash, err = checksum.GetHash(checksum.DigestAlgorithm(config.DigestAlgorithm)); err != nil {
		return nil, errors.Wrapf(err, "cannot get hash for %s", config.DigestAlgorithm)
	}
	return sl, nil
}

func (sl *StorageLayoutHashedNTuple) WithLogger(logger ocfllogger.OCFLLogger) extensiontypes.Extension {
	sl.logger = logger.With("extension", StorageLayoutHashedNTupleName)
	return sl
}

func (sl *StorageLayoutHashedNTuple) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}
	if err := json.Unmarshal(data, sl.StorageLayoutHashedNTupleConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal StorageLayoutHashedNTupleConfig '%s'", string(data))
	}
	if sl.NumberOfTuples > 32 {
		sl.NumberOfTuples = 32
	}
	if sl.TupleSize > 32 {
		sl.TupleSize = 32
	}
	if sl.TupleSize == 0 || sl.NumberOfTuples == 0 {
		sl.NumberOfTuples = 0
		sl.TupleSize = 0
	}
	if h, err2 := checksum.GetHash(checksum.DigestAlgorithm(sl.DigestAlgorithm)); err2 != nil {
		return errors.Wrapf(err2, "invalid hash %s", sl.DigestAlgorithm)
	} else {
		sl.hash = h
	}
	return nil
}

type StorageLayoutHashedNTupleConfig struct {
	*extensiontypes.ExtensionConfig
	DigestAlgorithm string `json:"digestAlgorithm"`
	TupleSize       int    `json:"tupleSize"`
	NumberOfTuples  int    `json:"numberOfTuples"`
	ShortObjectRoot bool   `json:"shortObjectRoot"`
}

type StorageLayoutHashedNTuple struct {
	*StorageLayoutHashedNTupleConfig
	hash   hash.Hash
	logger ocfllogger.OCFLLogger
}

func (sl *StorageLayoutHashedNTuple) Terminate() error {
	return nil
}

func (sl *StorageLayoutHashedNTuple) GetConfig() any {
	return sl.StorageLayoutHashedNTupleConfig
}

func (sl *StorageLayoutHashedNTuple) IsRegistered() bool {
	return true
}

func (sl *StorageLayoutHashedNTuple) GetName() string { return StorageLayoutHashedNTupleName }

func (sl *StorageLayoutHashedNTuple) SetParams(params map[string]string) error {
	return nil
}

func (sl *StorageLayoutHashedNTuple) WriteConfig(fsys appendfs.FS) error {
	configWriter, err := writefs.Create(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot open config.json")
	}
	defer configWriter.Close()
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.StorageLayoutHashedNTupleConfig); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *StorageLayoutHashedNTuple) BuildStorageRootPath(storageRoot storageroot.StorageRoot, id string) (string, error) {
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

func (sl *StorageLayoutHashedNTuple) WriteLayout(fsys appendfs.FS) error {
	configWriter, err := writefs.Create(fsys, "ocfl_layout.json")
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
	_ extensiontypes.Extension             = &StorageLayoutHashedNTuple{}
	_ storageroot.ExtensionStorageRootPath = &StorageLayoutHashedNTuple{}
)
