package ocfl

import (
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension/storageroot"
)

type StorageRootV1_0 struct {
	*StorageRootBase
}

func NewStorageRootV1_0(fs OCFLFS, defaultStorageLayout storageroot.StorageLayout, logger *logging.Logger) (*StorageRootV1_0, error) {
	srb, err := NewStorageRootBase(fs, "1.0", defaultStorageLayout, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create StorageRootBase Version 1.0")
	}

	sr := &StorageRootV1_0{StorageRootBase: srb}
	return sr, nil
}
