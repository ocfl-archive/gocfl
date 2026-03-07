package extension

import (
	"io/fs"

	"github.com/ocfl-archive/gocfl/v2/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfllogger"
)

type ExtensionConfig struct {
	ExtensionName string `json:"extensionName"`
}

type Extension interface {
	WithLogger(logger ocfllogger.OCFLLogger) Extension
	GetName() string
	Load(fsys fs.FS) error
	SetParams(params map[string]string) error
	WriteConfig(fsys appendfs.FS) error
	GetConfig() any
	IsRegistered() bool
	Terminate() error
}
