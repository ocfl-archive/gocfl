package extension

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/je4/filesystem/v2/pkg/writefs"
	"github.com/je4/gocfl/v2/pkg/ocfl"
	"github.com/je4/utils/v2/pkg/checksum"
	"hash"
	"io"
	"io/fs"
	"math"
	"strings"
)

const StorageLayoutPairTreeName = "NNNN-pairtree-storage-layout"

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
	hash hash.Hash
	fsys fs.FS
}

func (sl *StorageLayoutPairTree) IsRegistered() bool {
	return false
}

func (sl *StorageLayoutPairTree) WriteLayout(fsys fs.FS) error {
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

func (sl *StorageLayoutPairTree) SetFS(fsys fs.FS) {
	sl.fsys = fsys
}

type StorageLayoutPairTreeConfig struct {
	*ocfl.ExtensionConfig
	UriBase         string `json:"uriBase"`
	StoreDir        string `json:"storeDir"`
	ShortyLength    int    `json:"shortyLength"`
	DigestAlgorithm string `json:"digestAlgorithm"`
}

func NewStorageLayoutPairTreeFS(fsys fs.FS) (*StorageLayoutPairTree, error) {
	fp, err := fsys.Open("config.json")
	if err != nil {
		return nil, errors.Wrap(err, "cannot open config.json")
	}
	defer fp.Close()
	data, err := io.ReadAll(fp)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config.json")
	}
	var config = &StorageLayoutPairTreeConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal DirectCleanConfig '%s'", string(data))
	}
	return NewStorageLayoutPairTree(config)
}

func NewStorageLayoutPairTree(config *StorageLayoutPairTreeConfig) (*StorageLayoutPairTree, error) {
	sl := &StorageLayoutPairTree{StorageLayoutPairTreeConfig: config}
	var err error
	if sl.hash, err = checksum.GetHash(checksum.DigestAlgorithm(config.DigestAlgorithm)); err != nil {
		return nil, errors.Wrapf(err, "hash'%s'not found", config.DigestAlgorithm)
	}
	if config.ExtensionName != sl.GetName() {
		return nil, errors.New(fmt.Sprintf("invalid extension name'%s'for extension %s", config.ExtensionName, sl.GetName()))
	}

	return sl, nil
}

func (sl *StorageLayoutPairTree) IsObjectExtension() bool      { return false }
func (sl *StorageLayoutPairTree) IsStorageRootExtension() bool { return true }
func (sl *StorageLayoutPairTree) GetName() string              { return StorageLayoutPairTreeName }

func (sl *StorageLayoutPairTree) GetConfigString() string {
	str, _ := json.MarshalIndent(sl.StorageLayoutPairTreeConfig, "", "  ")
	return string(str)
}

func (sl *StorageLayoutPairTree) SetParams(params map[string]string) error {
	return nil
}

func (sl *StorageLayoutPairTree) WriteConfig() error {
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

func (sl *StorageLayoutPairTree) BuildStorageRootPath(storageRoot ocfl.StorageRoot, id string) (string, error) {
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
	_ ocfl.Extension                = &StorageLayoutPairTree{}
	_ ocfl.ExtensionStorageRootPath = &StorageLayoutPairTree{}
)
