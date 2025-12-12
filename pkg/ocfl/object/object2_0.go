package object

import (
	"context"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/extension"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"

	"io/fs"
)

const ObjectV20Version = "2.0"

type ObjectV2_0 struct {
	*ObjectBase
}

func newObjectV2_0(ctx context.Context, fsys fs.FS, extensionFactory *extension.ExtensionFactory, extensionManager extension.ExtensionManager, logger zLogger.ZLogger) (*ObjectV2_0, error) {
	ob, err := newObjectBase(ctx, fsys, version.Version2_0, extensionFactory, extensionManager, logger)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	obv20 := &ObjectV2_0{ObjectBase: ob}
	return obv20, nil
}

var (
	_ Object = &ObjectV2_0{}
)
