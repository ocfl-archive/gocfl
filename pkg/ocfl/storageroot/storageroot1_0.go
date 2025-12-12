package storageroot

import (
	"context"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"

	"io/fs"
)

//const Version1_0 version.OCFLVersion = "1.0"

type StorageRootV1_0 struct {
	*StorageRootBase
}

func NewStorageRootV1_0(ctx context.Context, fsys fs.FS, extensionFactory *extension.ExtensionFactory, manager ExtensionManager, logger zLogger.ZLogger) (*StorageRootV1_0, error) {
	srb, err := NewStorageRootBase(ctx, fsys, version.Version1_0, extensionFactory, manager, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create StorageRootBase Version %s", version.Version1_0)
	}

	sr := &StorageRootV1_0{StorageRootBase: srb}
	return sr, nil
}
