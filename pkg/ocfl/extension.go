package ocfl

import (
	"fmt"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
)

type ExtensionConfig struct {
	ExtensionName string `json:"extensionName"`
}

type Extension interface {
	GetName() string
	SetFS(fs OCFLFS)
	SetParams(params map[string]string) error
	WriteConfig() error
	GetConfigString() string
	//	Stat(w io.Writer, statInfo []StatInfo) error
}

const (
	ExtensionStorageRootPathName    = "StorageRootPath"
	ExtensionObjectContentPathName  = "ObjectContentPath"
	ExtensionObjectExtractPathName  = "ObjectExtractPath"
	ExtensionObjectExternalPathName = "ObjectExternalPath"
	ExtensionContentChangeName      = "ContentChange"
	ExtensionObjectChangeName       = "ObjectChange"
	ExtensionFixityDigestName       = "FixityDigest"
)

type ExtensionStorageRootPath interface {
	Extension
	WriteLayout(fs OCFLFS) error
	BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)
}

type ExtensionObjectContentPath interface {
	Extension
	BuildObjectContentPath(object Object, originalPath string, area string) (string, error)
}

var ExtensionObjectExtractPathWrongAreaError = fmt.Errorf("invalid area")

type ExtensionObjectExtractPath interface {
	Extension
	BuildObjectExtractPath(object Object, originalPath string) (string, error)
}

type ExtensionObjectExternalPath interface {
	Extension
	BuildObjectExternalPath(object Object, originalPath string) (string, error)
}

type ExtensionContentChange interface {
	Extension
	AddFileBefore(object Object, source, dest string) error
	UpdateFileBefore(object Object, source, dest string) error
	DeleteFileBefore(object Object, dest string) error
	AddFileAfter(object Object, source, dest string) error
	UpdateFileAfter(object Object, source, dest string) error
	DeleteFileAfter(object Object, dest string) error
}

type ExtensionObjectChange interface {
	Extension
	UpdateObjectBefore(object Object) error
	UpdateObjectAfter(object Object) error
}

type ExtensionFixityDigest interface {
	Extension
	GetFixityDigests() []checksum.DigestAlgorithm
}
