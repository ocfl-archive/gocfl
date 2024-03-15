package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"io/fs"
)

type ObjectV1_0 struct {
	*ObjectBase
}

func newObjectV1_0(ctx context.Context, fsys fs.FS, storageRoot StorageRoot, logger zLogger.ZWrapper) (*ObjectV1_0, error) {
	ob, err := newObjectBase(ctx, fsys, Version1_0, storageRoot, logger)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obv10 := &ObjectV1_0{ObjectBase: ob}
	return obv10, nil
}

var (
	_ Object = &ObjectV1_0{}
)
