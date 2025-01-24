package ocfl

import (
	"context"
	"net/url"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	archiveerror "github.com/ocfl-archive/error/pkg/error"
)

const (
	ContentDirectory1_1 = "content"
)

type InventoryV1_1 struct {
	*InventoryBase
}

func newInventoryV1_1(
	ctx context.Context,
	object Object,
	folder string,
	logger zLogger.ZLogger,
	errorFactory *archiveerror.Factory,
) (*InventoryV1_1, error) {
	ivUrl, _ := url.Parse(string(InventorySpec1_1))
	ib, err := newInventoryBase(ctx, object, folder, ivUrl, "", logger, errorFactory)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create InventoryBase")
	}

	i := &InventoryV1_1{InventoryBase: ib}
	return i, nil
}

func (i *InventoryV1_1) IsEqual(i2 Inventory) bool {
	i11_2, ok := i2.(*InventoryV1_1)
	if !ok {
		return false
	}
	return i.InventoryBase.isEqual(i11_2.InventoryBase)
}

var (
	_ Inventory = &InventoryV1_1{}
)
