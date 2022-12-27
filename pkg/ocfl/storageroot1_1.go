package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/op/go-logging"
)

const Version1_1 OCFLVersion = "1.1"

type StorageRootV1_1 struct {
	*StorageRootBase
}

func NewStorageRootV1_1(ctx context.Context, fs OCFLFSRead, extensionFactory *ExtensionFactory, logger *logging.Logger) (*StorageRootV1_1, error) {
	srb, err := NewStorageRootBase(ctx, fs, Version1_1, extensionFactory, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create StorageRootBase Version %s", Version1_1)
	}

	sr := &StorageRootV1_1{StorageRootBase: srb}
	return sr, nil
}
