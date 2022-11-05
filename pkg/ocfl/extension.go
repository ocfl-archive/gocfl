package ocfl

import (
	"emperror.dev/errors"
	"encoding/json"
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

func NewDefaultStorageRootExtension() (Extension, error) {
	var err error
	var cfg = &StorageLayoutDirectCleanConfig{
		ExtensionConfig:             &ExtensionConfig{ExtensionName: StorageLayoutDirectCleanName},
		MaxPathnameLen:              32000,
		MaxFilenameLen:              127,
		WhitespaceReplacementString: " ",
		ReplacementString:           "_",
		UTFEncode:                   true,
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot marshal config %v", cfg)
	}
	layout, err := NewStorageLayoutDirectClean(data)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot initialize %s", cfg.ExtensionName)
	}
	return layout, nil
}
