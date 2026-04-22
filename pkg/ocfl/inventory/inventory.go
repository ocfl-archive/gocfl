package inventory

import (
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/version"
)

type StateFileCallback func(internal []string, external []string, digest string) error

type Inventory interface {
	Finalize(inCreation bool) error
	Equals(i2 Inventory) bool
	//Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error
	WithID(id string) Inventory
	GetID() string
	WithDigestAlgorithm(algorithm checksum.DigestAlgorithm) Inventory
	GetDigestAlgorithm() checksum.DigestAlgorithm
	WithContentDir(contentDir string) Inventory
	GetContentDir() string
	GetRealContentDir() string
	GetHead() *VersionNumber
	GetSpec() InventorySpec
	CheckFiles(fileManifest map[checksum.DigestAlgorithm]map[string][]string) error
	Bytes() (inventory []byte, checksumString string, err error)
	GetOCFLVersion() version.OCFLVersion

	//	DeleteFile(stateFilename string) error
	//	RenameFile(stateSource, stateDest string) error
	//Rename(oldVirtualFilename, newVirtualFilename string) error
	AddFile(stateFilenames []string, manifestFilename string, checksums map[checksum.DigestAlgorithm]string) error
	//	CopyFile(dest string, digest string) error

	IterateFiles(version *VersionNumber, fn StateFileCallback) error
	//	GetStateFiles(version *VersionNumber, cs string) ([]string, error)

	//GetContentDirectory() string
	//GetVersionNumbers() []*VersionNumber
	WithVersions(versions Versions) Inventory
	GetVersions() Versions
	//GetFiles() map[*VersionNumber][]string
	WithManifest(manifest Manifest) Inventory
	GetManifest() Manifest
	WithFixity(fixity Fixity) Inventory
	GetFixity() Fixity
	//GetFixityDigestAlgorithm() iter.Seq[checksum.DigestAlgorithm]
	WithWriteable() Inventory
	IsWriteable() bool
	IsModified() bool
	BuildManifestName(stateFilename string) string
	BuildManifestNameVersion(stateFilename string, version *VersionNumber) string
	//NewVersion(msg, UserName, UserAddress string) error
	//GetDuplicates(checksum string) []string
	AlreadyExists(stateFilename, checksum string) (bool, error)
	//	IsUpdate(virtualFilename, checksum string) (bool, error)
	Clean() error
	//	EchoDelete(existing []string, pathprefix string) error
}
