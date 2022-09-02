package ocfl

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension/object"
	"io"
)

type Object interface {
	LoadInventory() (Inventory, error)
	StoreInventory() error
	StoreExtensions() error
	New(id string, path object.Path) error
	Load() error
	StartUpdate(msg string, UserName string, UserAddress string) error
	AddFile(virtualFilename string, reader io.Reader, digest string) error
	GetID() string
	Check() error
	Close() error
}

func NewObject(fs OCFLFS, pathPrefix, version string, id string, logger *logging.Logger) (Object, error) {
	switch version {
	case "1.0":
		o, err := NewObjectV1_0(fs, pathPrefix, id, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return o, nil
	default:
		return nil, errors.New(fmt.Sprintf("Object Version %s not supported", version))
	}
}
