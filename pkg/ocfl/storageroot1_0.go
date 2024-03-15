package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"io/fs"
)

const Version1_0 OCFLVersion = "1.0"

type StorageRootV1_0 struct {
	*StorageRootBase
}

func NewStorageRootV1_0(ctx context.Context, fsys fs.FS, extensionFactory *ExtensionFactory, logger zLogger.ZWrapper) (*StorageRootV1_0, error) {
	srb, err := NewStorageRootBase(ctx, fsys, Version1_0, extensionFactory, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create StorageRootBase Version %s", Version1_0)
	}

	sr := &StorageRootV1_0{StorageRootBase: srb}
	return sr, nil
}
