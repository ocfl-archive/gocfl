package ocfl

import (
	"context"
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"net/url"
)

const (
	ContentDirectory1_0 = "content"
)

type InventoryV1_0 struct {
	*InventoryBase
}

func NewInventoryV1_0(ctx context.Context, object Object, id string, digest checksum.DigestAlgorithm, logger *logging.Logger) (*InventoryV1_0, error) {
	ivUrl, _ := url.Parse(string(InventorySpec1_0))
	ib, err := NewInventoryBase(ctx, object, id, ivUrl, digest, ContentDirectory1_0, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create InventoryBase")
	}

	i := &InventoryV1_0{InventoryBase: ib}
	return i, nil
}
