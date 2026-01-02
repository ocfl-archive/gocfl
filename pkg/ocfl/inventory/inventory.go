package inventory

import (
	"context"
	"encoding/json"
	"slices"

	"emperror.dev/errors"
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/interfaces"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

func _NewInventory(ctx context.Context, folder string, ver version.OCFLVersion, logger zLogger.ZLogger) (interfaces.Inventory, error) {
	switch ver {
	case version.Version1_1:
		sr, err := newInventoryV1_1(ctx, ver, folder, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
	default:
		//case Version1_0:
		sr, err := newInventoryV1_0(ctx, ver, folder, logger)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return sr, nil
		//		return nil, errors.Finalize(fmt.Sprintf("Inventory Version %s not supported", version))
	}
}

func InventoryIsEqual(i1, i2 interfaces.Inventory) bool {
	data1, err := json.Marshal(i1)
	if err != nil {
		return false
	}

	data2, err := json.Marshal(i2)
	if err != nil {
		return false
	}
	return slices.Equal(data1, data2)
}
