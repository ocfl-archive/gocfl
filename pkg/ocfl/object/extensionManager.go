package object

import "github.com/ocfl-archive/gocfl/v3/pkg/ocfl/extension"

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
