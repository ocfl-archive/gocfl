package ocfl

import "go.ub.unibas.ch/gocfl/v2/pkg/checksum"

type ExtensionExternalParam struct {
	Param       string
	File        string
	Description string
}

type ExtensionConfig struct {
	ExtensionName string `json:"extensionName"`
}

type Extension interface {
	GetName() string
	SetFS(fs OCFLFS)
	WriteConfig() error
}
type ExtensionStoragerootPath interface {
	WriteLayout(fs OCFLFS) error
	BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)
}

type ExtensionObjectContentPath interface {
	BuildObjectContentPath(object Object, originalPath string) (string, error)
}

type ExtensionContentChange interface {
	AddFileBefore(object Object, source, dest string) error
	UpdateFileBefore(object Object, source, dest string) error
	DeleteFileBefore(object Object, dest string) error
	AddFileAfter(object Object, source, dest string) error
	UpdateFileAfter(object Object, source, dest string) error
	DeleteFileAfter(object Object, dest string) error
}

type ExtensionObjectChange interface {
	UpdateBefore(object Object) error
	UpdateAfter(object Object) error
}

type ExtensionFixityDigest interface {
	GetDigests() []checksum.DigestAlgorithm
}
