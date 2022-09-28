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
	DeleteFile(virtualFilename string, reader io.Reader, digest string) error
	GetID() string
	Check() error
	Close() error
}

func NewObject(fsys OCFLFS, pathPrefix string, version OCFLVersion, id string, logger *logging.Logger) (Object, error) {
	objectFS := fsys.SubFS(pathPrefix)
	switch version {
	case Version1_0:
		o, err := NewObjectV1_0(objectFS, id, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return o, nil
	case Version1_1:
		o, err := NewObjectV1_1(objectFS, id, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return o, nil
	default:
		return nil, errors.New(fmt.Sprintf("Object Version %s not supported", version))
	}
}
