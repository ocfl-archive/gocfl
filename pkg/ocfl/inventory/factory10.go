package inventory

import (
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/types"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

var f10 = NewFactory10(nil)

func NewFactory10(logger zLogger.ZLogger) types.Factory {
	return &factory10{
		Factory: NewFactoryBase(version.Version1_0, types.InventorySpec1_0, logger),
	}
}

type factory10 struct {
	types.Factory
}

var _ types.Factory = (*factory10)(nil)
