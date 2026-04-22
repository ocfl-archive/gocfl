// Package object contains the interfaces and implementations for OCFL objects.
package object

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"
)

const (
	// ExtensionStorageRootPathName is the name of the StorageRootPath extension.
	ExtensionStorageRootPathName = "StorageRootPath"
	// ExtensionObjectContentPathName is the name of the ObjectContentPath extension.
	ExtensionObjectContentPathName = "ObjectContentPath"
	// ExtensionObjectExtractPathName is the name of the ObjectExtractPath extension.
	ExtensionObjectExtractPathName = "ObjectExtractPath"
	// ExtensionObjectExternalPathName is the name of the ObjectExternalPath extension.
	ExtensionObjectExternalPathName = "ObjectExternalPath"
	// ExtensionContentChangeName is the name of the ContentChange extension.
	ExtensionContentChangeName = "ContentChange"
	// ExtensionObjectChangeName is the name of the ObjectChange extension.
	ExtensionObjectChangeName = "ObjectChange"
	// ExtensionFixityDigestName is the name of the FixityDigest extension.
	ExtensionFixityDigestName = "FixityDigest"
	// ExtensionMetadataName is the name of the Metadata extension.
	ExtensionMetadataName = "Metadata"
	// ExtensionAreaName is the name of the Area extension.
	ExtensionAreaName = "Area"
	// ExtensionStreamName is the name of the Stream extension.
	ExtensionStreamName = "Stream"
	// ExtensionNewVersionName is the name of the NewVersion extension.
	ExtensionNewVersionName = "NewVersion"
	// ExtensionVersionDoneName is the name of the VersionDone extension.
	ExtensionVersionDoneName = "VersionDone"
	// ExtensionInitialName is the name of the Initial extension.
	ExtensionInitialName = "Initial"
)

// ExtensionObjectContentPath is an interface for extensions that can build object manifest paths.
type ExtensionObjectContentPath interface {
	extension.Extension
	// BuildObjectManifestPath builds an object manifest path from an original path and an area.
	BuildObjectManifestPath(originalPath string, area string) (string, error)
}

// ExtensionObjectExtractPathWrongAreaError is returned when an invalid area is provided to an ObjectExtractPath extension.
var ExtensionObjectExtractPathWrongAreaError = fmt.Errorf("invalid area")

// ExtensionObjectExtractPath is an interface for extensions that can build object extract paths.
type ExtensionObjectExtractPath interface {
	extension.Extension
	// BuildObjectExtractPath builds an object extract path from an original path and an area.
	BuildObjectExtractPath(originalPath string, area string) (string, error)
}

// ExtensionObjectStatePath is an interface for extensions that can build object state paths.
type ExtensionObjectStatePath interface {
	extension.Extension
	// BuildObjectStatePath builds an object state path from an original path and an area.
	BuildObjectStatePath(originalPath string, area string) (string, error)
}

// ExtensionArea is an interface for extensions that can provide area paths.
type ExtensionArea interface {
	extension.Extension
	// GetAreaPath returns the path for a given area.
	GetAreaPath(area string) (string, error)
}

// ExtensionStream is an interface for extensions that can stream objects.
type ExtensionStream interface {
	extension.Extension
	// StreamObject streams an object's content.
	StreamObject(object VersionWriter, reader io.Reader, stateFiles []string, dest string) error
}

// ExtensionContentChange is an interface for extensions that are notified of content changes.
type ExtensionContentChange interface {
	extension.Extension
	// AddFileBefore is called before a file is added to an object.
	AddFileBefore(object VersionWriter, sourceFS fs.FS, source string, dest string, area string, isDir bool) error
	// UpdateFileBefore is called before a file is updated in an object.
	UpdateFileBefore(object VersionWriter, sourceFS fs.FS, source, dest, area string, isDir bool) error
	// DeleteFileBefore is called before a file is deleted from an object.
	DeleteFileBefore(versionWriter VersionWriter, dest string, area string) error
	// AddFileAfter is called after a file is added to an object.
	AddFileAfter(versionWriter VersionWriter, sourceFS fs.FS, source []string, internalPath, digest, area string, isDir bool) error
	// UpdateFileAfter is called after a file is updated in an object.
	UpdateFileAfter(object VersionWriter, sourceFS fs.FS, source, area string, isDir bool) error
	// DeleteFileAfter is called after a file is deleted from an object.
	DeleteFileAfter(object VersionWriter, dest string, area string) error
}

// ExtensionObjectChange is an interface for extensions that are notified of object changes.
type ExtensionObjectChange interface {
	extension.Extension
	// UpdateObjectBefore is called before an object is updated.
	UpdateObjectBefore(object VersionWriter) error
	// UpdateObjectAfter is called after an object is updated.
	UpdateObjectAfter(object VersionWriter) error
}

// ExtensionFixityDigest is an interface for extensions that provide fixity digests.
type ExtensionFixityDigest interface {
	extension.Extension
	// GetFixityDigests returns the fixity digests provided by the extension.
	GetFixityDigests() []checksum.DigestAlgorithm
}

// ExtensionMetadata is an interface for extensions that provide object metadata.
type ExtensionMetadata interface {
	extension.Extension
	// GetMetadata returns the metadata for an object.
	GetMetadata(sourceFS fs.FS, obj Object) (map[string]any, error)
}

// ExtensionVersionDone is an interface for extensions that are notified when a version is done.
type ExtensionVersionDone interface {
	extension.Extension
	// VersionDone is called when a version is done.
	VersionDone(object Object) error
}

// ExtensionNewVersion is an interface for extensions that can determine if a new version is needed and perform actions for a new version.
type ExtensionNewVersion interface {
	extension.Extension
	// NeedNewVersion returns true if a new version is needed.
	NeedNewVersion(object VersionWriter) (bool, error)
	// DoNewVersion performs actions for a new version.
	DoNewVersion(object VersionWriter) error
}
