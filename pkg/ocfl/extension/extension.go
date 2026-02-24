package extension

import (
	"io/fs"

	"github.com/ocfl-archive/gocfl/v2/pkg/streamfs"
)

type ExtensionConfig struct {
	ExtensionName string `json:"extensionName"`
}

type Extension interface {
	GetName() string
	Load(fsys fs.FS) error
	SetParams(params map[string]string) error
	WriteConfig(fsys streamfs.FS) error
	GetConfig() any
	IsRegistered() bool
	Terminate() error
}
