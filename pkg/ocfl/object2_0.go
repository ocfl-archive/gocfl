package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	"io/fs"
)

const ObjectV20Version = "2.0"

type ObjectV2_0 struct {
	*ObjectBase
}

func newObjectV2_0(ctx context.Context, fsys fs.FS, storageRoot StorageRoot, extensionManager ExtensionManager, logger zLogger.ZWrapper) (*ObjectV2_0, error) {
	ob, err := newObjectBase(ctx, fsys, Version2_0, storageRoot, extensionManager, logger)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obv20 := &ObjectV2_0{ObjectBase: ob}
	return obv20, nil
}

var (
	_ Object = &ObjectV2_0{}
)
