package inventory

import (
	"github.com/je4/utils/v2/pkg/zLogger"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/interfaces"
	"github.com/ocfl-archive/gocfl/v2/pkg/ocfl/version"
)

var f11 = NewFactory11(nil)

func NewFactory11(logger zLogger.ZLogger) interfaces.Factory {
	return &factory11{
		logger:  logger,
		Factory: NewFactoryBase(version.Version1_1, InventorySpec1_1, logger),
	}
}

type factory11 struct {
	interfaces.Factory
	logger zLogger.ZLogger
}

var _ interfaces.Factory = (*factory11)(nil)
