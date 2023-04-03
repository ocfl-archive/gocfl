package ocfl

import (
	"context"
	"emperror.dev/errors"
	"encoding/json"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/op/go-logging"
	"golang.org/x/exp/slices"
)

type Inventory interface {
	Finalize(inCreation bool) error
	IsEqual(i2 Inventory) bool
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error
	GetID() string
	GetContentDir() string
	GetRealContentDir() string
	GetHead() string
	GetSpec() InventorySpec
	CheckFiles(fileManifest map[checksum.DigestAlgorithm]map[string][]string) error

	DeleteFile(stateFilename string) error
	//Rename(oldVirtualFilename, newVirtualFilename string) error
	AddFile(stateFilenames []string, manifestFilename string, checksums map[checksum.DigestAlgorithm]string) error
	CopyFile(dest string, digest string) error

	IterateStateFiles(version string, fn func(internal, external, digest string) error) error
	GetStateFiles(version string, cs string) ([]string, error)

	//GetContentDirectory() string
	GetVersionStrings() []string
	GetVersions() map[string]*Version
	GetFiles() map[string][]string
	GetManifest() map[string][]string
	GetFixity() Fixity
	GetFilesFlat() []string
	GetDigestAlgorithm() checksum.DigestAlgorithm
	GetFixityDigestAlgorithm() []checksum.DigestAlgorithm
	IsWriteable() bool
	IsModified() bool
	BuildManifestName(stateFilename string) string
	NewVersion(msg, UserName, UserAddress string) error
	GetDuplicates(checksum string) []string
	AlreadyExists(stateFilename, checksum string) (bool, error)
	//	IsUpdate(virtualFilename, checksum string) (bool, error)
	Clean() error

	VersionLessOrEqual(v1, v2 string) bool
	echoDelete(existing []string, pathprefix string) error
}

func newInventory(ctx context.Context, object Object, folder string, version OCFLVersion, logger *logging.Logger) (Inventory, error) {
	switch version {
	case Version1_1:
		sr, err := newInventoryV1_1(ctx, object, folder, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	default:
		//case Version1_0:
		sr, err := newInventoryV1_0(ctx, object, folder, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
		//		return nil, errors.Finalize(fmt.Sprintf("Inventory Version %s not supported", version))
	}
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
