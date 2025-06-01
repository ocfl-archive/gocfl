package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/checksum"
	"github.com/je4/utils/v2/pkg/zLogger"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
	"io/fs"
)

type VersionV1_1 struct {
	*VersionBase
}

func newVersionV1_1(objectID, version string, ctx context.Context, fsys fs.FS, inventory Inventory, manager ExtensionManager, digestAlgorithm checksum.DigestAlgorithm, logger zLogger.ZLogger, factory *archiveerror.Factory) (*VersionV1_1, error) {
	ob, err := newVersionBase(objectID, version, ctx, fsys, Version1_1, inventory, nil, manager, logger, factory)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obv10 := &VersionV1_1{VersionBase: ob}
	return obv10, nil
}

var (
	_ Version = &VersionV1_1{}
)
