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
	Finalize() error
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error
	GetID() string
	GetContentDir() string
	GetHead() string
	GetSpec() InventorySpec
	CheckFiles(fileManifest map[checksum.DigestAlgorithm]map[string][]string) error

	DeleteFile(virtualFilename string) error
	//Rename(oldVirtualFilename, newVirtualFilename string) error
	AddFile(virtualFilename string, internalFilename string, checksums map[checksum.DigestAlgorithm]string) error
	RenameFile(dest string, digest string) error

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
	GetDuplicates(checksum string) []string
	AlreadyExists(virtualFilename, checksum string) (bool, error)
	//	IsUpdate(virtualFilename, checksum string) (bool, error)
	Clean() error

	VersionLessOrEqual(v1, v2 string) bool
}

func newInventory(ctx context.Context, object Object, version OCFLVersion, logger *logging.Logger) (Inventory, error) {
	switch version {
	case Version1_1:
		sr, err := newInventoryV1_1(ctx, object, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	default:
		//case Version1_0:
		sr, err := newInventoryV1_0(ctx, object, logger)
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
