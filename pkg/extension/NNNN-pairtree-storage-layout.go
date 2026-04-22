package extension

import (
	"encoding/json"
	"fmt"
	"hash"
	"io/fs"

	"math"
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

const StorageLayoutPairTreeName = "NNNN-pairtree-storage-layout"
const StorageLayoutPairTreeDescription = "pairtree-like storage layout"

func init() {
	extension.RegisterExtension(StorageLayoutPairTreeName, NewStorageLayoutPairTree, nil)
}

/*
	https://pythonhosted.org/Pairtree/pairtree.pairtree_client.PairtreeStorageClient-class.html
*/

var rareChars = []rune{'"', '*', '+', 'c', '<', '=', '>', '?', '^', '|'}

var convert = map[rune]rune{
	'/': '=',
	':': '+',
	'.': ',',
}

type StorageLayoutPairTree struct {
	*StorageLayoutPairTreeConfig
	hash   hash.Hash
	logger ocfllogger.OCFLLogger
}

func (sl *StorageLayoutPairTree) WithLogger(logger ocfllogger.OCFLLogger) extensiontypes.Extension {
	sl.logger = logger.With("extension", StorageLayoutPairTreeName)
	return sl
}

func (sl *StorageLayoutPairTree) Load(fsys fs.FS) error {
	data, err := fs.ReadFile(fsys, "config.json")
	if err != nil {
		return errors.Wrap(err, "cannot read config.json")
	}
	if err := json.Unmarshal(data, sl.StorageLayoutPairTreeConfig); err != nil {
		return errors.Wrapf(err, "cannot unmarshal StorageLayoutPairTreeConfig '%s'", string(data))
	}
	if sl.hash, err = checksum.GetHash(checksum.DigestAlgorithm(sl.DigestAlgorithm)); err != nil {
		return errors.Wrapf(err, "hash'%s'not found", sl.DigestAlgorithm)
	}
	return nil
}

func (sl *StorageLayoutPairTree) Terminate() error {
	return nil
}

func (sl *StorageLayoutPairTree) GetConfig() any {
	return sl.StorageLayoutPairTreeConfig
}

func (sl *StorageLayoutPairTree) GetFS() fs.FS {
	return nil
}

func (sl *StorageLayoutPairTree) IsRegistered() bool {
	return false
}

func (sl *StorageLayoutPairTree) WriteLayout(fsys appendfs.FS) error {
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
		Extension:   StorageLayoutFlatDirectName,
		Description: StorageLayoutFlatDirectDescription,
	}); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}

func (sl *StorageLayoutPairTree) SetFS(fsys fs.FS, create bool) {}

type StorageLayoutPairTreeConfig struct {
	*extensiontypes.ExtensionConfig
	UriBase         string `json:"uriBase"`
	StoreDir        string `json:"storeDir"`
	ShortyLength    int    `json:"shortyLength"`
	DigestAlgorithm string `json:"digestAlgorithm"`
}

func NewStorageLayoutPairTree() (extensiontypes.Extension, error) {
	config := &StorageLayoutPairTreeConfig{
		ExtensionConfig: &extensiontypes.ExtensionConfig{ExtensionName: StorageLayoutPairTreeName},
		ShortyLength:    2,
		DigestAlgorithm: string(checksum.DigestSHA512),
	}
	sl := &StorageLayoutPairTree{
		StorageLayoutPairTreeConfig: config,
	}
	return sl, nil
}

func (sl *StorageLayoutPairTree) IsObjectExtension() bool      { return false }
func (sl *StorageLayoutPairTree) IsStorageRootExtension() bool { return true }
func (sl *StorageLayoutPairTree) GetName() string              { return StorageLayoutPairTreeName }

func (sl *StorageLayoutPairTree) SetParams(params map[string]string) error {
	return nil
}

func (sl *StorageLayoutPairTree) WriteConfig(fsys appendfs.FS) error {
	configWriter, err := writefs.Create(fsys, "config.json")
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

func (sl *StorageLayoutPairTree) BuildStorageRootPath(storageRoot storageroot.StorageRoot, id string) (string, error) {
	id = sl.idEncode(id)
	dirparts := []string{}
	numParts := int(math.Ceil(float64(len(id)) / float64(sl.ShortyLength)))
	for i := 0; i < numParts; i++ {
		left := i * sl.ShortyLength
		right := (i + 1) * sl.ShortyLength
		if right >= len(id) {
			right = len(id)
		}
		dirparts = append(dirparts, id[left:right])
	}
	return strings.Join(dirparts, "/"), nil
}

func (sl *StorageLayoutPairTree) idEncode(str string) string {
	var result = []rune{}
	for _, c := range []rune(str) {
		isVisible := 0x21 <= c && c <= 0x7e
		if isVisible {
			for _, rare := range rareChars {
				if c == rare {
					isVisible = false
					break
				}
			}
		}
		if isVisible {
			result = append(result, c)
		} else {
			result = append(result, '^')
			result = append(result, []rune(fmt.Sprintf("%x", c))...)
		}
	}
	str = string(result)
	result = []rune{}
	for _, c := range []rune(str) {
		doConvert := false
		for src, dest := range convert {
			if src == c {
				doConvert = true
				result = append(result, dest)
				break
			}
		}
		if !doConvert {
			result = append(result, c)
		}
	}
	return string(result)
}

// check interface satisfaction
var (
	_ extensiontypes.Extension             = &StorageLayoutPairTree{}
	_ storageroot.ExtensionStorageRootPath = &StorageLayoutPairTree{}
)
