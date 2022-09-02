package ocfl

import (
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"net/url"
)

const (
	InventoryType1_0    = "https://ocfl.io/1.0/spec/#inventory"
	DigestAlg1_0        = checksum.DigestSHA512
	ContentDirectory1_0 = "content"
)

type InventoryV1_0 struct {
	*InventoryBase
}

func NewInventoryV1_0(id string, logger *logging.Logger) (*InventoryV1_0, error) {
	ivUrl, _ := url.Parse(InventoryType1_0)
	ib, err := NewInventoryBase(id, ivUrl, DigestAlg1_0, ContentDirectory1_0, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create InventoryBase")
	}

	i := &InventoryV1_0{InventoryBase: ib}
	return i, nil
}
