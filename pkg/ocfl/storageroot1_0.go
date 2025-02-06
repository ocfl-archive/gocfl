package ocfl

import (
	"context"
	"io/fs"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
)

const Version1_0 OCFLVersion = "1.0"

type StorageRootV1_0 struct {
	*StorageRootBase
}

func NewStorageRootV1_0(
	ctx context.Context,
	fsys fs.FS,
	extensionFactory *ExtensionFactory,
	manager ExtensionManager,
	logger zLogger.ZLogger,
	errorFactory *archiveerror.Factory,
) (*StorageRootV1_0, error) {
	srb, err := NewStorageRootBase(ctx, fsys, Version1_0, extensionFactory, manager, logger, errorFactory)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create StorageRootBase Version %s", Version1_0)
	}

	sr := &StorageRootV1_0{StorageRootBase: srb}
	return sr, nil
}
