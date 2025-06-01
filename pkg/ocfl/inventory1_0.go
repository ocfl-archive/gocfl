package ocfl

import (
	"context"
	"net/url"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"

	archiveerror "github.com/ocfl-archive/error/pkg/error"
)

const (
	ContentDirectory1_0 = "content"
)

type InventoryV1_0 struct {
	*InventoryBase
}

func newInventoryV1_0(
	ctx context.Context,
	folder string,
	ocflVersion OCFLVersion,
	logger zLogger.ZLogger,
	errorFactory *archiveerror.Factory,
) (*InventoryV1_0, error) {
	ivUrl, _ := url.Parse(string(InventorySpec1_0))
	ib, err := newInventoryBase(ctx, folder, ocflVersion, ivUrl, "", logger, errorFactory)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create InventoryBase")
	}

	i := &InventoryV1_0{InventoryBase: ib}
	return i, nil
}

func (i *InventoryV1_0) IsEqual(i2 Inventory) bool {
	i10_2, ok := i2.(*InventoryV1_0)
	if !ok {
		return false
	}

	return i.InventoryBase.isEqual(i10_2.InventoryBase)
}

func (i *InventoryV1_0) Contains(i2 Inventory) bool {
	i10_2, ok := i2.(*InventoryV1_0)
	if !ok {
		return false
	}
	return i.InventoryBase.contains(i10_2.InventoryBase)
}

var (
	_ Inventory = &InventoryV1_0{}
)
