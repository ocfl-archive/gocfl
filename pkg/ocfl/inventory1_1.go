package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"net/url"
)

const (
	ContentDirectory1_1 = "content"
)

type InventoryV1_1 struct {
	*InventoryBase
}

func newInventoryV1_1(ctx context.Context, object Object, folder string, logger *logging.Logger) (*InventoryV1_1, error) {
	ivUrl, _ := url.Parse(string(InventorySpec1_1))
	ib, err := newInventoryBase(ctx, object, folder, ivUrl, "", logger)
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
