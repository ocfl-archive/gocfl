package ocfl

import (
	"io"
)

type ExtensionConfig struct {
	ExtensionName string `json:"extensionName"`
}

type Extension interface {
	GetName() string
	IsObjectExtension() bool
	IsStoragerootExtension() bool
	WriteConfig(configWriter io.Writer) error
}

type StoragerootPath interface {
	WriteLayout(layoutwriter io.Writer) error
	BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)
}

type ObjectContentPath interface {
	BuildObjectContentPath(object Object, originalPath string) (string, error)
}

type ContentChange interface {
	AddFile(object Object, source, dest string)
	UpdateFile(object Object, source, dest string)
	DeleteFile(object Object, dest string)
}
