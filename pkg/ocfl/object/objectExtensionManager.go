package object

import "github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"

type ExtensionManager interface {
	extension.ExtensionManager
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
