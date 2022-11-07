package ocfl

import (
	"io"
)

type ExtensionConfig struct {
	ExtensionName string `json:"extensionName"`
}

type Extension interface {
	GetName() string
	WriteConfig(configWriter io.Writer) error
}

type StoragerootPath interface {
	BuildStorageRootPath(storageRoot StorageRoot, id string) (string, error)
}

type ObjectContentPath interface {
	BuildObjectContentPath(object Object, originalPath string) (string, error)
}
