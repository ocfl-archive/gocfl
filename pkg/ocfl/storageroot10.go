package ocfl

import (
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension"
)

type StorageRootV10 struct {
	*StorageRootBase
}

func NewStorageRootV10(fs OCFLFS, defaultStorageLayout extension.StorageLayout, logger *logging.Logger) (*StorageRootV10, error) {
	srb, err := NewStorageRootBase(fs, "1.0", defaultStorageLayout, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create StorageRootBase Version 1.0")
	}

	sr := &StorageRootV10{StorageRootBase: srb}
	return sr, nil
}
