package extension

import "io"

const MAX_DIR_LEN = 255

type StorageLayout interface {
	ExecutePath(id string) (string, error)
	Name() string
	WriteConfig(config io.Writer) error
}

type Config struct {
	ExtensionName string `json:"extensionName"`
}
