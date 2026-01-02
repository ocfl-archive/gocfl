package inventory

import (
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/interfaces"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

var f10 = NewFactory10(nil)

func NewFactory10(logger zLogger.ZLogger) interfaces.Factory {
	return &factory10{
		Factory: NewFactoryBase(version.Version1_0, InventorySpec1_0, logger),
	}
}

type factory10 struct {
	interfaces.Factory
}

var _ interfaces.Factory = (*factory10)(nil)
