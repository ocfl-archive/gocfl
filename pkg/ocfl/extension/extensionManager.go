package extension

import (
	"io/fs"
)

const DefaultExtensionManagerName = "NNNN-gocfl-extension-manager"
const DefaultExtensionInitialName = "initial"

type ExtensionInitial interface {
	Extension
	GetExtension() string
	SetExtension(ext string)
}

type ExtensionManager interface {
	Extension
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
