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
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/storageroot"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
)

const StorageLayoutHashAndIdNTupleName = "0003-hash-and-id-n-tuple-storage-layout"
const StorageLayoutHashAndIdNTupleDescription = "Hashed Truncated N-tuple Trees with Object ID Encapsulating Directory for OCFL Storage Hierarchies"

func NewStorageLayoutHashAndIdNTuple(logger ocfllogger.OCFLLogger) (*StorageLayoutHashAndIdNTuple, error) {
	config := &StorageLayoutHashAndIdNTupleConfig{
		ExtensionConfig: &extension.ExtensionConfig{ExtensionName: StorageLayoutHashAndIdNTupleName},
		DigestAlgorithm: string(checksum.DigestSHA512),
		TupleSize:       0,
		NumberOfTuples:  0,
	}
	sl := &StorageLayoutHashAndIdNTuple{StorageLayoutHashAndIdNTupleConfig: config, logger: logger.With("extension", StorageLayoutHashAndIdNTupleName)}
	var err error
	if sl.hash, err = checksum.GetHash(checksum.DigestAlgorithm(config.DigestAlgorithm)); err != nil {
		return nil, errors.Wrapf(err, "invalid hash %s", config.DigestAlgorithm)
	}
	return sl, nil
}

func (sl *StorageLayoutHashAndIdNTuple) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}
	if err := json.Unmarshal(data, sl.StorageLayoutHashAndIdNTupleConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal StorageLayoutHashAndIdNTupleConfig '%s'", string(data))
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

type StorageLayoutHashAndIdNTupleConfig struct {
	*extension.ExtensionConfig
	DigestAlgorithm string `json:"digestAlgorithm"`
	TupleSize       int    `json:"tupleSize"`
	NumberOfTuples  int    `json:"numberOfTuples"`
}

type StorageLayoutHashAndIdNTuple struct {
	*StorageLayoutHashAndIdNTupleConfig
	hash   hash.Hash
	fsys   fs.FS
	logger ocfllogger.OCFLLogger
}

func (sl *StorageLayoutHashAndIdNTuple) Terminate() error {
	return nil
}

func (sl *StorageLayoutHashAndIdNTuple) GetFS() fs.FS {
	return sl.fsys
}

func (sl *StorageLayoutHashAndIdNTuple) GetConfig() any {
	return sl.StorageLayoutHashAndIdNTupleConfig
}

func (sl *StorageLayoutHashAndIdNTuple) IsRegistered() bool {
	return true
}

func (sl *StorageLayoutHashAndIdNTuple) GetName() string {
	return StorageLayoutHashAndIdNTupleName
}

func (sl *StorageLayoutHashAndIdNTuple) SetFS(fsys fs.FS, create bool) {
	sl.fsys = fsys
}

func (sl *StorageLayoutHashAndIdNTuple) SetParams(params map[string]string) error {
	return nil
}

func (sl *StorageLayoutHashAndIdNTuple) WriteConfig(streamfs.FS) error {
	if sl.fsys == nil {
		return errors.New("no filesystem set")
	}
	configWriter, err := writefs.Create(sl.fsys, "config.json")
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

func (sl *StorageLayoutHashAndIdNTuple) BuildStorageRootPath(storageRoot storageroot.StorageRoot, id string) (string, error) {
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

func (sl *StorageLayoutHashAndIdNTuple) WriteLayout(fsys fs.FS) error {
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
		Extension:   StorageLayoutHashAndIdNTupleName,
		Description: StorageLayoutHashAndIdNTupleDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

// check interface satisfaction
var (
	_ extension.Extension                  = &StorageLayoutHashAndIdNTuple{}
	_ storageroot.ExtensionStorageRootPath = &StorageLayoutHashAndIdNTuple{}
)
