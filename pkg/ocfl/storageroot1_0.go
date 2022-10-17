package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/extension/storageroot"
)

const Version1_0 OCFLVersion = "1.0"

type StorageRootV1_0 struct {
	*StorageRootBase
}

func NewStorageRootV1_0(ctx context.Context, fs OCFLFS, defaultStorageLayout storageroot.StorageLayout, logger *logging.Logger) (*StorageRootV1_0, error) {
	srb, err := NewStorageRootBase(ctx, fs, Version1_0, defaultStorageLayout, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create StorageRootBase Version %s", Version1_0)
	}

	sr := &StorageRootV1_0{StorageRootBase: srb}
	return sr, nil
}
