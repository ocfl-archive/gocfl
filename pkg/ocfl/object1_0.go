package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/op/go-logging"
)

type ObjectV1_0 struct {
	*ObjectBase
}

func newObjectV1_0(ctx context.Context, fs OCFLFS, storageRoot StorageRoot, logger *logging.Logger) (*ObjectV1_0, error) {
	ob, err := newObjectBase(ctx, fs, Version1_0, storageRoot, logger)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obv10 := &ObjectV1_0{ObjectBase: ob}
	return obv10, nil
}
