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
	//SetFS(fsys fs.FS, create bool)
	//GetFS() fs.FS
	SetParams(params map[string]string) error
	WriteConfig(fsys streamfs.FS) error
	//GetConfigString() string
	GetConfig() any
	IsRegistered() bool
	//	Stat(w io.Writer, statInfo []StatInfo) error
	Terminate() error
}
