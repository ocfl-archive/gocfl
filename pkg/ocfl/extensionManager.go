package ocfl

import "io/fs"

const DefaultExtensionManagerName = "NNNN-gocfl-extension-manager"
const DefaultExtensionInitialName = "initial"

type ExtensionManager interface {
	Extension
	ExtensionStorageRootPath
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
	ExtensionVersionDone
	GetConfig() any
	GetExtensions() []Extension
	Add(ext Extension) error
	Finalize()
	GetConfigName(extName string) (any, error)
	GetFSName(extName string) (fs.FS, error)
	StoreRootLayout(fsys fs.FS) error
	SetInitial(initial ExtensionInitial)
}

type ExtensionManagerConfig struct {
	*ExtensionConfig
	Sort      map[string][]string   `json:"sort"`
	Exclusion map[string][][]string `json:"exclusion"`
}
