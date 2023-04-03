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

func newObjectV1_1(ctx context.Context, fs OCFLFSRead, storageRoot StorageRoot, logger *logging.Logger) (*ObjectV1_1, error) {
	ob, err := newObjectBase(ctx, fs, Version1_1, storageRoot, logger)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obv11 := &ObjectV1_1{ObjectBase: ob}
	return obv11, nil
}

var (
	_ Object = &ObjectV1_1{}
)
