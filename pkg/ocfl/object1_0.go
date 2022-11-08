package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/op/go-logging"
)

type ObjectV1_0 struct {
	*ObjectBase
}

func NewObjectV1_0(ctx context.Context, fs OCFLFS, id string, storageroot StorageRoot, logger *logging.Logger) (*ObjectV1_0, error) {
	ob, err := NewObjectBase(ctx, fs, Version1_0, id, storageroot, logger)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obv10 := &ObjectV1_0{ObjectBase: ob}
	return obv10, nil
}
