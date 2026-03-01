package extension

import (
	"io/fs"

	"github.com/ocfl-archive/gocfl/v2/pkg/appendfs"
)

const DefaultExtensionManagerName = "NNNN-gocfl-extension-manager"
const DefaultExtensionInitialName = "initial"

type CreatorFunc func(sourceFS fs.FS) (Extension, error)

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
	StoreRootLayout(fsys appendfs.FS) error
	SetInitial(initial Initial)
}

type ManagerConfig struct {
	*ExtensionConfig
	Sort      map[string][]string   `json:"sort"`
	Exclusion map[string][][]string `json:"exclusion"`
}
