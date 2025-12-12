package extension

import "io/fs"

type ExtensionConfig struct {
	ExtensionName string `json:"extensionName"`
}

type Extension interface {
	GetName() string
	SetFS(fsys fs.FS, create bool)
	GetFS() fs.FS
	SetParams(params map[string]string) error
	WriteConfig() error
	//GetConfigString() string
	GetConfig() any
	IsRegistered() bool
	//	Stat(w io.Writer, statInfo []StatInfo) error
	Terminate() error
}
