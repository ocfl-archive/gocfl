package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/op/go-logging"
)

const ObjectV11Version = "1.1"

type ObjectV1_1 struct {
	*ObjectBase
}

func NewObjectV1_1(ctx context.Context, fs OCFLFS, id string, storageroot StorageRoot, logger *logging.Logger) (*ObjectV1_1, error) {
	ob, err := NewObjectBase(ctx, fs, Version1_1, id, storageroot, logger)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obv11 := &ObjectV1_1{ObjectBase: ob}
	return obv11, nil
}
