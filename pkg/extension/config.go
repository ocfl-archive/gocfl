package extension

import "emperror.dev/errors"

var ErrNotSupported = errors.New("extension not supported")

type Config struct {
	ExtensionName string `json:"extensionName"`
}
