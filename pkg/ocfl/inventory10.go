package ocfl

import (
	"emperror.dev/errors"
	"github.com/op/go-logging"
	"go.ub.unibas.ch/gocfl/v2/pkg/checksum"
	"net/url"
)

const (
	InventoryType10    = "https://ocfl.io/1.0/spec/#inventory"
	DigestAlg10        = checksum.DigestSHA512
	ContentDirectory10 = "content"
)

type InventoryV10 struct {
	*InventoryBase
}

func NewInventoryV10(id string, logger *logging.Logger) (*InventoryV10, error) {
	ivUrl, _ := url.Parse(InventoryType10)
	ib, err := NewInventoryBase(id, ivUrl, DigestAlg10, ContentDirectory10, logger)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create InventoryBase")
	}

	i := &InventoryV10{InventoryBase: ib}
	return i, nil
}
