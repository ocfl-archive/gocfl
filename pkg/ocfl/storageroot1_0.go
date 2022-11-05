package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/op/go-logging"
)

const Version1_0 OCFLVersion = "1.0"

type StorageRootV1_0 struct {
	*StorageRootBase
}

func NewStorageRootV1_0(ctx context.Context, fs OCFLFS, defaultStorageLayout Extension, extensionFactory *ExtensionFactory, logger *logging.Logger) (*StorageRootV1_0, error) {
	srb, err := NewStorageRootBase(ctx, fs, Version1_0, defaultStorageLayout, extensionFactory, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create StorageRootBase Version %s", Version1_0)
	}

	sr := &StorageRootV1_0{StorageRootBase: srb}
	return sr, nil
}
