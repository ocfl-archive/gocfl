package ocfl

import (
	"context"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
)

type ObjectV1_0 struct {
	*ObjectBase
}

func newObjectV1_0(
	ctx context.Context,
	fsys fs.FS,
	storageRoot StorageRoot,
	extensionManager ExtensionManager,
	logger zLogger.ZLogger,
	errorFactory *archiveerror.Factory,
) (*ObjectV1_0, error) {
	ob, err := newObjectBase(ctx, fsys, Version1_0, storageRoot, extensionManager, VersionPlain, logger, errorFactory)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obv10 := &ObjectV1_0{ObjectBase: ob}
	return obv10, nil
}

var (
	_ Object = &ObjectV1_0{}
)
