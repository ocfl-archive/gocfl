package types

type ExtensionManager interface {
	ExtensionManagerCore
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
