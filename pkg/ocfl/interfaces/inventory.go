package interfaces

import (
	"context"
	"encoding/json"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/inventory"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
	"golang.org/x/exp/slices"
)

type StateFileCallback func(internal []string, external []string, digest string) error

type Inventory interface {
	Finalize(inCreation bool) error
	IsEqual(i2 Inventory) bool
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error
	GetID() string
	GetContentDir() string
	GetRealContentDir() string
	GetHead() *VersionNumber
	GetSpec() inventory.InventorySpec
	CheckFiles(fileManifest map[checksum.DigestAlgorithm]map[string][]string) error

	//	DeleteFile(stateFilename string) error
	//	RenameFile(stateSource, stateDest string) error
	//Rename(oldVirtualFilename, newVirtualFilename string) error
	AddFile(stateFilenames []string, manifestFilename string, checksums map[checksum.DigestAlgorithm]string) error
	//	CopyFile(dest string, digest string) error

	IterateStateFiles(version *VersionNumber, fn StateFileCallback) error
	//	GetStateFiles(version *VersionNumber, cs string) ([]string, error)

	//GetContentDirectory() string
	//GetVersionNumbers() []*VersionNumber
	GetVersions() Versions
	//GetFiles() map[*VersionNumber][]string
	GetManifest() Manifest
	GetFixity() Fixity
	GetDigestAlgorithm() checksum.DigestAlgorithm
	//GetFixityDigestAlgorithm() iter.Seq[checksum.DigestAlgorithm]
	IsWriteable() bool
	IsModified() bool
	BuildManifestName(stateFilename string) string
	BuildManifestNameVersion(stateFilename string, version *VersionNumber) string
	NewVersion(msg, UserName, UserAddress string) error
	//GetDuplicates(checksum string) []string
	AlreadyExists(stateFilename, checksum string) (bool, error)
	//	IsUpdate(virtualFilename, checksum string) (bool, error)
	Clean() error

	//	EchoDelete(existing []string, pathprefix string) error
}
