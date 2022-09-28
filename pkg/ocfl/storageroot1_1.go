package ocfl

import (
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension/storageroot"
)

const Version1_1 OCFLVersion = "1.1"

type StorageRootV1_1 struct {
	*StorageRootBase
}

func NewStorageRootV1_1(fs OCFLFS, defaultStorageLayout storageroot.StorageLayout, logger *logging.Logger) (*StorageRootV1_1, error) {
	srb, err := NewStorageRootBase(fs, Version1_1, defaultStorageLayout, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create StorageRootBase Version %s", Version1_1)
	}

	sr := &StorageRootV1_1{StorageRootBase: srb}
	return sr, nil
}
