package object

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/types"
)

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
	ExtensionVersionDoneName        = "VersionDone"
	ExtensionInitialName            = "Initial"
)

type ExtensionObjectContentPath interface {
	extensiontypes.Extension
	BuildObjectManifestPath(object Object, originalPath string, area string) (string, error)
}

var ExtensionObjectExtractPathWrongAreaError = fmt.Errorf("invalid area")

type ExtensionObjectExtractPath interface {
	extensiontypes.Extension
	BuildObjectExtractPath(object Object, originalPath string, area string) (string, error)
}

type ExtensionObjectStatePath interface {
	extensiontypes.Extension
	BuildObjectStatePath(object Object, originalPath string, area string) (string, error)
}

type ExtensionArea interface {
	extensiontypes.Extension
	GetAreaPath(object Object, area string) (string, error)
}

type ExtensionStream interface {
	extensiontypes.Extension
	StreamObject(object Object, reader io.Reader, stateFiles []string, dest string) error
}

type ExtensionContentChange interface {
	extensiontypes.Extension
	AddFileBefore(object Object, sourceFS fs.FS, source string, dest string, area string, isDir bool) error
	UpdateFileBefore(object Object, sourceFS fs.FS, source, dest, area string, isDir bool) error
	DeleteFileBefore(object Object, dest string, area string) error
	AddFileAfter(object Object, sourceFS fs.FS, source []string, internalPath, digest, area string, isDir bool) error
	UpdateFileAfter(object Object, sourceFS fs.FS, source, dest, area string, isDir bool) error
	DeleteFileAfter(object Object, dest string, area string) error
}

type ExtensionObjectChange interface {
	extensiontypes.Extension
	UpdateObjectBefore(object Object) error
	UpdateObjectAfter(object Object) error
}

type ExtensionFixityDigest interface {
	extensiontypes.Extension
	GetFixityDigests() []checksum.DigestAlgorithm
}

type ExtensionMetadata interface {
	extensiontypes.Extension
	GetMetadata(object Object) (map[string]any, error)
}

type ExtensionVersionDone interface {
	extensiontypes.Extension
	VersionDone(object Object) error
}

type ExtensionNewVersion interface {
	extensiontypes.Extension
	NeedNewVersion(object Object) (bool, error)
	DoNewVersion(object Object) error
}
