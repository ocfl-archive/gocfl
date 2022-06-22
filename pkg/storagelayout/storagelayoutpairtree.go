package storagelayout

import (
	"errors"
	"fmt"
	"github.com/goph/emperror"
	"gitlab.switch.ch/ub-unibas/gocfl/v2/pkg/checksum"
	"hash"
	"math"
	"strings"
)

/*
	https://pythonhosted.org/Pairtree/pairtree.pairtree_client.PairtreeStorageClient-class.html
*/

var rareChars = []rune{'"', '*', '+', 'c', '<', '=', '>', '?', '^', '|'}

var convert = map[rune]rune{
	'/': '=',
	':': '+',
	'.': ',',
}

type PairTreeStorageLayout struct {
	uriBase         string
	storeDir        string
	shortyLength    int
	digestAlgorithm checksum.DigestAlgorithm
	hash            hash.Hash
}

type PairTreeStorageLayoutConfig struct {
	Config
	UriBase         string `json:"uriBase"`
	StoreDir        string `json:"storeDir"`
	ShortyLength    int    `json:"shortyLength"`
	DigestAlgorithm string `json:"digestAlgorithm"`
}

func NewPairTreeStorageLayout(config *PairTreeStorageLayoutConfig) (*PairTreeStorageLayout, error) {
	ptsl := &PairTreeStorageLayout{
		uriBase:         config.UriBase,
		storeDir:        config.StoreDir,
		shortyLength:    config.ShortyLength,
		digestAlgorithm: checksum.DigestAlgorithm(config.DigestAlgorithm),
	}
	var err error
	if ptsl.hash, err = checksum.GetHash(checksum.DigestAlgorithm(config.DigestAlgorithm)); err != nil {
		return nil, emperror.Wrapf(err, "hash %s not found", config.DigestAlgorithm)
	}
	if config.ExtensionName != ptsl.Name() {
		return nil, errors.New(fmt.Sprintf("invalid extension name %s for extension %s", config.ExtensionName, ptsl.Name()))
	}

	return ptsl, nil
}

func (ptsl *PairTreeStorageLayout) Name() string {
	return "gocfl-pairtree"
}

func (ptsl *PairTreeStorageLayout) ID2Path(id string) (string, error) {
	id = ptsl.idEncode(id)
	dirparts := []string{}
	numParts := int(math.Ceil(float64(len(id)) / float64(ptsl.shortyLength)))
	for i := 0; i < numParts; i++ {
		left := i * ptsl.shortyLength
		right := (i + 1) * ptsl.shortyLength
		if right >= len(id) {
			right = len(id)
		}
		dirparts = append(dirparts, id[left:right])
	}
	return strings.Join(dirparts, "/"), nil
}

func (ptsl *PairTreeStorageLayout) idEncode(str string) string {
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
