package object

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
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
	extension.Extension
	BuildObjectManifestPath(originalPath string, area string) (string, error)
}

var ExtensionObjectExtractPathWrongAreaError = fmt.Errorf("invalid area")

type ExtensionObjectExtractPath interface {
	extension.Extension
	BuildObjectExtractPath(originalPath string, area string) (string, error)
}

type ExtensionObjectStatePath interface {
	extension.Extension
	BuildObjectStatePath(originalPath string, area string) (string, error)
}

type ExtensionArea interface {
	extension.Extension
	GetAreaPath(area string) (string, error)
}

type ExtensionStream interface {
	extension.Extension
	StreamObject(object VersionWriter, reader io.Reader, stateFiles []string, dest string) error
}

type ExtensionContentChange interface {
	extension.Extension
	AddFileBefore(object VersionWriter, sourceFS fs.FS, source string, dest string, area string, isDir bool) error
	UpdateFileBefore(object VersionWriter, sourceFS fs.FS, source, dest, area string, isDir bool) error
	DeleteFileBefore(versionWriter VersionWriter, dest string, area string) error
	AddFileAfter(versionWriter VersionWriter, sourceFS fs.FS, source []string, internalPath, digest, area string, isDir bool) error
	UpdateFileAfter(object VersionWriter, sourceFS fs.FS, source, area string, isDir bool) error
	DeleteFileAfter(object VersionWriter, dest string, area string) error
}

type ExtensionObjectChange interface {
	extension.Extension
	UpdateObjectBefore(object VersionWriter) error
	UpdateObjectAfter(object VersionWriter) error
}

type ExtensionFixityDigest interface {
	extension.Extension
	GetFixityDigests() []checksum.DigestAlgorithm
}

type ExtensionMetadata interface {
	extension.Extension
	GetMetadata(object Object) (map[string]any, error)
}

type ExtensionVersionDone interface {
	extension.Extension
	VersionDone(object Object) error
}

type ExtensionNewVersion interface {
	extension.Extension
	NeedNewVersion(object VersionWriter) (bool, error)
	DoNewVersion(object VersionWriter) error
}
