package object

import "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"

// ExtensionManager combines all extension interfaces relevant for OCFL objects.
type ExtensionManager interface {
	extension.ManagerCore[ExtensionManager]
	ExtensionObjectContentPath
	ExtensionObjectStatePath
	ExtensionContentChange
	ExtensionObjectChange
	ExtensionFixityDigest
	ExtensionObjectExtractPath
	ExtensionMetadata
	ExtensionArea
	ExtensionStream
	ExtensionNewVersion
}
