package extension

import "emperror.dev/errors"

const errorTopic string = "gocfl"

var ErrNotSupported = errors.New("extension not supported")

type Config struct {
	ExtensionName string `json:"extensionName"`
}
