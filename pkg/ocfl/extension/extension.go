package extension

import (
	"encoding/json"

	"github.com/ocfl-archive/gocfl/v3/pkg/appendfs"
	"github.com/ocfl-archive/gocfl/v3/pkg/ocfllogger"
)

type ExtensionConfig struct {
	ExtensionName string `json:"extensionName"`
}

type Extension interface {
	WithLogger(logger ocfllogger.OCFLLogger) Extension
	GetName() string
	GetDescription() string
	GetDocumentation() string
	Load(data json.RawMessage) error
	SetParams(params map[string]string) error
	WriteConfig(fsys appendfs.FS) error
	GetConfig() any
	IsRegistered() bool
	Terminate() error
}
