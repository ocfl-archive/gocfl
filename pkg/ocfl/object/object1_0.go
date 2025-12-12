package object

import (
	"context"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"

	"io/fs"
)

type ObjectV1_0 struct {
	*ObjectBase
}

func newObjectV1_0(ctx context.Context, fsys fs.FS, extensionFactory *extension.ExtensionFactory, extensionManager extension.ExtensionManager, logger zLogger.ZLogger) (*ObjectV1_0, error) {
	ob, err := newObjectBase(ctx, fsys, version.Version1_0, extensionFactory, extensionManager, logger)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obv10 := &ObjectV1_0{ObjectBase: ob}
	return obv10, nil
}

var (
	_ Object = &ObjectV1_0{}
)
