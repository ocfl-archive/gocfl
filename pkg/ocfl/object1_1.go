package ocfl

import (
	"context"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
)

const ObjectV11Version = "1.1"

type ObjectV1_1 struct {
	*ObjectBase
}

func newObjectV1_1(
	ctx context.Context,
	fsys fs.FS,
	storageRoot StorageRoot,
	extensionManager ExtensionManager,
	logger zLogger.ZLogger,
	errorFactory *archiveerror.Factory,
) (*ObjectV1_1, error) {
	ob, err := newObjectBase(ctx, fsys, Version1_1, storageRoot, extensionManager, logger, errorFactory)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obv11 := &ObjectV1_1{ObjectBase: ob}
	return obv11, nil
}

var (
	_ Object = &ObjectV1_1{}
)
