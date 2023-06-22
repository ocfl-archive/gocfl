package ocfl

import (
	"fmt"
	"github.com/je4/utils/v2/pkg/checksum"
	"io"
	"io/fs"
)

type ExtensionConfig struct {
	ExtensionName string `json:"extensionName"`
}

type Extension interface {
	GetName() string
	SetFS(fsys fs.FS)
	GetFS() fs.FS
	SetParams(params map[string]string) error
	WriteConfig() error
	//GetConfigString() string
	GetConfig() any
	IsRegistered() bool
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
	ExtensionMetadataName           = "Metadata"
	ExtensionAreaName               = "Area"
	ExtensionStreamName             = "Stream"
	ExtensionNewVersionName         = "NewVersion"
)

type ExtensionStorageRootPath interface {
	Extension
	WriteLayout(fsys fs.FS) error
	BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)
}

type ExtensionObjectContentPath interface {
	Extension
	BuildObjectManifestPath(object Object, originalPath string, area string) (string, error)
}

var ExtensionObjectExtractPathWrongAreaError = fmt.Errorf("invalid area")

type ExtensionObjectExtractPath interface {
	Extension
	BuildObjectExtractPath(object Object, originalPath string, area string) (string, error)
}

type ExtensionObjectStatePath interface {
	Extension
	BuildObjectStatePath(object Object, originalPath string, area string) (string, error)
}

type ExtensionArea interface {
	Extension
	GetAreaPath(object Object, area string) (string, error)
}

type ExtensionStream interface {
	Extension
	StreamObject(object Object, reader io.Reader, stateFiles []string, dest string) error
}

type ExtensionContentChange interface {
	Extension
	AddFileBefore(object Object, sourceFS fs.FS, source string, dest string, area string, isDir bool) error
	UpdateFileBefore(object Object, sourceFS fs.FS, source, dest, area string, isDir bool) error
	DeleteFileBefore(object Object, dest string, area string) error
	AddFileAfter(object Object, sourceFS fs.FS, source []string, internalPath, digest, area string, isDir bool) error
	UpdateFileAfter(object Object, sourceFS fs.FS, source, dest, area string, isDir bool) error
	DeleteFileAfter(object Object, dest string, area string) error
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

type ExtensionMetadata interface {
	Extension
	GetMetadata(object Object) (map[string]any, error)
}

type ExtensionNewVersion interface {
	Extension
	NeedNewVersion(object Object) (bool, error)
	DoNewVersion(object Object) error
}
