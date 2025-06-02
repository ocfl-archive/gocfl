package ocfl

import (
	"context"
	"encoding/json"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
	"golang.org/x/exp/slices"
)

type StateFileCallback func(internal []string, external []string, digest string) error

type Inventory interface {
	Finalize(inCreation bool) error
	IsEqual(i2 Inventory) bool
	Contains(i2 Inventory) bool
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error
	GetID() string
	GetContentDir() string
	GetRealContentDir() string
	GetHead() string
	GetSpec() InventorySpec
	CheckFiles(fileManifest map[checksum.DigestAlgorithm]map[string][]string, filenamePrefix string) error

	DeleteFile(stateFilename string) error
	RenameFile(stateSource, stateDest string) error
	//Rename(oldVirtualFilename, newVirtualFilename string) error
	AddFile(stateFilenames []string, manifestFilename string, checksums map[checksum.DigestAlgorithm]string) error
	CopyFile(dest string, digest string) error

	IterateStateFiles(version string, fn StateFileCallback) error
	GetStateFiles(version string, cs string) ([]string, error)

	//GetContentDirectory() string
	GetVersionStrings() []string
	GetVersions() map[string]*VersionData
	GetFiles() map[string][]string
	GetManifest() map[string][]string
	GetFixity() Fixity
	GetFilesFlat(filenamePrefix string) []string
	GetDigestAlgorithm() checksum.DigestAlgorithm
	GetFixityDigestAlgorithm() []checksum.DigestAlgorithm
	IsWriteable() bool
	IsModified() bool
	BuildManifestName(stateFilename string) string
	BuildManifestNameVersion(stateFilename string, version string) string
	NewVersion(msg, UserName, UserAddress string) error
	GetDuplicates(checksum string) []string
	AlreadyExists(stateFilename, checksum string) (bool, error)
	//	IsUpdate(virtualFilename, checksum string) (bool, error)
	Clean() error

	VersionLessOrEqual(v1, v2 string) bool
	echoDelete(existing []string, pathprefix string) error
}

func newInventory(
	ctx context.Context,
	folder string,
	version OCFLVersion,
	logger zLogger.ZLogger,
	errorFactory *archiveerror.Factory,
) (Inventory, error) {
	switch version {
	case Version1_1:
		sr, err := newInventoryV1_1(ctx, folder, version, logger, errorFactory)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	case Version2_0:
		sr, err := newInventoryV2_0(ctx, folder, version, logger, errorFactory)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	default:
		//case Version1_0:
		sr, err := newInventoryV1_0(ctx, folder, version, logger, errorFactory)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
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
