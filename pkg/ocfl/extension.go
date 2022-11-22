package ocfl

import "go.ub.unibas.ch/gocfl/v2/pkg/checksum"

type ExtensionConfig struct {
	ExtensionName string `json:"extensionName"`
}

type Extension interface {
	GetName() string
	WriteConfig(fs OCFLFS) error
}

type StoragerootPath interface {
	WriteLayout(fs OCFLFS) error
	BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)
}

type ObjectContentPath interface {
	BuildObjectContentPath(object Object, originalPath string) (string, error)
}

type ContentChange interface {
	AddFileBefore(object Object, source, dest string) error
	UpdateFileBefore(object Object, source, dest string) error
	DeleteFileBefore(object Object, dest string) error
	AddFileAfter(object Object, source, dest string) error
	UpdateFileAfter(object Object, source, dest string) error
	DeleteFileAfter(object Object, dest string) error
}

type FixityDigest interface {
	GetDigests() []checksum.DigestAlgorithm
}
