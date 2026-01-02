package types

import (
	"github.com/je4/utils/v2/pkg/checksum"
)

type StateFileCallback func(internal []string, external []string, digest string) error

type Inventory interface {
	Finalize(inCreation bool) error
	IsEqual(i2 Inventory) bool
	Init(id string, digest checksum.DigestAlgorithm, fixity []checksum.DigestAlgorithm) error
	GetID() string
	WithContentDir(contentDir string) Inventory
	GetContentDir() string
	GetRealContentDir() string
	GetHead() *VersionNumber
	GetSpec() InventorySpec
	CheckFiles(fileManifest map[checksum.DigestAlgorithm]map[string][]string) error

	//	DeleteFile(stateFilename string) error
	//	RenameFile(stateSource, stateDest string) error
	//Rename(oldVirtualFilename, newVirtualFilename string) error
	AddFile(stateFilenames []string, manifestFilename string, checksums map[checksum.DigestAlgorithm]string) error
	//	CopyFile(dest string, digest string) error

	IterateFiles(version *VersionNumber, fn StateFileCallback) error
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
	//NewVersion(msg, UserName, UserAddress string) error
	//GetDuplicates(checksum string) []string
	AlreadyExists(stateFilename, checksum string) (bool, error)
	//	IsUpdate(virtualFilename, checksum string) (bool, error)
	Clean() error

	//	EchoDelete(existing []string, pathprefix string) error
}
