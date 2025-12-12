package inventory

import (
	"context"
	"net/url"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

const (
	ContentDirectory1_0 = "content"
)

type InventoryV1_0 struct {
	*InventoryBase
}

func newInventoryV1_0(ctx context.Context, ver version.OCFLVersion, folder string, logger zLogger.ZLogger) (*InventoryV1_0, error) {
	ivUrl, _ := url.Parse(string(InventorySpec1_0))
	ib, err := newInventoryBase(ctx, ver, folder, ivUrl, "", logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create InventoryBase")
	}

	i := &InventoryV1_0{InventoryBase: ib}
	return i, nil
}

func (i *InventoryV1_0) isEqual(i2 Inventory) bool {
	i10_2, ok := i2.(*InventoryV1_0)
	if !ok {
		return false
	}

	return i.InventoryBase.IsEqual(i10_2.InventoryBase)
}

var (
	_ Inventory = &InventoryV1_0{}
)
