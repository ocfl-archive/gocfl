package extension

import (
	"io/fs"
)

const DefaultExtensionManagerName = "NNNN-gocfl-extension-manager"
const DefaultExtensionInitialName = "initial"

type CreatorFunc func() Extension

type Initial interface {
	Extension
	GetExtension() string
	SetExtension(ext string)
}

type ManagerCore interface {
	Extension
	GetConfig() any
	GetExtensions() []Extension
	Add(ext Extension) error
	Finalize()
	GetConfigName(extName string) (any, error)
	//GetFSName(extName string) (fs.FS, error)
	StoreRootLayout(fsys streamfs.FS) error
	SetInitial(initial Initial)
}

type ManagerConfig struct {
	*ExtensionConfig
	Sort      map[string][]string   `json:"sort"`
	Exclusion map[string][][]string `json:"exclusion"`
}
