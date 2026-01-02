package inventory

import (
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/types"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

var f20 = NewFactory20(nil)

func NewFactory20(logger zLogger.ZLogger) types.Factory {
	return &factory20{
		logger:  logger,
		Factory: NewFactoryBase(version.Version1_1, types.InventorySpec1_1, logger),
	}
}

type factory20 struct {
	types.Factory
	logger zLogger.ZLogger
}

var _ types.Factory = (*factory20)(nil)
