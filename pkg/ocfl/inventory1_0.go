package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"net/url"
)

const (
	ContentDirectory1_0 = "content"
)

type InventoryV1_0 struct {
	*InventoryBase
}

func newInventoryV1_0(ctx context.Context, object Object, folder string, logger *logging.Logger) (*InventoryV1_0, error) {
	ivUrl, _ := url.Parse(string(InventorySpec1_0))
	ib, err := newInventoryBase(ctx, object, folder, ivUrl, "", logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create InventoryBase")
	}

	i := &InventoryV1_0{InventoryBase: ib}
	return i, nil
}
