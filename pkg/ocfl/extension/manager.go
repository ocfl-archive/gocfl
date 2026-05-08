package extension

import (
	"encoding/json"
	"io/fs"

	"emperror.dev/errors"
	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
)

const DefaultExtensionManagerName = "NNNN-gocfl-extension-manager"
const DefaultExtensionInitialName = "initial"

var ExtensionManagerTypeAssertionError = errors.New("cannot convert manager to type")

type CreatorFunc func(data json.RawMessage, extFS fs.FS) (Extension, error)

type Initial interface {
	Extension
	GetExtension() string
	SetExtension(ext string)
}

type ManagerCore[T any] interface {
	Extension
	GetConfig() any
	GetExtensions() []Extension
	Add(ext Extension) error
	Finalize()
	GetConfigName(extName string) (any, error)
	StoreRootLayout(fsys appendfs.FS) error
	SetInitial(initial Initial)
}

type ManagerConfig struct {
	*ExtensionConfig
	Sort      map[string][]string   `json:"sort"`
	Exclusion map[string][][]string `json:"exclusion"`
}
