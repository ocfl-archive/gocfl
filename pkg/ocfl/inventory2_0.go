package ocfl

import (
	"context"
	"net/url"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
)

const (
	ContentDirectory2_0 = "content"
)

type InventoryV2_0 struct {
	*InventoryBase
}

func newInventoryV2_0(
	ctx context.Context,
	object Object,
	folder string,
	logger zLogger.ZLogger,
	errorFactory *archiveerror.Factory,
) (*InventoryV2_0, error) {
	ivUrl, _ := url.Parse(string(InventorySpec2_0))
	ib, err := newInventoryBase(ctx, object, folder, ivUrl, "", logger, errorFactory)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create InventoryBase")
	}

	i := &InventoryV2_0{InventoryBase: ib}
	return i, nil
}

func (i *InventoryV2_0) IsEqual(i2 Inventory) bool {
	i11_2, ok := i2.(*InventoryV2_0)
	if !ok {
		return false
	}
	return i.InventoryBase.isEqual(i11_2.InventoryBase)
}

var (
	_ Inventory = &InventoryV2_0{}
)
