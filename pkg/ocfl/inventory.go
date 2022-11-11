package ocfl

import (
	"context"
	"emperror.dev/errors"
	"encoding/json"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"golang.org/x/exp/slices"
)

type Inventory interface {
	Init() error
	AfterLoad()
	GetID() string
	GetContentDir() string
	GetHead() string
	GetSpec() InventorySpec
	CheckFiles(fileManifest map[checksum.DigestAlgorithm]map[string][]string) error

	DeleteFile(virtualFilename string) error
	Rename(oldVirtualFilename, newVirtualFilename string) error
	AddFile(virtualFilename string, realFilename string, checksum string) error

	//GetContentDirectory() string
	GetVersionStrings() []string
	GetVersions() map[string]*Version
	GetFiles() map[string][]string
	GetManifest() map[string][]string
	GetFixity() map[checksum.DigestAlgorithm]map[string][]string
	GetFilesFlat() []string
	GetDigestAlgorithm() checksum.DigestAlgorithm
	GetFixityDigestAlgorithm() []checksum.DigestAlgorithm
	IsWriteable() bool
	//	IsModified() bool
	BuildRealname(virtualFilename string) string
	NewVersion(msg, UserName, UserAddress string) error
	IsDuplicate(checksum string) bool
	AlreadyExists(virtualFilename, checksum string) (bool, error)
	//	IsUpdate(virtualFilename, checksum string) (bool, error)
	Clean() error

	VersionLessOrEqual(v1, v2 string) bool
}

func NewInventory(ctx context.Context, object Object, id string, version OCFLVersion, digest checksum.DigestAlgorithm, logger *logging.Logger) (Inventory, error) {
	switch version {
	case Version1_1:
		sr, err := NewInventoryV1_1(ctx, object, id, digest, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	default:
		//case Version1_0:
		sr, err := NewInventoryV1_0(ctx, object, id, digest, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
		//		return nil, errors.New(fmt.Sprintf("Inventory Version %s not supported", version))
	}
}

func LoadInventory(ctx context.Context, object Object, data []byte, logger *logging.Logger) (Inventory, error) {
	anyMap := map[string]any{}
	if err := json.Unmarshal(data, &anyMap); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal json '%s'", string(data))
	}
	var version OCFLVersion
	t, ok := anyMap["type"]
	if !ok {
		return nil, errors.New("no type in inventory")
	}
	sStr, ok := t.(string)
	if !ok {
		return nil, errors.Errorf("type not a string in inventory - '%v'", t)
	}
	switch sStr {
	case "https://ocfl.io/1.1/spec/#inventory":
		version = Version1_1
	case "https://ocfl.io/1.0/spec/#inventory":
		version = Version1_0
	default:
		// if we don't know anything use the old stuff
		version = Version1_0
	}
	inventory, err := NewInventory(ctx, object, "", version, object.GetDigest(), logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create empty inventory")
	}
	if err := json.Unmarshal(data, inventory); err != nil {
		// now lets try it again
		jsonMap := map[string]any{}
		// check for json format error
		if err2 := json.Unmarshal(data, &jsonMap); err2 != nil {
			addValidationErrors(ctx, GetValidationError(version, E033).AppendDescription("json syntax error: %v", err2))
			addValidationErrors(ctx, GetValidationError(version, E034).AppendDescription("json syntax error: %v", err2))
		} else {
			if _, ok := jsonMap["head"].(string); !ok {
				addValidationErrors(ctx, GetValidationError(version, E040).AppendDescription("head is not of string type: %v", jsonMap["head"]))
			}
		}
		//return nil, errors.Wrapf(err, "cannot marshal data - %s", string(data))
	}
	inventory.AfterLoad()
	return inventory, nil
}

func InventoryIsEqual(i1, i2 Inventory) bool {
	data1, err := json.Marshal(i1)
	if err != nil {
		return false
	}

	data2, err := json.Marshal(i2)
	if err != nil {
		return false
	}
	return slices.Equal(data1, data2)
}
