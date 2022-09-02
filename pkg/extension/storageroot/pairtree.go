package storageroot

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"hash"
	"io"
	"math"
	"strings"
)

const StorageLayoutPairTreeName = "gocfl-pairtree"

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
}

type StorageLayoutPairTreeConfig struct {
	*Config
	UriBase         string `json:"uriBase"`
	StoreDir        string `json:"storeDir"`
	ShortyLength    int    `json:"shortyLength"`
	DigestAlgorithm string `json:"digestAlgorithm"`
}

func NewStorageLayoutPairTree(config *StorageLayoutPairTreeConfig) (*StorageLayoutPairTree, error) {
	sl := &StorageLayoutPairTree{StorageLayoutPairTreeConfig: config}
	var err error
	if sl.hash, err = checksum.GetHash(checksum.DigestAlgorithm(config.DigestAlgorithm)); err != nil {
		return nil, errors.Wrapf(err, "hash %s not found", config.DigestAlgorithm)
	}
	if config.ExtensionName != sl.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, sl.Name()))
	}

	return sl, nil
}

func (sl *StorageLayoutPairTree) Name() string {
	return StorageLayoutPairTreeName
}

func (sl *StorageLayoutPairTree) ExecuteID(id string) (string, error) {
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

func (sl *StorageLayoutPairTree) WriteConfig(configWriter io.Writer) error {
	jenc := json.NewEncoder(configWriter)
	jenc.SetIndent("", "   ")
	if err := jenc.Encode(sl.Config); err != nil {
		return errors.Wrapf(err, "cannot encode config to file")
	}
	return nil
}
