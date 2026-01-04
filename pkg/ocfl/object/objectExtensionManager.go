package object

import (
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension/types"
)

type ExtensionManager interface {
	extensiontypes.ExtensionManagerCore
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
